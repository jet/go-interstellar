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
	"testing"
)

func TestErrorString(t *testing.T) {
	err := Error("test error")
	if err.Error() != "test error" {
		t.Fatalf("constant equality check failed")
	}
	if err != Error("test error") {
		t.Fatalf("constant equality check failed")
	}
}

func TestErrorStatus(t *testing.T) {
	var err error = ErrResourceNotModified
	type hasStatus interface{ Status() int }
	if hs, ok := err.(hasStatus); !ok || hs.Status() != http.StatusNotModified {
		t.Fatalf("constant equality check failed")
	}
	err = ErrResourceNotFound
	if hs, ok := err.(hasStatus); !ok || hs.Status() != http.StatusNotFound {
		t.Fatalf("constant equality check failed")
	}
}
