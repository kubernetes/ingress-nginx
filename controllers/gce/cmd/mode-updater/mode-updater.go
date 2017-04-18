package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

var (
	projectID           string
	regionName          string
	targetBalancingMode string

	instanceGroupName string

	s         *compute.Service
	zones     []*compute.Zone
	igs       map[string]*compute.InstanceGroup
	instances []*compute.Instance
)

const (
	instanceGroupTemp = "k8s-ig--migrate"
	balancingModeRATE = "RATE"
	balancingModeUTIL = "UTILIZATION"

	operationPollInterval = 3 * time.Second
	// Creating Route in very large clusters, may take more than half an hour.
	operationPollTimeoutDuration = time.Hour
	version                      = 0.1
)

func main() {
	fmt.Println("Backend-Service BalancingMode Updater", version)
	//flag.Usage
	flag.Parse()

	args := flag.Args()
	if len(args) != 3 {
		log.Fatalf("Expected three arguments: project_id region balancing_mode")
	}
	projectID, regionName, targetBalancingMode = args[0], args[1], args[2]

	switch targetBalancingMode {
	case balancingModeRATE, balancingModeUTIL:
	default:
		panic(fmt.Errorf("expected either %s or %s, actual: %v", balancingModeRATE, balancingModeUTIL, targetBalancingMode))
	}

	igs = make(map[string]*compute.InstanceGroup)

	tokenSource, err := google.DefaultTokenSource(
		oauth2.NoContext,
		compute.CloudPlatformScope,
		compute.ComputeScope)
	if err != nil {
		panic(err)
	}

	client := oauth2.NewClient(oauth2.NoContext, tokenSource)
	s, err = compute.New(client)
	if err != nil {
		panic(err)
	}

	// Get Zones
	zoneFilter := fmt.Sprintf("(region eq %s)", createRegionLink(regionName))
	zoneList, err := s.Zones.List(projectID).Filter(zoneFilter).Do()
	if err != nil {
		panic(err)
	}
	zones = zoneList.Items

	// Get instance groups
	for _, z := range zones {
		igl, err := s.InstanceGroups.List(projectID, z.Name).Do()
		if err != nil {
			panic(err)
		}
		for _, ig := range igl.Items {
			if !strings.HasPrefix(ig.Name, "k8s-ig--") {
				continue
			}

			if instanceGroupName == "" {
				instanceGroupName = ig.Name
			}

			// Note instances
			r := &compute.InstanceGroupsListInstancesRequest{InstanceState: "ALL"}
			instList, err := s.InstanceGroups.ListInstances(projectID, getResourceName(ig.Zone, "zones"), ig.Name, r).Do()
			if err != nil {
				panic(err)
			}

			for _, i := range instList.Items {
				inst, err := s.Instances.Get(projectID, getResourceName(ig.Zone, "zones"), getResourceName(i.Instance, "instances")).Do()
				if err != nil {
					panic(err)
				}

				instances = append(instances, inst)
			}

			// Note instance group in zone
			igs[z.Name] = ig
		}
	}

	if instanceGroupName == "" {
		panic(errors.New("Could not determine k8s load balancer instance group"))
	}

	bs := getBackendServices()
	fmt.Println("Backend Services:", len(bs))
	fmt.Println("Instance Groups:", len(igs))

	// Early return for special cases
	switch len(bs) {
	case 0:
		fmt.Println("There are 0 backend services - no action necessary")
		return
	case 1:
		updateSingleBackend(bs[0])
		return
	}

	// Check there's work to be done
	if typeOfBackends(bs) == targetBalancingMode {
		fmt.Println("Backends are already set to target mode")
		return
	}

	// Check no orphan instance groups will throw us off
	clusters := getIGClusterIds()
	if len(clusters) != 1 {
		fmt.Println("Expecting only cluster of instance groups in GCE, found", clusters)
		return
	}

	// Performing update for 2+ backend services
	updateMultipleBackends()
}

