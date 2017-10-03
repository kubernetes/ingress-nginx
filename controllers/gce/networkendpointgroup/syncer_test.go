package networkendpointgroup

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
	"k8s.io/ingress/controllers/gce/utils"
	"reflect"
	"testing"
	"time"
)

const (
	NegName          = "test-neg-name"
	ServiceNamespace = "test-ns"
	ServiceName      = "test-name"
	NamedPort        = "named-port"
)

func NewTestSyncer() *syncer {
	kubeClient := fake.NewSimpleClientset()
	context := utils.NewControllerContext(kubeClient, apiv1.NamespaceAll, 1*time.Second, true)
	svcPort := servicePort{
		namespace:  ServiceNamespace,
		name:       ServiceName,
		targetPort: "80",
	}

	return newSyncer(svcPort,
		NegName,
		record.NewFakeRecorder(100),
		NewFakeNetworkEndpointGroupCloud("test-subnetwork", "test-newtork"),
		NewFakeZoneGetter(),
		context.ServiceInformer.GetIndexer(),
		context.EndpointInformer.GetIndexer())
}

func TestStartAndStopSyncer(t *testing.T) {
	syncer := NewTestSyncer()
	if !syncer.IsStopped() {
		t.Fatalf("Syncer is not stopped after creation.")
	}
	if syncer.IsShuttingDown() {
		t.Fatalf("Syncer is shutting down after creation.")
	}

	if err := syncer.Start(); err != nil {
		t.Fatalf("Failed to start syncer: %v", err)
	}
	if syncer.IsStopped() {
		t.Fatalf("Syncer is stopped after Start.")
	}
	if syncer.IsShuttingDown() {
		t.Fatalf("Syncer is shutting down after Start.")
	}

	syncer.Stop()
	if !syncer.IsStopped() {
		t.Fatalf("Syncer is not stopped after Stop.")
	}

	if err := wait.PollImmediate(time.Second, 30*time.Second, func() (bool, error) {
		return !syncer.IsShuttingDown() && syncer.IsStopped(), nil
	}); err != nil {
		t.Fatalf("Syncer failed to shutdown: %v", err)
	}

	if err := syncer.Start(); err != nil {
		t.Fatalf("Failed to restart syncer: %v", err)
	}
	if syncer.IsStopped() {
		t.Fatalf("Syncer is stopped after restart.")
	}
	if syncer.IsShuttingDown() {
		t.Fatalf("Syncer is shutting down after restart.")
	}

	syncer.Stop()
	if !syncer.IsStopped() {
		t.Fatalf("Syncer is not stopped after Stop.")
	}
}

func TestEnsureNetworkEndpointGroups(t *testing.T) {
	syncer := NewTestSyncer()
	if err := syncer.ensureNetworkEndpointGroups(); err != nil {
		t.Errorf("Failed to ensure NEGs: %v", err)
	}

	ret, _ := syncer.cloud.AggregatedListNetworkEndpointGroup()
	expectZones := []string{TestZone1, TestZone2}
	for _, zone := range expectZones {
		negs, ok := ret[zone]
		if !ok {
			t.Errorf("Failed to find zone %q from ret %v", zone, ret)
			continue
		}

		if len(negs) != 1 {
			t.Errorf("Unexpected negs %v", negs)
		} else {
			if negs[0].Name != NegName {
				t.Errorf("Unexpected neg %q", negs[0].Name)
			}
		}
	}
}

func TestToZoneNetworkEndpointMap(t *testing.T) {
	syncer := NewTestSyncer()
	testCases := []struct {
		targetPort string
		expect     map[string]sets.String
	}{
		{
			targetPort: "80",
			expect: map[string]sets.String{
				TestZone1: sets.NewString("10.100.1.1||instance1||80", "10.100.1.2||instance1||80", "10.100.2.1||instance2||80"),
				TestZone2: sets.NewString("10.100.3.1||instance3||80"),
			},
		},
		{
			targetPort: NamedPort,
			expect: map[string]sets.String{
				TestZone1: sets.NewString("10.100.2.2||instance2||81"),
				TestZone2: sets.NewString("10.100.4.1||instance4||81", "10.100.3.2||instance3||8081", "10.100.4.2||instance4||8081"),
			},
		},
	}

	for _, tc := range testCases {
		syncer.targetPort = tc.targetPort
		res, _ := syncer.toZoneNetworkEndpointMap(getDefaultEndpoint())

		if !reflect.DeepEqual(res, tc.expect) {
			t.Errorf("Expect %v, but got %v.", tc.expect, res)
		}
	}
}

