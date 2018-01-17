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

// Package httpstats contains OpenCensus stats integrations with net/http.
package httpstats

import (
	"net/http"

	"go.opencensus.io/stats"
)

var (
	// Available client measures
	ClientErrorCount       *stats.MeasureInt64
	ClientRoundTripLatency *stats.MeasureFloat64
	ClientRequestBytes     *stats.MeasureInt64
	ClientResponseBytes    *stats.MeasureInt64
	ClientStartedCount     *stats.MeasureInt64
	ClientFinishedCount    *stats.MeasureInt64
	ClientRequestCount     *stats.MeasureInt64
	ClientResponseCount    *stats.MeasureInt64

	// Available server measures
	ServerErrorCount    *stats.MeasureInt64
	ServerElapsedTime   *stats.MeasureFloat64
	ServerRequestBytes  *stats.MeasureInt64
	ServerResponseBytes *stats.MeasureInt64
	ServerStartedCount  *stats.MeasureInt64
	ServerFinishedCount *stats.MeasureInt64
	ServerRequestCount  *stats.MeasureInt64
	ServerResponseCount *stats.MeasureInt64
)

// Transport is an http.RoundTripper that traces the outgoing requests.
type Transport struct {
	// Base is the base http.RoundTripper to be used to do the actual request.
	//
	// Optional. If nil, http.DefaultTransport is used.
	Base http.RoundTripper
}

// RoundTrip records stats about the request.
// If request context contains any tags, stats will be recorded by them.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base().RoundTrip(req)
	return resp, err
}

// CancelRequest cancels an in-flight request by closing its connection.
func (t *Transport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := t.base().(canceler); ok {
		cr.CancelRequest(req)
	}
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// NewHandler returns a http.Handler that records stats for
// the incoming requests.
// If the incoming request contains any tags, stats will be recorded by them.
func NewHandler(base http.Handler) http.Handler {
	return &handler{handler: base}
}

type handler struct {
	handler http.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}
