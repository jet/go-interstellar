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
	"io/ioutil"
	"testing"
)

// ReadFileBytes reads the entire file, or calls t.Fatal if the file cannot be read
func ReadFileBytes(t *testing.T, file string) []byte {
	t.Helper()
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("unable to load file: %s: %v", file, err)
	}
	return bs
}