func updateMultipleBackends() {
	// Create temoprary instance groups
	for zone, ig := range igs {
		_, err := s.InstanceGroups.Get(projectID, zone, instanceGroupTemp).Do()
		if err != nil {
			newIg := &compute.InstanceGroup{
				Name:       instanceGroupTemp,
				Zone:       zone,
				NamedPorts: ig.NamedPorts,
			}
			fmt.Println("Creating", instanceGroupTemp, "zone:", zone)
			op, err := s.InstanceGroups.Insert(projectID, zone, newIg).Do()
			if err != nil {
				panic(err)
			}

			if err = waitForZoneOp(op, zone); err != nil {
				panic(err)
			}
		}
	}

	// Straddle both groups
	fmt.Println("Straddle both groups in backend services")
	setBackendsTo(true, balancingModeInverse(targetBalancingMode), true, balancingModeInverse(targetBalancingMode))

	fmt.Println("Migrate instances to temporary group")
	migrateInstances(instanceGroupName, instanceGroupTemp)

	// Remove original backends to get rid of old balancing mode
	fmt.Println("Remove original backends")
	setBackendsTo(false, "", true, balancingModeInverse(targetBalancingMode))

	// Straddle both groups (creates backend services to original groups with target mode)
	fmt.Println("Create backends pointing to original instance groups")
	setBackendsTo(true, targetBalancingMode, true, balancingModeInverse(targetBalancingMode))

	fmt.Println("Migrate instances back to original groups")
	migrateInstances(instanceGroupTemp, instanceGroupName)

	fmt.Println("Remove temporary backends")
	setBackendsTo(true, targetBalancingMode, false, "")

	fmt.Println("Delete temporary instance groups")
	for z := range igs {
		op, err := s.InstanceGroups.Delete(projectID, z, instanceGroupTemp).Do()
		if err != nil {
			fmt.Println("Couldn't delete temporary instance group", instanceGroupTemp)
		}

		if err = waitForZoneOp(op, z); err != nil {
			fmt.Println("Couldn't wait for operation: deleting temporary instance group", instanceGroupName)
		}
	}
}

func sleep(d time.Duration) {
	fmt.Println("Sleeping for", d.String())
	time.Sleep(d)
}

func setBackendsTo(orig bool, origMode string, temp bool, tempMode string) {
	bs := getBackendServices()
	for _, bsi := range bs {
		var union []*compute.Backend
		for zone := range igs {
			if orig {
				b := &compute.Backend{
					Group:              createInstanceGroupLink(zone, instanceGroupName),
					BalancingMode:      origMode,
					CapacityScaler:     0.8,
					MaxRatePerInstance: 1.0,
				}
				union = append(union, b)
			}
			if temp {
				b := &compute.Backend{
					Group:              createInstanceGroupLink(zone, instanceGroupTemp),
					BalancingMode:      tempMode,
					CapacityScaler:     0.8,
					MaxRatePerInstance: 1.0,
				}
				union = append(union, b)
			}
		}
		bsi.Backends = union
		op, err := s.BackendServices.Update(projectID, bsi.Name, bsi).Do()
		if err != nil {
			panic(err)
		}

		if err = waitForGlobalOp(op); err != nil {
			panic(err)
		}
	}
}

func balancingModeInverse(m string) string {
	switch m {
	case balancingModeRATE:
		return balancingModeUTIL
	case balancingModeUTIL:
		return balancingModeRATE
	default:
		return ""
	}
}

func getBackendServices() (bs []*compute.BackendService) {
	bsl, err := s.BackendServices.List(projectID).Do()
	if err != nil {
		panic(err)
	}

	for _, bsli := range bsl.Items {
		// Ignore regional backend-services and only grab Kubernetes resources
		if bsli.Region == "" && strings.HasPrefix(bsli.Name, "k8s-be-") {
			bs = append(bs, bsli)
		}
	}
	return bs
}

func typeOfBackends(bs []*compute.BackendService) string {
	if len(bs) == 0 {
		return ""
	}
	return bs[0].Backends[0].BalancingMode
}

func migrateInstances(fromIG, toIG string) error {
	wg := sync.WaitGroup{}
	for _, i := range instances {
		wg.Add(1)
		go func(i *compute.Instance) {
			z := getResourceName(i.Zone, "zones")
			fmt.Printf(" - %s (%s)\n", i.Name, z)
			rr := &compute.InstanceGroupsRemoveInstancesRequest{Instances: []*compute.InstanceReference{{Instance: i.SelfLink}}}
			op, err := s.InstanceGroups.RemoveInstances(projectID, z, fromIG, rr).Do()
			if err != nil {
				fmt.Println("Skipping error when removing instance from group", err)
			}

			if err = waitForZoneOp(op, z); err != nil {
				fmt.Println("Failed to wait for operation: removing instance from group", err)
			}

			ra := &compute.InstanceGroupsAddInstancesRequest{Instances: []*compute.InstanceReference{{Instance: i.SelfLink}}}
			op, err = s.InstanceGroups.AddInstances(projectID, z, toIG, ra).Do()
			if err != nil {
				if !strings.Contains(err.Error(), "memberAlreadyExists") { // GLBC already added the instance back to the IG
					fmt.Println("failed to add instance to new IG", i.Name, err)
				}
			}

			if err = waitForZoneOp(op, z); err != nil {
				fmt.Println("Failed to wait for operation: adding instance to group", err)
			}
			wg.Done()
		}(i)
		time.Sleep(10 * time.Second)
	}
	wg.Wait()
	return nil
}

