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

package interstellar_test

import (
	"net/http"

	"github.com/jet/go-interstellar"
)

func ExampleClient_custom() {
	key, _ := interstellar.ParseMasterKey("C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw==")
	_ = &interstellar.Client{
		UserAgent:  interstellar.DefaultUserAgent,
		Endpoint:   "https://localhost:8081",
		Authorizer: key,
		Requester:  http.DefaultClient,
	}
}

func ExampleNewClient() {
	cstring := "AccountEndpoint=https://localhost:8081/;AccountKey=C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
	cs, _ := interstellar.ParseConnectionString(cstring)
	_, _ = interstellar.NewClient(cs, nil)
}
