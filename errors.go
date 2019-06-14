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

import "net/http"

// Error is an interstellar generated error
// This type is an alias for 'string' and is used to ensure the interstellar sential errors can be made constant
type Error string

// Error implements the error interface for the Error type
func (e Error) Error() string {
	return string(e)
}

// StatusCode relay the status code
func (e Error) Status() int {
	switch e {
	case ErrResourceNotModified:
		return http.StatusNotModified
	case ErrResourceNotFound:
		return http.StatusNotFound
	default:
		return 0
	}
}
