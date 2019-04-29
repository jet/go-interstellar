// SPDX-License-Identifier: Apache-2.0
//
// Copyright (c) 2019-present, Jet.com, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License."

package testutil

import (
	"net/http"
	"net/http/httputil"
	"os"
	"testing"

	"github.com/jet/go-interstellar"
)

// TestLoggingRequester is a requester which logs each http request/response to the test runner logger
type TestLoggingRequester struct {
	T *testing.T
	interstellar.Requester
}

// NewTestLoggingRequester creates a new TestLoggingRequester which will log if the environment variable "DEBUG_LOGGING=Y" is
func NewTestLoggingRequester(t *testing.T, r interstellar.Requester) interstellar.Requester {
	if os.Getenv("DEBUG_LOGGING") != "Y" {
		return r
	}
	return TestLoggingRequester{T: t, Requester: r}
}

// Do runs the request and logs the request and response dumps to the Test logger
func (r TestLoggingRequester) Do(req *http.Request) (*http.Response, error) {
	r.T.Helper()
	debugreq, _ := httputil.DumpRequest(req, true)
	r.T.Logf("HTTP REQUEST\n%s", string(debugreq))
	resp, err := r.Requester.Do(req)
	debugres, _ := httputil.DumpResponse(resp, true)
	r.T.Logf("HTTP RESPONSE\n%s", string(debugres))
	return resp, err
}
