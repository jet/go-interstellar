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

// Package interstellar is a client for using the SQL CosmoDB API over REST/HTTP.
// It provides a generic way of interacting with arbitrary resources in Cosmos DB via the Client
// And also convenience functions for interacting with resources types like Databases, Collections, Documents, and Offers.
//
//
package interstellar // import "github.com/jet/go-interstellar"

// APIVersion is the version of the CosmosDB REST API supported
// See https://docs.microsoft.com/en-us/rest/api/cosmos-db/#supported-rest-api-versions for available versions and changelog
const APIVersion = "2017-02-22"
