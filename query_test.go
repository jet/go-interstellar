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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"testing"

	"github.com/jet/go-interstellar"
	"github.com/jet/go-interstellar/internal/testutil"
)

func TestQueryParameterFormatter(t *testing.T) {
	examples := []struct {
		param    interstellar.QueryParameter
		expected string
	}{
		{interstellar.QueryParameter{Name: "@id", Value: "123"}, `@id: "123"`},
		{interstellar.QueryParameter{Name: "@age", Value: 99}, `@age: 99`},
		{interstellar.QueryParameter{Name: "@pi", Value: 3.1415}, `@pi: 3.1415`},
		{interstellar.QueryParameter{Name: "@enabled", Value: true}, `@enabled: true`},
		{interstellar.QueryParameter{Name: "@flag", Value: byte(254)}, `@flag: 0xfe`},
		{interstellar.QueryParameter{Name: "@email", Value: byte(254), Sensitive: true}, `@email: !(sensitive)`},
	}
	for i, ex := range examples {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if fmt.Sprintf("%v", ex.param) != ex.expected {
				t.Fatalf("expected=%s\nactual='%s'", ex.param, ex.expected)
			}
		})
	}
}

func TestQueryFormatter(t *testing.T) {
	examples := []struct {
		query          *interstellar.Query
		expectedJSON   string
		expectedString string
	}{
		{query: nil, expectedJSON: `null`},
		{
			query: &interstellar.Query{
				Query: "SELECT * FROM Events e WHERE e.id = @id",
				Parameters: []interstellar.QueryParameter{
					interstellar.QueryParameter{Name: "@id", Value: "123"},
				},
			},
			expectedJSON:   `{"query":"SELECT * FROM Events e WHERE e.id = @id","parameters":[{"name":"@id","value":"123"}]}`,
			expectedString: `SELECT * FROM Events e WHERE e.id = @id; [@id: "123"]`,
		},
		{
			query: &interstellar.Query{
				Query: "SELECT * FROM Families f WHERE f.lastName = @lastName AND f.address.state = @addressState",
				Parameters: []interstellar.QueryParameter{
					interstellar.QueryParameter{Name: "@lastName", Value: "Wakefield"},
					interstellar.QueryParameter{Name: "@addressState", Value: "NY"},
				},
			},
			expectedJSON:   `{"query":"SELECT * FROM Families f WHERE f.lastName = @lastName AND f.address.state = @addressState","parameters":[{"name":"@lastName","value":"Wakefield"},{"name":"@addressState","value":"NY"}]}`,
			expectedString: `SELECT * FROM Families f WHERE f.lastName = @lastName AND f.address.state = @addressState; [@lastName: "Wakefield", @addressState: "NY"]`,
		},
		{
			query: &interstellar.Query{
				Query: "SELECT * FROM Customers c WHERE c.email = @email",
				Parameters: []interstellar.QueryParameter{
					interstellar.QueryParameter{Name: "@email", Value: "john.doe@example.com", Sensitive: true},
				},
			},
			expectedJSON:   `{"query":"SELECT * FROM Customers c WHERE c.email = @email","parameters":[{"name":"@email","value":"john.doe@example.com"}]}`,
			expectedString: `SELECT * FROM Customers c WHERE c.email = @email; [@email: !(sensitive)]`,
		},
	}
	for i, ex := range examples {
		t.Run(fmt.Sprintf("%d-string", i), func(t *testing.T) {
			if fmt.Sprintf("%v", ex.query) != ex.expectedString {
				t.Fatalf("expected=%s\nactual='%s'", ex.query, ex.expectedString)
			}
		})
		t.Run(fmt.Sprintf("%d-json", i), func(t *testing.T) {
			js, _ := json.Marshal(ex.query)
			if string(js) != ex.expectedJSON {
				t.Fatalf("expected=%s\nactual='%s'", string(js), ex.expectedJSON)
			}
		})
	}
}

func TestQueryRequestOptions(t *testing.T) {
	expectedFile := filepath.Join("./testdata", "query", "expected-request.txt")
	expected := testutil.ReadFileBytes(t, expectedFile)
	query := &interstellar.Query{
		Query:                "SELECT * FROM Families f WHERE f.lastName = @lastName AND f.address.state = @addressState",
		MaxItemCount:         100,
		SessionToken:         "0:102",
		Continuation:         `{"token":"+RID:YrMqAKFnpn5IAAAAAAAAAA==#RT:3#TRC:15#FPC:AgEAAAAGAEiAAcAAgg==","range":{"min":"","max":"FF"}}`,
		ConsistencytLevel:    interstellar.ConsistencyEventual,
		EnableCrossPartition: true,
		RequestOptions: interstellar.RequestOptionsFunc(func(req *http.Request) {
			req.Header.Set("x-request-id", "foo")
		}),
	}
	query.AddParameterSensitive("@lastName", "Wakefield")
	query.AddParameter("@addressState", "NY")

	qjson, _ := json.Marshal(query)
	client := &interstellar.Client{
		Endpoint:   "https://localhost:8081",
		Authorizer: testutil.TestKey("TESTING"),
		UserAgent:  "Test/1.0",
		Requester:  nil,
	}
	req, _ := client.NewHTTPRequest(nil, interstellar.ClientRequest{
		Method:       http.MethodPost,
		Path:         "/dbs/db1/colls/col1/docs",
		ResourceLink: "/dbs/db1/colls/col1",
		ResourceType: interstellar.ResourceDocuments,
		Options:      query,
		Body:         bytes.NewBuffer(qjson),
	})
	actual, _ := httputil.DumpRequest(req, true)

	if !bytes.Equal(expected, actual) {
		actualFile := filepath.Join("./testdata", "query", "actual-request.txt")
		ioutil.WriteFile(actualFile, actual, 0644)
		t.Fatalf("expected query request does not equal actual. Compare %s with %s", expectedFile, actualFile)
	}
}