func createInstanceGroupLink(zone, igName string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instanceGroups/%s", projectID, zone, igName)
}

func createRegionLink(region string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/nicksardo-playground/regions/%v", region)
}

func getResourceName(link string, resourceType string) string {
	s := strings.Split(link, "/")

	for i := 0; i < len(s); i++ {
		if s[i] == resourceType {
			if i+1 <= len(s) {
				return s[i+1]
			}
		}
	}
	return ""
}

func updateSingleBackend(bs *compute.BackendService) {
	needsUpdate := false
	for _, b := range bs.Backends {
		if b.BalancingMode != targetBalancingMode {
			needsUpdate = true
			b.BalancingMode = targetBalancingMode
		}
	}

	if !needsUpdate {
		fmt.Println("Single backend had all targetBalancingMode - no change necessary")
		return
	}

	op, err := s.BackendServices.Update(projectID, bs.Name, bs).Do()
	if err != nil {
		panic(err)
	}

	if err = waitForGlobalOp(op); err != nil {
		panic(err)
	}

	fmt.Println("Updated single backend service to target balancing mode.")
}

func getIGClusterIds() []string {
	clusterIds := make(map[string]struct{})
	for _, ig := range igs {
		s := strings.Split(ig.Name, "--")
		if len(s) > 2 {
			panic(fmt.Errorf("Expected two parts to instance group name, got %v", s))
		}
		clusterIds[s[1]] = struct{}{}
	}
	var ids []string
	for v, _ := range clusterIds {
		ids = append(ids, v)
	}
	return ids
}

func waitForOp(op *compute.Operation, getOperation func(operationName string) (*compute.Operation, error)) error {
	if op == nil {
		return fmt.Errorf("operation must not be nil")
	}

	if opIsDone(op) {
		return getErrorFromOp(op)
	}

	opStart := time.Now()
	opName := op.Name
	return wait.Poll(operationPollInterval, operationPollTimeoutDuration, func() (bool, error) {
		start := time.Now()
		duration := time.Now().Sub(start)
		if duration > 5*time.Second {
			glog.Infof("pollOperation: throttled %v for %v", duration, opName)
		}
		pollOp, err := getOperation(opName)
		if err != nil {
			glog.Warningf("GCE poll operation %s failed: pollOp: [%v] err: [%v] getErrorFromOp: [%v]",
				opName, pollOp, err, getErrorFromOp(pollOp))
		}
		done := opIsDone(pollOp)
		if done {
			duration := time.Now().Sub(opStart)
			if duration > 1*time.Minute {
				// Log the JSON. It's cleaner than the %v structure.
				enc, err := pollOp.MarshalJSON()
				if err != nil {
					glog.Warningf("waitForOperation: long operation (%v): %v (failed to encode to JSON: %v)",
						duration, pollOp, err)
				} else {
					glog.Infof("waitForOperation: long operation (%v): %v",
						duration, string(enc))
				}
			}
		}
		return done, getErrorFromOp(pollOp)
	})
}

func opIsDone(op *compute.Operation) bool {
	return op != nil && op.Status == "DONE"
}

func getErrorFromOp(op *compute.Operation) error {
	if op != nil && op.Error != nil && len(op.Error.Errors) > 0 {
		err := &googleapi.Error{
			Code:    int(op.HttpErrorStatusCode),
			Message: op.Error.Errors[0].Message,
		}
		glog.Errorf("GCE operation failed: %v", err)
		return err
	}

	return nil
}

func waitForGlobalOp(op *compute.Operation) error {
	return waitForOp(op, func(operationName string) (*compute.Operation, error) {
		return s.GlobalOperations.Get(projectID, operationName).Do()
	})
}

func waitForRegionOp(op *compute.Operation, region string) error {
	return waitForOp(op, func(operationName string) (*compute.Operation, error) {
		return s.RegionOperations.Get(projectID, region, operationName).Do()
	})
}

func waitForZoneOp(op *compute.Operation, zone string) error {
	return waitForOp(op, func(operationName string) (*compute.Operation, error) {
		return s.ZoneOperations.Get(projectID, zone, operationName).Do()
	})
}