func TestEncodeDecodeEndpoint(t *testing.T) {
	ip := "10.0.0.10"
	instance := "somehost"
	port := "8080"

	retIp, retInstance, retPort := decodeEndpoint(encodeEndpoint(ip, instance, port))

	if ip != retIp || instance != retInstance || retPort != port {
		t.Fatalf("Encode and decode endpoint failed. Expect %q, %q, %q but got %q, %q, %q.", ip, instance, port, retIp, retInstance, retPort)
	}
}

func TestCalculateDifference(t *testing.T) {
	testCases := []struct {
		targetSet  map[string]sets.String
		currentSet map[string]sets.String
		addSet     map[string]sets.String
		removeSet  map[string]sets.String
	}{
		// unchanged
		{
			targetSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
			},
			addSet:    map[string]sets.String{},
			removeSet: map[string]sets.String{},
		},
		// unchanged
		{
			targetSet:  map[string]sets.String{},
			currentSet: map[string]sets.String{},
			addSet:     map[string]sets.String{},
			removeSet:  map[string]sets.String{},
		},
		// add in one zone
		{
			targetSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{},
			addSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
			},
			removeSet: map[string]sets.String{},
		},
		// add in 2 zones
		{
			targetSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
				TestZone2: sets.NewString("e", "f", "g"),
			},
			currentSet: map[string]sets.String{},
			addSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
				TestZone2: sets.NewString("e", "f", "g"),
			},
			removeSet: map[string]sets.String{},
		},
		// remove in one zone
		{
			targetSet: map[string]sets.String{},
			currentSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
			},
			addSet: map[string]sets.String{},
			removeSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
			},
		},
		// remove in 2 zones
		{
			targetSet: map[string]sets.String{},
			currentSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
				TestZone2: sets.NewString("e", "f", "g"),
			},
			addSet: map[string]sets.String{},
			removeSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
				TestZone2: sets.NewString("e", "f", "g"),
			},
		},
		// add and delete in one zone
		{
			targetSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				TestZone1: sets.NewString("b", "c", "d"),
			},
			addSet: map[string]sets.String{
				TestZone1: sets.NewString("a"),
			},
			removeSet: map[string]sets.String{
				TestZone1: sets.NewString("d"),
			},
		},
		// add and delete in 2 zones
		{
			targetSet: map[string]sets.String{
				TestZone1: sets.NewString("a", "b", "c"),
				TestZone2: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				TestZone1: sets.NewString("b", "c", "d"),
				TestZone2: sets.NewString("b", "c", "d"),
			},
			addSet: map[string]sets.String{
				TestZone1: sets.NewString("a"),
				TestZone2: sets.NewString("a"),
			},
			removeSet: map[string]sets.String{
				TestZone1: sets.NewString("d"),
				TestZone2: sets.NewString("d"),
			},
		},
	}

	for _, tc := range testCases {
		addSet, removeSet := calculateDifference(tc.targetSet, tc.currentSet)

		if !reflect.DeepEqual(addSet, tc.addSet) {
			t.Errorf("Failed to calculate difference for add, expecting %v, but got %v", tc.addSet, addSet)
		}

		if !reflect.DeepEqual(removeSet, tc.removeSet) {
			t.Errorf("Failed to calculate difference for remove, expecting %v, but got %v", tc.removeSet, removeSet)
		}
	}
}

