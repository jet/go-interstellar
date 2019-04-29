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
	"context"
	"encoding/json"

	"github.com/jet/go-interstellar"
)

func ExampleCollectionClient_QueryDocumentsRaw() {
	// error handling omitted for brevity

	// Document inside the collection for Unmarshaling
	type Document struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	// Get a client
	cs, _ := interstellar.ParseConnectionString("AccountEndpoint=https://localhost:8081/;AccountKey=C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw==")
	client, _ := interstellar.NewClient(cs, nil)

	// Construct a query which returns documents which have a name prefixed with `ab`, 10 per page.
	query := &interstellar.Query{
		Query: "SELECT * FROM Documents d WHERE STARTSWITH(d.name,@prefix)",
		Parameters: []interstellar.QueryParameter{
			interstellar.QueryParameter{Name: "@prefix", Value: "ab"},
		},
		MaxItemCount: 10,
	}

	// Results
	var docs []Document

	// Get the CollectionClient scoped to `dbs/db1/colls/col1`
	cc := client.WithDatabase("db1").WithCollection("col1")

	// Perform the query, and paginate through all the results
	cc.QueryDocumentsRaw(context.Background(), query, func(resList []json.RawMessage, meta interstellar.ResponseMetadata) (bool, error) {
		for _, raw := range resList {
			var doc Document
			if err := json.Unmarshal(raw, &doc); err != nil {
				return false, err
			}
			docs = append(docs, doc)
		}

		// true = get next page
		return true, nil
	})
}
