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
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jet/go-interstellar"
	"github.com/jet/go-interstellar/internal/testutil"
	"github.com/jet/go-interstellar/internal/testutil/deep"
)

func TestDatabaseResourceMarshallJSON(t *testing.T) {
	testdata := testutil.ReadFileBytes(t, filepath.Join("./testdata", "databases.json"))
	var tests []json.RawMessage
	if err := json.Unmarshal(testdata, &tests); err != nil {
		t.Fatal(err)
	}
	for i, test := range tests {
		test = testutil.FormatJSON(test)
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var resource interstellar.DatabaseResource
			if err := json.Unmarshal([]byte(test), &resource); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(&resource)
			if err != nil {
				t.Fatal(err)
			}
			data = testutil.FormatJSON(data)
			if len(data) != len(test) {
				t.Errorf("data length (%d) != test length (%d)", len(test), len(data))
				t.Logf("\n%s\n%s\n", string(test), string(data))
			}
			var resource2 interstellar.DatabaseResource
			if err := json.Unmarshal(data, &resource2); err != nil {
				t.Fatal(err)
			}
			if diff := deep.Equal(&resource, &resource2); diff != nil {
				t.Fatal(diff)
			}
		})
	}
}
