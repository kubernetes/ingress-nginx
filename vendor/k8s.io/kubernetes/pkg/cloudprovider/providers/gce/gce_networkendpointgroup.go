/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gce

import (
	"github.com/golang/glog"
	computealpha "google.golang.org/api/compute/v0.alpha"
	"strings"
)

const (
	NEGLoadBalancerType          = "LOAD_BALANCING"
	NEGIPPortNetworkEndpointType = "GCE_VM_IP_PORT"
)

func newNetworkEndpointGroupMetricContext(request string, zone string) *metricContext {
	return newGenericMetricContext("networkendpointgroup_", request, unusedMetricLabel, zone, computeAlphaVersion)
}

func (gce *GCECloud) GetNetworkEndpointGroup(name string, zone string) (*computealpha.NetworkEndpointGroup, error) {
	if err := gce.alphaFeatureEnabled(AlphaFeatureNetworkEndpointGroup); err != nil {
		return nil, err
	}
	mc := newNetworkEndpointGroupMetricContext("get", zone)
	v, err := gce.serviceAlpha.NetworkEndpointGroups.Get(gce.GetProjectID(), zone, name).Do()
	return v, mc.Observe(err)
}

func (gce *GCECloud) ListNetworkEndpointGroup(zone string) ([]*computealpha.NetworkEndpointGroup, error) {
	if err := gce.alphaFeatureEnabled(AlphaFeatureNetworkEndpointGroup); err != nil {
		return nil, err
	}
	mc := newNetworkEndpointGroupMetricContext("list", zone)
	networkEndpointGroups := []*computealpha.NetworkEndpointGroup{}
	pageToken := ""
	page := 0
	for ; page == 0 || (pageToken != "" && page < maxPages); page++ {
		listCall := gce.serviceAlpha.NetworkEndpointGroups.List(gce.GetProjectID(), zone)

		if pageToken != "" {
			listCall.PageToken(pageToken)
		}

		res, err := listCall.Do()
		mc.Observe(err)
		if err != nil {
			glog.Errorf("Error listing network endpoint group from GCE: %v", err)
			return nil, err
		}
		pageToken = res.NextPageToken

		networkEndpointGroups = append(networkEndpointGroups, res.Items...)

		if page >= maxPages {
			glog.Errorf("ListNetworkEndpointGroup exceeded maxPages=%d: truncating.", maxPages)
		}
	}
	return networkEndpointGroups, nil
}

func (gce *GCECloud) AggregatedListNetworkEndpointGroup() (map[string][]*computealpha.NetworkEndpointGroup, error) {
	if err := gce.alphaFeatureEnabled(AlphaFeatureNetworkEndpointGroup); err != nil {
		return nil, err
	}
	mc := newNetworkEndpointGroupMetricContext("aggregated_list", "")
	zoneNetworkEndpointGroupMap := map[string][]*computealpha.NetworkEndpointGroup{}
	pageToken := ""
	page := 0
	for ; page == 0 || (pageToken != "" && page < maxPages); page++ {
		listCall := gce.serviceAlpha.NetworkEndpointGroups.AggregatedList(gce.GetProjectID())

		if pageToken != "" {
			listCall.PageToken(pageToken)
		}

		res, err := listCall.Do()
		mc.Observe(err)
		if err != nil {
			glog.Errorf("Error listing network endpoint group from GCE: %v", err)
			return nil, err
		}
		pageToken = res.NextPageToken

		for key, negs := range res.Items {
			if len(negs.NetworkEndpointGroups) == 0 {
				continue
			}
			// key has the format of "zones/${zone_name}"
			zone := strings.Split(key, "/")[1]
			if _, ok := zoneNetworkEndpointGroupMap[zone]; !ok {
				zoneNetworkEndpointGroupMap[zone] = []*computealpha.NetworkEndpointGroup{}
			}
			zoneNetworkEndpointGroupMap[zone] = append(zoneNetworkEndpointGroupMap[zone], negs.NetworkEndpointGroups...)
		}

		if page >= maxPages {
			glog.Errorf("ListNetworkEndpointGroup exceeded maxPages=%d: truncating.", maxPages)
		}
	}
	return zoneNetworkEndpointGroupMap, nil
}

