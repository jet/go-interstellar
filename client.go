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

package interstellar

import (
	"net/http"

	"github.com/jet/go-mantis/rest"
)

// DefaultUserAgent which is set on outgoing http requests if none is set on the client
const DefaultUserAgent = "Go-Interstellar/0.1"

// Client for making API calls against CosmosDB
type Client struct {
	UserAgent string
	Endpoint  string
	Authorizer
	Requester
}

// Requester is an interface for sending HTTP requests and receiving responses
// *http.Client implicitly implements this interface already
// But more complicated logic such as logging are possible here as well
type Requester interface {
	Do(req *http.Request) (*http.Response, error)
}

// Authorizer is an interface used to authorize Cosmos DB Requests
type Authorizer interface {
	Authorize(r *http.Request, resourceType ResourceType, resourceLink string) (*http.Request, error)
}

// NewClient creates client to the given CoasmosDB account in the ConnectionString
// And will use the Requester to send HTTP requests and read responses
//
func NewClient(cs ConnectionString, req Requester) (*Client, error) {
	if req == nil {
		req = rest.HTTPClient()
	}
	return &Client{
		UserAgent:  DefaultUserAgent,
		Endpoint:   cs.Endpoint,
		Authorizer: cs.AccountKey,
		Requester: &rest.RetryAfterRequester{
			// Defaults ...
			//StatusCodes: []int{http.StatusTooManyRequests},
			//HeaderName: "Retry-After",
			Requester: req,
		},
	}, nil
}
