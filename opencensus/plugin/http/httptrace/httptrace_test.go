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

package httptrace

import (
	"encoding/binary"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"go.opencensus.io/trace"
)

type testTransport struct {
	ch chan *trace.Span
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.ch <- trace.FromContext(req.Context())
	return nil, errors.New("noop")
}

func TestTransport_RoundTrip(t *testing.T) {
	parent := trace.NewSpan("parent", trace.StartSpanOptions{})
	tests := []struct {
		name   string
		parent *trace.Span
	}{
		{
			name:   "no parent",
			parent: nil,
		},
		{
			name:   "parent",
			parent: parent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trt := &testTransport{ch: make(chan *trace.Span, 1)}
			rt := &Transport{Base: trt}

			req, _ := http.NewRequest("GET", "http://foo.com", nil)
			if tt.parent != nil {
				req = req.WithContext(trace.WithSpan(req.Context(), tt.parent))
			}
			rt.RoundTrip(req)

			span := <-trt.ch
			if span == nil {
				t.Fatalf("Got no spans in req context; want one")
			}
			if tt.parent != nil {
				if got, want := span.SpanContext().TraceID, tt.parent.SpanContext().TraceID; got != want {
					t.Errorf("span.SpanContext().TraceID=%v; want %v", got, want)
				}
			}
		})
	}
}

func TestSpanContext(t *testing.T) {
	traceID := [16]byte{16, 84, 69, 170, 120, 67, 188, 139, 242, 6, 177, 32, 0, 16, 0, 0}
	var spanID [8]byte
	binary.PutUvarint(spanID[:], 123)
	tests := []struct {
		sc       trace.SpanContext
		incoming string
	}{
		{
			incoming: "105445aa7843bc8bf206b12000100000/123;o=1",
			sc: trace.SpanContext{
				TraceID:      traceID,
				SpanID:       spanID,
				TraceOptions: 1,
			},
		},
		{
			incoming: "105445aa7843bc8bf206b12000100000/123;o=0",
			sc: trace.SpanContext{
				TraceID:      traceID,
				SpanID:       spanID,
				TraceOptions: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.incoming, func(t *testing.T) {
			spanContext, ok := traceInfoFromHeader(tt.incoming)
			if !ok {
				t.Errorf("traceInfoFromHeader() = false; want true")
			}
			if got, want := spanContext, tt.sc; !reflect.DeepEqual(got, want) {
				t.Errorf("traceInfoFromHeader() = %v; want %v", got, want)
			}
			if got := spanContextToHeader(spanContext); got != tt.incoming {
				t.Errorf("spanContextToHeader() = %q, want %q", got, tt.incoming)
			}
		})
	}
}
