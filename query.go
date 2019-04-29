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
	"bytes"
	"fmt"
	"net/http"
)

// Query encapsulates a SQL-like query on the  Collection
// Parameters can be added by using the AddParameter method
// Additional query options can be addy by supplying a QueryOptions
type Query struct {
	// Query is the SQL-like query text.
	// See https://docs.microsoft.com/en-us/azure/cosmos-db/how-to-sql-query for the Grammar reference
	Query string `json:"query"`

	// Parameters are the collection of named query parameters and their values that may be referenced by the Query string.
	Parameters []QueryParameter `json:"parameters,omitempty"`

	// MaxItemCount sets the desired maximum number of items returned in a single page of results
	MaxItemCount int `json:"-"`

	// Continuation is used to get the next page of results from a query.
	// Setting this value, will set the corresponding x-ms-continuation header on the query request.
	// This value is opaque, and is returned when there are additional results aside from what was returned in the response.
	Continuation string `json:"-"`

	// EnableCrossPartition enables the query to span across multiple partitions.
	EnableCrossPartition bool `json:"-"`

	// ConsistencytLevel sets the consistency level override.
	// This must be the same or weaker than the account's configured consistency level.
	ConsistencytLevel ConsistencyLevel `json:"-"`

	// SessionToken must be set when using a consistency level of "Session".
	// The "SessionToken" recevied from a response must be echo'd back in the next request.
	SessionToken string `json:"-"`

	// RequestOptions applies additional request options to the query
	RequestOptions RequestOptions `json:"-"`
}

// AddParameter adds a new named parameter to the query
func (q *Query) AddParameter(name string, value interface{}) {
	q.Parameters = append(q.Parameters, QueryParameter{
		Name:  name,
		Value: value,
	})
}

// AddParameterSensitive adds a new named parameter to the query and sets the Sensitive flag
func (q *Query) AddParameterSensitive(name string, value interface{}) {
	q.Parameters = append(q.Parameters, QueryParameter{
		Name:      name,
		Value:     value,
		Sensitive: true,
	})
}

// String returns the string representation of the Query with its parameters.
// This is not used in the API, it is only useful for debugging.
func (q *Query) String() string {
	if q == nil {
		return ""
	}
	buf := &bytes.Buffer{}
	buf.WriteString(q.Query)
	buf.WriteString(";")
	if len(q.Parameters) > 0 {
		buf.WriteString(" [")
		for i, p := range q.Parameters {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(p.String())
		}
		buf.WriteString("]")
	}
	return buf.String()
}

// QueryParameter encapsulates a named query parameter for a  query along with its value
type QueryParameter struct {
	// Name is the name of the query parameter.
	// All names should begin with an '@'; such as '@id' or '@name'
	Name string `json:"name"`
	// Value is the corresponding value of the named parameter.
	// The value should be JSON-marshalable, such as a primitive type or a type that implements the json.Marshaler behavior
	Value interface{} `json:"value"`
	// Sensitive is a flag that, if set to true, prevents the Value from being printed as part of the String() method
	// This value is not used by the Query API, and is only useful in debugging
	Sensitive bool `json:"-"`
}

// String returns the query parameter as a string key=value
func (p QueryParameter) String() string {
	if p.Sensitive {
		return fmt.Sprintf("%s: !(sensitive)", p.Name)
	}
	return fmt.Sprintf("%s: %#v", p.Name, p.Value)
}

// Format implements fmt.Formatter for QueryParameter
func (p QueryParameter) Format(f fmt.State, c rune) {
	f.Write([]byte(p.String()))
}

// ApplyOptions applies the additional query options to the API request
func (q *Query) ApplyOptions(req *http.Request) {
	if q.SessionToken != "" {
		req.Header.Set(HeaderSessionToken, q.SessionToken)
	}
	if q.ConsistencytLevel != "" {
		req.Header.Set(HeaderConsistencyLevel, string(q.ConsistencytLevel))
	}
	if q.EnableCrossPartition {
		req.Header.Set(HeaderDocDBQueryEnableCrossPartition, "true")
	}
	if q.Continuation != "" {
		req.Header.Set(HeaderContinuation, q.Continuation)
	}
	if q.MaxItemCount != 0 {
		req.Header.Set(HeaderMaxItemCount, fmt.Sprintf("%d", q.MaxItemCount))
	}
}
