/*
Copyright 2018 The Kubernetes Authors.

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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"

	proto "github.com/kubernetes/ingress-nginx/images/grpc-fortune-teller/proto/fortune"
	"github.com/vromero/gofortune/lib/fortune"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = 50051
)

func main() {

	baseDir := "/tmp/fortune-teller"
	mustMkdirAll(baseDir)

	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(200),
	}

	grpcServer := grpc.NewServer(opts...)

	fortuneTeller := &FortuneTeller{
		fs: createFortuneFilesystemNodeDescriptor(baseDir),
	}
	proto.RegisterFortuneTellerServer(grpcServer, fortuneTeller)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Error while starting grpc server: %v\n", err)
	}

	log.Printf("Listening for gRPC requests at %d\n", grpcPort)
	grpcServer.Serve(lis)
}

// FortuneTeller - struct that will implement the grpc service interface.
type FortuneTeller struct {
	fs *fortune.FileSystemNodeDescriptor
}

// Predict - implementation for the grpc unary request method.
func (f *FortuneTeller) Predict(ctx context.Context, r *proto.PredictionRequest) (*proto.PredictionResponse, error) {
	_, data, err := fortune.GetRandomFortune(*f.fs)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Unable to render fortune: %v", err)
	}
	return &proto.PredictionResponse{
		Message: data,
	}, nil
}

func createFortuneFilesystemNodeDescriptor(baseDir string) *fortune.FileSystemNodeDescriptor {

	// Restore the packed fortune data
	fortuneDir := path.Join(baseDir, "usr/share/games/fortunes")

	mustRestore(baseDir, fortuneFiles, nil)

	// init gofortune fs
	fs, err := fortune.LoadPaths([]fortune.ProbabilityPath{
		{Path: fortuneDir},
	})
	if err != nil {
		log.Fatalf("Unable to load fortune paths: %v", err)
	}

	fortune.SetProbabilities(&fs, true) // consider all equal probabilities
	return &fs
}

// mustRestore - Restore assets.
func mustRestore(baseDir string, assets map[string][]byte, mappings map[string]string) {
	// unpack variable is provided by the go_embed data and is a
	// map[string][]byte such as {"/usr/share/games/fortune/literature.dat":
	// bytes... }
	for basename, bytes := range assets {
		if mappings != nil {
			replacement := mappings[basename]
			if replacement != "" {
				basename = replacement
			}
		}
		filename := path.Join(baseDir, basename)
		dirname := path.Dir(filename)
		//log.Printf("file %s, dir %s, rel %d, abs %s, absdir: %s", file, dir, rel, abs, absdir)
		if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
			log.Fatalf("Failed to create asset dir %s: %v", dirname, err)
		}

		if err := ioutil.WriteFile(filename, bytes, os.ModePerm); err != nil {
			log.Fatalf("Failed to write asset %s: %v", filename, err)
		}
		log.Printf("Restored %s", filename)
	}

	log.Printf("Assets restored to %s", baseDir)
}

// mustMkdirAll - make all dirs and panic if fail
func mustMkdirAll(dirs ...string) {
	for _, dir := range dirs {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			panic(fmt.Sprintf("Failed mkdir %s: %v", dir, err))
		}
	}
}
