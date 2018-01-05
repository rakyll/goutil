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

// Package httptrace contains OpenCensus tracing integrations with net/http.
package httptrace

import (
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"go.opencensus.io/trace"
)

const httpHeader = `X-Cloud-Trace-Context`

// Transport is an http.RoundTripper that traces the outgoing requests.
type Transport struct {
	// Base is the base http.RoundTripper to be used to do the actual request.
	//
	// Optional. If nil, http.DefaultTransport is used.
	Base http.RoundTripper
}

// RoundTrip creates a trace.Span and inserts it into the outgoing request's headers.
// The created span can follow a parent span, if a parent is presented in
// the request's context.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	name := "Sent" + strings.Replace(req.URL.String(), req.URL.Scheme, ".", -1)
	ctx := trace.StartSpan(req.Context(), name)
	req = req.WithContext(ctx)

	span := trace.FromContext(ctx)
	req.Header.Set(httpHeader, spanContextToHeader(span.SpanContext()))

	resp, err := t.base().RoundTrip(req)

	// TODO(jbd): Add status and attributes.
	trace.EndSpan(ctx)
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

// Handler returns a http.Handler from the given handler
// that is aware of the incoming request's span.
// The span can be extracted from the incoming request in handler
// functions from incoming request's context:
//
//    span := trace.FromContext(r.Context())
//
// The span will be auto finished by the handler.
func Handler(base http.Handler) http.Handler {
	return &handler{handler: base}
}

type handler struct {
	handler http.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := "Recv" + strings.Replace(r.URL.String(), r.URL.Scheme, ".", -1)

	ctx := r.Context()
	traceID, spanID, options, _, ok := traceInfoFromHeader(r.Header.Get(httpHeader))
	if ok {
		ctx = trace.StartSpanWithRemoteParent(ctx, name, trace.SpanContext{
			TraceID:      traceID,
			SpanID:       spanID,
			TraceOptions: options,
		}, trace.StartSpanOptions{})
	} else {
		ctx = trace.StartSpan(r.Context(), name)
	}
	defer trace.EndSpan(ctx)

	// TODO(jbd): Add status and attributes.
	r = r.WithContext(ctx)
	h.handler.ServeHTTP(w, r)
}

func traceInfoFromHeader(h string) (traceID trace.TraceID, spanID trace.SpanID, options trace.TraceOptions, optionsOk bool, ok bool) {
	// See https://cloud.google.com/trace/docs/faq for the header format.
	// Return if the header is empty or missing, or if the header is unreasonably
	// large, to avoid making unnecessary copies of a large string.
	if h == "" || len(h) > 200 {
		return trace.TraceID{}, trace.SpanID{}, 0, false, false
	}

	// Parse the trace id field.
	slash := strings.Index(h, `/`)
	if slash == -1 {
		return trace.TraceID{}, trace.SpanID{}, 0, false, false

	}
	tid, h := h[:slash], h[slash+1:]

	buf, err := hex.DecodeString(tid)
	if err != nil {
		return trace.TraceID{}, trace.SpanID{}, 0, false, false
	}
	copy(traceID[:], buf)

	// Parse the span id field.
	spanstr := h
	semicolon := strings.Index(h, `;`)
	if semicolon != -1 {
		spanstr, h = h[:semicolon], h[semicolon+1:]
	}
	sid, err := strconv.ParseUint(spanstr, 10, 64)
	if err != nil {
		return trace.TraceID{}, trace.SpanID{}, 0, false, false
	}

	buf = make([]byte, 8)
	binary.PutUvarint(buf, sid)
	copy(spanID[:], buf)

	// Parse the options field, options field is optional.
	if !strings.HasPrefix(h, "o=") {
		return traceID, spanID, 0, false, true

	}
	o, err := strconv.ParseUint(h[2:], 10, 64)
	if err != nil {
		return trace.TraceID{}, trace.SpanID{}, 0, false, false

	}
	options = trace.TraceOptions(o)
	return traceID, spanID, options, true, true
}

func spanContextToHeader(sc trace.SpanContext) string {
	traceID := hex.EncodeToString(sc.TraceID[:])
	sid, _ := binary.Uvarint(sc.SpanID[:])
	spanID := strconv.FormatUint(sid, 10)
	opts := strconv.FormatInt(int64(sc.TraceOptions), 10)
	return traceID + "/" + spanID + ";" + opts
}