func (gce *GCECloud) CreateNetworkEndpointGroup(neg *computealpha.NetworkEndpointGroup, zone string) error {
	if err := gce.alphaFeatureEnabled(AlphaFeatureNetworkEndpointGroup); err != nil {
		return err
	}
	mc := newNetworkEndpointGroupMetricContext("create", zone)
	op, err := gce.serviceAlpha.NetworkEndpointGroups.Insert(gce.GetProjectID(), zone, neg).Do()
	if err != nil {
		return mc.Observe(err)
	}
	return gce.waitForZoneOp(op, zone, mc)
}

func (gce *GCECloud) DeleteNetworkEndpointGroup(name string, zone string) error {
	if err := gce.alphaFeatureEnabled(AlphaFeatureNetworkEndpointGroup); err != nil {
		return err
	}
	mc := newNetworkEndpointGroupMetricContext("delete", zone)
	op, err := gce.serviceAlpha.NetworkEndpointGroups.Delete(gce.GetProjectID(), zone, name).Do()
	if err != nil {
		return mc.Observe(err)
	}
	return gce.waitForZoneOp(op, zone, mc)
}

func (gce *GCECloud) AttachNetworkEndpoints(name, zone string, endpoints []*computealpha.NetworkEndpoint) error {
	if err := gce.alphaFeatureEnabled(AlphaFeatureNetworkEndpointGroup); err != nil {
		return err
	}
	mc := newNetworkEndpointGroupMetricContext("attach", zone)
	op, err := gce.serviceAlpha.NetworkEndpointGroups.AttachNetworkEndpoints(gce.GetProjectID(), zone, name, &computealpha.NetworkEndpointGroupsAttachEndpointsRequest{
		NetworkEndpoints: endpoints,
	}).Do()
	if err != nil {
		return mc.Observe(err)
	}
	return gce.waitForZoneOp(op, zone, mc)
}

func (gce *GCECloud) DetachNetworkEndpoints(name, zone string, endpoints []*computealpha.NetworkEndpoint) error {
	if err := gce.alphaFeatureEnabled(AlphaFeatureNetworkEndpointGroup); err != nil {
		return err
	}
	mc := newNetworkEndpointGroupMetricContext("detach", zone)
	op, err := gce.serviceAlpha.NetworkEndpointGroups.DetachNetworkEndpoints(gce.GetProjectID(), zone, name, &computealpha.NetworkEndpointGroupsDetachEndpointsRequest{
		NetworkEndpoints: endpoints,
	}).Do()
	if err != nil {
		return mc.Observe(err)
	}
	return gce.waitForZoneOp(op, zone, mc)
}

func (gce *GCECloud) ListNetworkEndpoints(name, zone string, showHealthStatus bool) ([]*computealpha.NetworkEndpointWithHealthStatus, error) {
	if err := gce.alphaFeatureEnabled(AlphaFeatureNetworkEndpointGroup); err != nil {
		return nil, err
	}
	healthStatus := "SKIP"
	if showHealthStatus {
		healthStatus = "SHOW"
	}
	mc := newNetworkEndpointGroupMetricContext("list_networkendpoints", zone)
	networkEndpoints := []*computealpha.NetworkEndpointWithHealthStatus{}
	pageToken := ""
	page := 0
	for ; page == 0 || (pageToken != "" && page < maxPages); page++ {
		listCall := gce.serviceAlpha.NetworkEndpointGroups.ListNetworkEndpoints(gce.GetProjectID(), zone, name, &computealpha.NetworkEndpointGroupsListEndpointsRequest{
			HealthStatus: healthStatus,
		})
		if pageToken != "" {
			listCall.PageToken(pageToken)
		}

		res, err := listCall.Do()
		mc.Observe(err)
		if err != nil {
			return nil, err
		}
		pageToken = res.NextPageToken

		networkEndpoints = append(networkEndpoints, res.Items...)

		if page >= maxPages {
			glog.Errorf("ListNetworkEndpoints exceeded maxPages=%d: truncating.", maxPages)
		}
	}
	return networkEndpoints, nil
}
