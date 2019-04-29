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

	"github.com/jet/go-interstellar"
)

// TestKey is a test authorizer which just sets the authorization header to whatever its value is
type TestKey string

// Authorize sets the authorization header on the request
func (k TestKey) Authorize(r *http.Request, resourceType interstellar.ResourceType, resourceLink string) (*http.Request, error) {
	r.Header.Set(interstellar.HeaderAuthorization, string(k))
	return r, nil
}
