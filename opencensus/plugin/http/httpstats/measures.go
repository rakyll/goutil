// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpstats

import (
	"log"

	"go.opencensus.io/stats"
)

const (
	unitByte        = "By"
	unitCount       = "1"
	unitMillisecond = "ms"
)

func init() {
	ClientErrorCount = createMeasureInt64("net/http/client/error_count", "HTTP client error count", unitCount)
	ClientRoundTripLatency = createMeasureFloat64("net/http/client/roundtrip_latency", "HTTP client round trip latency", unitMillisecond)
	ClientRequestBytes = createMeasureInt64("net/http/client/request_bytes", "HTTP client request size", unitByte)
	ClientResponseBytes = createMeasureInt64("net/http/client/response_bytes", "HTTP client response size", unitByte)
	ClientStartedCount = createMeasureInt64("net/http/client/started_count", "Number of started requests at HTTP client", unitCount)
	ClientFinishedCount = createMeasureInt64("net/http/client/finished_count", "Number of finished requests at HTTP client", unitCount)
	ClientRequestCount = createMeasureInt64("net/http/client/request_count", "Number of requests at HTTP client", unitCount)
	ClientResponseCount = createMeasureInt64("net/http/client/response_count", "Number of responses at HTTP client", unitCount)

	ServerErrorCount = createMeasureInt64("net/http/server/error_count", "HTTP server error count", unitCount)
	ServerElapsedTime = createMeasureFloat64("net/http/server/elapsed_time", "HTTP server elapsed time", unitMillisecond)
	ServerRequestBytes = createMeasureInt64("net/http/server/request_bytes", "HTTP server request size", unitByte)
	ServerResponseBytes = createMeasureInt64("net/http/server/response_bytes", "HTTP server  Response size", unitByte)
	ServerStartedCount = createMeasureInt64("net/http/server/started_count", "Number of started requests at HTTP server", unitCount)
	ServerFinishedCount = createMeasureInt64("net/http/server/finished_count", "Number of finished requests at HTTP server", unitCount)
	ServerRequestCount = createMeasureInt64("net/http/server/request_count", "Number of requests at HTTP server", unitCount)
	ServerResponseCount = createMeasureInt64("net/http/server/response_count", "Number of responses at HTTP server", unitCount)
}

func createMeasureInt64(name, desc, unit string) *stats.MeasureInt64 {
	m, err := stats.NewMeasureInt64(name, desc, unit)
	if err != nil {
		log.Fatalf("Cannot create measure %q: %v", name, err)
	}
	return m
}

func createMeasureFloat64(name, desc, unit string) *stats.MeasureFloat64 {
	m, err := stats.NewMeasureFloat64(name, desc, unit)
	if err != nil {
		log.Fatalf("Cannot create measure %q: %v", name, err)
	}
	return m
}
