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
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

const (
	// ErrKeyNotFound is returned when the required key was missing from the response object
	ErrKeyNotFound = Error("interstellar: key was not found in the response object")
)

// ParseObjectResponse parses the response into a json object
// An error will be returend if the read bytes do not parse to an object with string keys
//
// Example Input JSON:
//
//     {
//     	 "key1": [1,2,"3",true],
//     	 "key2": "foo",
//     	 "key3": { "bar": "baz" }
//     }
//
func ParseObjectResponse(r io.Reader) (map[string]json.RawMessage, error) {
	dec := json.NewDecoder(r)
	var obj map[string]json.RawMessage
	if err := dec.Decode(&obj); err != nil {
		return nil, errors.Wrapf(err, "interstellar: could not decode json into map")
	}
	return obj, nil
}

// ParseArrayResponse parses the response into a JSON array
//
// Example Input JSON:
//
//     [1,2,"3",true]
//
func ParseArrayResponse(r io.Reader) ([]json.RawMessage, error) {
	dec := json.NewDecoder(r)
	var arr []json.RawMessage
	if err := dec.Decode(&arr); err != nil {
		return nil, errors.Wrapf(err, "interstellar: could not decode json into slice")
	}
	return arr, nil
}

// ParseArrayFromResponse parses a list of results from the response object mapped by given key.
// An error will be returend if the read bytes do not parse to a JSON object containing the desired key => list pair
//
// Example Input JSON:
//
//     { "key": [1,2,"3",true] }
//
// If the key is not found in the object, ErrKeyNotFound is returned
func ParseArrayFromResponse(r io.Reader, key string) ([]json.RawMessage, error) {
	obj, err := ParseObjectResponse(r)
	if err != nil {
		return nil, err
	}
	rawlist, ok := obj[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return ParseArrayResponse(bytes.NewReader(rawlist))
}
