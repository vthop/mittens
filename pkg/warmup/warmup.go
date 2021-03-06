//Copyright 2019 Expedia, Inc.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package warmup

import (
	"fmt"
	"log"
	"mittens/pkg/grpc"
	"mittens/pkg/http"
	"sync"
	"time"
)

type Options struct {
	TimeoutSeconds int
	Concurrency    int
}

type Warmup struct {
	target  Target
	options Options
}

type Response struct {
	Type     string
	Error    error
	Duration time.Duration
}

func NewWarmup(readinessHttpClient http.Client, readinessGrpcClient grpc.Client, httpClient http.Client, grpcClient grpc.Client, options Options, targetOptions TargetOptions,
	done <-chan struct{}) (Warmup, error) {

	target, err := NewTarget(readinessHttpClient, readinessGrpcClient, httpClient, grpcClient, targetOptions, done)
	if err != nil {
		return Warmup{}, fmt.Errorf("new target: %v", err)
	}
	return Warmup{target: target, options: options}, err
}

func (w Warmup) HttpWarmup(headers map[string]string, requests <-chan http.Request) <-chan Response {

	response := make(chan Response)
	go func() {
		var wg sync.WaitGroup
		for request := range requests {
			go func(r http.Request) {
				wg.Add(1)
				defer wg.Done()
				startTime := time.Now()
				err := w.target.httpClient.Request(r.Method, r.Path, headers, r.Body)
				response <- Response{Type: "http", Error: err, Duration: time.Now().Sub(startTime)}
			}(request)
		}
		wg.Wait()
		close(response)
	}()
	return response
}

func (w Warmup) GrpcWarmup(headers []string, requests <-chan grpc.Request) <-chan Response {

	response := make(chan Response)
	go func() {
		var wg sync.WaitGroup
		for request := range requests {
			go func(r grpc.Request) {
				wg.Add(1)
				defer wg.Done()
				startTime := time.Now()
				err := w.target.grpcClient.Request(r.ServiceMethod, r.Message, headers)
				response <- Response{Type: "grpc", Error: err, Duration: time.Now().Sub(startTime)}
			}(request)
		}
		wg.Wait()
		if err := w.target.grpcClient.Close(); err != nil {
			log.Printf("grpc client close: %v", err)
		}
		close(response)
	}()
	return response
}
