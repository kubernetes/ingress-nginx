package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
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
	clusterID           string
	regionName          string
	targetBalancingMode string

	instanceGroupName string

	s         *compute.Service
	zones     []*compute.Zone
	igs       map[string]*compute.InstanceGroup
	instances map[string][]string
)

const (
	instanceGroupTemp = "k8s-ig--migrate"
	balancingModeRATE = "RATE"
	balancingModeUTIL = "UTILIZATION"

	loadBalancerUpdateTime = 3 * time.Minute

	operationPollInterval        = 1 * time.Second
	operationPollTimeoutDuration = time.Hour

	version = 0.1
)

func main() {
	fmt.Println("Backend-Service BalancingMode Updater", version)
	flag.Parse()

	args := flag.Args()
	if len(args) != 4 {
		log.Fatalf("Expected four arguments: project_id cluster_id region balancing_mode")
	}
	projectID, clusterID, regionName, targetBalancingMode = args[0], args[1], args[2], args[3]

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

	if len(zones) == 0 {
		panic(fmt.Errorf("Expected at least one zone in region: %v", regionName))
	}

	instanceGroupName = fmt.Sprintf("k8s-ig--%s", clusterID)
	instances = make(map[string][]string)

	// Get instance groups
	for _, z := range zones {
		igl, err := s.InstanceGroups.List(projectID, z.Name).Do()
		if err != nil {
			panic(err)
		}
		for _, ig := range igl.Items {
			if instanceGroupName != ig.Name {
				continue
			}

			// Note instances
			r := &compute.InstanceGroupsListInstancesRequest{InstanceState: "ALL"}
			instList, err := s.InstanceGroups.ListInstances(projectID, getResourceName(ig.Zone, "zones"), ig.Name, r).Do()
			if err != nil {
				panic(err)
			}

			var instanceLinks []string
			for _, i := range instList.Items {
				instanceLinks = append(instanceLinks, i.Instance)
			}

			// Note instance group in zone
			igs[z.Name] = ig
			instances[z.Name] = instanceLinks
		}
	}

	if len(igs) == 0 {
		panic(fmt.Errorf("Expected at least one instance group named: %v", instanceGroupName))
	}

	bs := getBackendServices()
	fmt.Println("Backend Services:")
	for _, b := range bs {
		fmt.Println(" - ", b.Name)
	}
	fmt.Println("Instance Groups:")
	for z, g := range igs {
		fmt.Printf(" - %v (%v)\n", g.Name, z)
	}

	// Early return for special cases
	switch len(bs) {
	case 0:
		fmt.Println("\nThere are 0 backend services - no action necessary")
		return
	case 1:
		updateSingleBackend(bs[0])
		return
	}

	// Check there's work to be done
	if typeOfBackends(bs) == targetBalancingMode {
		fmt.Println("\nBackends are already set to target mode")
		return
	}

	// Performing update for 2+ backend services
	updateMultipleBackends()
}