func TestSyncNetworkEndpoints(t *testing.T) {
	syncer := NewTestSyncer()
	if err := syncer.ensureNetworkEndpointGroups(); err != nil {
		t.Fatalf("Failed to ensure NEG: %v", err)
	}

	testCases := []struct {
		expectSet map[string]sets.String
		addSet    map[string]sets.String
		removeSet map[string]sets.String
	}{
		{
			expectSet: map[string]sets.String{
				TestZone1: sets.NewString("10.100.1.1||instance1||80", "10.100.2.1||instance2||80"),
				TestZone2: sets.NewString("10.100.3.1||instance3||80", "10.100.4.1||instance4||80"),
			},
			addSet: map[string]sets.String{
				TestZone1: sets.NewString("10.100.1.1||instance1||80", "10.100.2.1||instance2||80"),
				TestZone2: sets.NewString("10.100.3.1||instance3||80", "10.100.4.1||instance4||80"),
			},
			removeSet: map[string]sets.String{},
		},
		{
			expectSet: map[string]sets.String{
				TestZone1: sets.NewString("10.100.1.2||instance1||80"),
				TestZone2: sets.NewString(),
			},
			addSet: map[string]sets.String{
				TestZone1: sets.NewString("10.100.1.2||instance1||80"),
			},
			removeSet: map[string]sets.String{
				TestZone1: sets.NewString("10.100.1.1||instance1||80", "10.100.2.1||instance2||80"),
				TestZone2: sets.NewString("10.100.3.1||instance3||80", "10.100.4.1||instance4||80"),
			},
		},
		{
			expectSet: map[string]sets.String{
				TestZone1: sets.NewString("10.100.1.2||instance1||80"),
				TestZone2: sets.NewString("10.100.3.2||instance3||80"),
			},
			addSet: map[string]sets.String{
				TestZone2: sets.NewString("10.100.3.2||instance3||80"),
			},
			removeSet: map[string]sets.String{},
		},
	}

	for _, tc := range testCases {
		if err := syncer.syncNetworkEndpoints(tc.addSet, tc.removeSet); err != nil {
			t.Fatalf("Failed to sync network endpoints: %v", err)
		}
		examineNetworkEndpoints(tc.expectSet, syncer, t)
	}
}

func examineNetworkEndpoints(expectSet map[string]sets.String, syncer *syncer, t *testing.T) {
	for zone, endpoints := range expectSet {
		expectEndpoints, err := syncer.toNetworkEndpointBatch(endpoints)
		if err != nil {
			t.Fatalf("Failed to convert endpoints to network endpoints: %v", err)
		}
		if cloudEndpoints, err := syncer.cloud.ListNetworkEndpoints(syncer.negName, zone, false); err == nil {
			if len(expectEndpoints) != len(cloudEndpoints) {
				t.Errorf("Expect number of endpoints to be %v, but got %v.", len(expectEndpoints), len(cloudEndpoints))
			}
			for _, expectEp := range expectEndpoints {
				found := false
				for _, cloudEp := range cloudEndpoints {
					if reflect.DeepEqual(*expectEp, *cloudEp.NetworkEndpoint) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Endpoint %v not found.", expectEp)
				}
			}
		} else {
			t.Errorf("Failed to list network endpoints in zone %q: %v.", zone, err)
		}
	}
}

func getDefaultEndpoint() *apiv1.Endpoints {
	instance1 := TestInstance1
	instance2 := TestInstance2
	instance3 := TestInstance3
	instance4 := TestInstance4
	return &apiv1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: ServiceNamespace,
		},
		Subsets: []apiv1.EndpointSubset{
			{
				Addresses: []apiv1.EndpointAddress{
					{
						IP:       "10.100.1.1",
						NodeName: &instance1,
					},
					{
						IP:       "10.100.1.2",
						NodeName: &instance1,
					},
					{
						IP:       "10.100.2.1",
						NodeName: &instance2,
					},
					{
						IP:       "10.100.3.1",
						NodeName: &instance3,
					},
				},
				Ports: []apiv1.EndpointPort{
					{
						Name:     "",
						Port:     int32(80),
						Protocol: apiv1.ProtocolTCP,
					},
				},
			},
			{
				Addresses: []apiv1.EndpointAddress{
					{
						IP:       "10.100.2.2",
						NodeName: &instance2,
					},
					{
						IP:       "10.100.4.1",
						NodeName: &instance4,
					},
				},
				Ports: []apiv1.EndpointPort{
					{
						Name:     NamedPort,
						Port:     int32(81),
						Protocol: apiv1.ProtocolTCP,
					},
				},
			},
			{
				Addresses: []apiv1.EndpointAddress{
					{
						IP:       "10.100.3.2",
						NodeName: &instance3,
					},
					{
						IP:       "10.100.4.2",
						NodeName: &instance4,
					},
				},
				Ports: []apiv1.EndpointPort{
					{
						Name:     NamedPort,
						Port:     int32(8081),
						Protocol: apiv1.ProtocolTCP,
					},
				},
			},
		},
	}
}
