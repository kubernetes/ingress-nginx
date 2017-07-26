/* Copyright 2017 Cody Maloney

Implements lyft/envoy Discovery Service APIs backed with DiscovyerInfo data
 - [LDS](https://lyft.github.io/envoy/docs/configuration/listeners/lds.html)
 - [RDS](https://lyft.github.io/envoy/docs/configuration/http_conn_man/rds.html)
 - [SDS](https://lyft.github.io/envoy/docs/configuration/cluster_manager/sds.html)
 - [CDS](https://lyft.github.io/envoy/docs/configuration/cluster_manager/cds.html)

TODO:
    Once lyft/envoy v2 API which is push/grpc based is done, move to ingress push to that, rather
    than serving an HTTP API which we then tell envoy to just agressively poll. */

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"unsafe"

	"github.com/gorilla/mux"
)

type discoveryService struct {
	responses      unsafe.Pointer
	server         http.Server
	envoyStartFunc func()
}

func (ds *discoveryService) Start() {
	log.Print("Starting discovery service server")
	go func() {
		if err := ds.server.ListenAndServe(); err != nil {
			log.Panicf("Unable to ListenAndServe: %s", err)
		}
	}()
	// TODO(cmaloney): delay this until the listener is definitely listening?
	ds.envoyStartFunc()
}

func (ds *discoveryService) Stop() {
	if err := ds.server.Shutdown(nil); err != nil {
		log.Panicf("Unable to shutdown discovery service server: %s", err)
	}
}

func (ds *discoveryService) GetResponsesToSwap() *unsafe.Pointer {
	return &ds.responses
}

func (ds *discoveryService) sendResponse(endpointName string, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		log.Printf("Error encoding json for %s response: %s", endpointName, err)
	}
}

func (ds *discoveryService) HandleRdsRequest(w http.ResponseWriter, r *http.Request) {
	// TODO(cmaloney): Allow multiple RouteConfigName and dispatch properly
	ds.sendResponse("RDS", w, (*Responses)(ds.responses).Rds)
}

func (ds *discoveryService) HandleLdsRequest(w http.ResponseWriter, r *http.Request) {
	ds.sendResponse("LDS", w, (*Responses)(ds.responses).Lds)
}

func (ds *discoveryService) HandleCdsRequest(w http.ResponseWriter, r *http.Request) {
	ds.sendResponse("CDS", w, (*Responses)(ds.responses).Cds)
}

func (ds *discoveryService) HandleSdsRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName, ok := vars["serviceName"]
	if !ok {
		log.Panicf("Unable to find serviceName in vars: %+v", vars)
	}

	// Unknown service/cluster -> empty set of endpoints.
	hosts, ok := (*Responses)(ds.responses).Sds[serviceName]
	if !ok {
		hosts = SdsResponse{Hosts: []EnvoyHost{}}
	}

	ds.sendResponse("SDS", w, hosts)
}

func NewDiscoveryService(envoyStartFunc func()) *discoveryService {
	ds := &discoveryService{envoyStartFunc: envoyStartFunc}

	r := mux.NewRouter()
	// RDS
	r.HandleFunc("/v1/routes/{routeConfigName}/{serviceCluster}/{serviceNode}", ds.HandleRdsRequest)
	// LDS
	r.HandleFunc("/v1/listeners/{serviceCluster}/{serviceNode}", ds.HandleLdsRequest)
	// CDS
	r.HandleFunc("/v1/clusters/{serviceCluster}/{serviceNode}", ds.HandleCdsRequest)
	// SDS
	r.HandleFunc("/v1/registration/{serviceName}", ds.HandleSdsRequest)

	ds.server = http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      r,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}
	return ds
}