func updateMultipleBackends() {
	fmt.Println("\nStep 1: Creating temporary instance groups in relevant zones")
	// Create temoprary instance groups
	for zone, ig := range igs {
		_, err := s.InstanceGroups.Get(projectID, zone, instanceGroupTemp).Do()
		if err != nil {
			newIg := &compute.InstanceGroup{
				Name:       instanceGroupTemp,
				Zone:       zone,
				NamedPorts: ig.NamedPorts,
			}
			fmt.Printf(" - %v (%v)\n", instanceGroupTemp, zone)
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
	fmt.Println("\nStep 2: Update backend services to point to original and temporary instance groups")
	setBackendsTo(true, balancingModeInverse(targetBalancingMode), true, balancingModeInverse(targetBalancingMode))

	sleep(loadBalancerUpdateTime)

	fmt.Println("\nStep 3: Migrate instances to temporary group")
	migrateInstances(instanceGroupName, instanceGroupTemp)

	// Remove original backends to get rid of old balancing mode
	fmt.Println("\nStep 4: Update backend services to point only to temporary instance groups")
	setBackendsTo(false, "", true, balancingModeInverse(targetBalancingMode))

	// Straddle both groups (creates backend services to original groups with target mode)
	fmt.Println("\nStep 5: Update backend services to point to both temporary and original (with new balancing mode) instance groups")
	setBackendsTo(true, targetBalancingMode, true, balancingModeInverse(targetBalancingMode))

	sleep(loadBalancerUpdateTime)

	fmt.Println("\nStep 6: Migrate instances back to original groups")
	migrateInstances(instanceGroupTemp, instanceGroupName)

	sleep(loadBalancerUpdateTime)

	fmt.Println("\nStep 7: Update backend services to point only to original instance groups")
	setBackendsTo(true, targetBalancingMode, false, "")

	fmt.Println("\nStep 8: Delete temporary instance groups")
	for z := range igs {
		fmt.Printf(" - %v (%v)\n", instanceGroupTemp, z)
		op, err := s.InstanceGroups.Delete(projectID, z, instanceGroupTemp).Do()
		if err != nil {
			fmt.Println("Couldn't delete temporary instance group", instanceGroupTemp)
		}

		if err = waitForZoneOp(op, z); err != nil {
			fmt.Println("Couldn't wait for operation: deleting temporary instance group", instanceGroupName)
		}
	}
}

func sleep(t time.Duration) {
	fmt.Println("\nSleeping for", t)
	time.Sleep(t)
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
		fmt.Printf(" - %v\n", bsi.Name)
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
		if bsli.Region == "" && strings.HasPrefix(bsli.Name, "k8s-be-") && strings.HasSuffix(bsli.Name, clusterID) {
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

func migrateInstances(fromIG, toIG string) {
	for z, links := range instances {
		for _, i := range links {
			name := getResourceName(i, "instances")
			fmt.Printf(" - %s (%s): ", name, z)
			if err := migrateInstance(z, i, fromIG, toIG); err != nil {
				fmt.Printf(" err: %v", err)
			} else {
				fmt.Println()
			}
		}
	}
}

func migrateInstance(zone, instanceLink, fromIG, toIG string) error {
	attempts := 0
	return wait.Poll(3*time.Second, 10*time.Minute, func() (bool, error) {
		attempts++
		if attempts > 1 {
			fmt.Printf(" (attempt %v) ", attempts)
		}
		// Remove from old group
		rr := &compute.InstanceGroupsRemoveInstancesRequest{Instances: []*compute.InstanceReference{{Instance: instanceLink}}}
		op, err := s.InstanceGroups.RemoveInstances(projectID, zone, fromIG, rr).Do()
		if err != nil {
			fmt.Printf("failed to remove from group %v, err: %v,", fromIG, err)
		} else if err = waitForZoneOp(op, zone); err != nil {
			fmt.Printf("failed to wait for operation: removing instance from %v, err: %v,", fromIG, err)
		} else {
			fmt.Printf("removed from %v,", fromIG)
		}

		// Add to new group
		ra := &compute.InstanceGroupsAddInstancesRequest{Instances: []*compute.InstanceReference{{Instance: instanceLink}}}
		op, err = s.InstanceGroups.AddInstances(projectID, zone, toIG, ra).Do()
		if err != nil {
			if strings.Contains(err.Error(), "memberAlreadyExists") { // GLBC already added the instance back to the IG
				fmt.Printf(" already exists in %v", toIG)
			} else {
				fmt.Printf(" failed to add to group %v, err: %v\n", toIG, err)
				return false, nil
			}
		} else if err = waitForZoneOp(op, zone); err != nil {
			fmt.Printf(" failed to wait for operation: adding instance to %v, err: %v", toIG, err)
		} else {
			fmt.Printf(" added to %v", toIG)
		}

		return true, nil
	})
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

// Below operations are copied from the GCE CloudProvider and modified to be static

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
