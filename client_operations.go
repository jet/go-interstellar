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
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/jet/go-mantis/rest"
	"github.com/pkg/errors"
)

const (
	// ErrPreconditionFailed is returned when an optimistic concurrency check fails
	ErrPreconditionFailed = Error("interstellar: request precondition failed")

	// ErrResourceNotFound is returned when a resource is not found
	ErrResourceNotFound = Error("interstellar: resource not found")

	// ErrResourceNotModified is returned from an http status code 304
	ErrResourceNotModified = Error("interstellar: resource not modified")
)

// PaginateRawResources is run by the List* operations with each page of results from the API.
// Returning `false` from the function will stop pagination and return a nil error.
// Returning a non-nil `error` from this function will stop pagination and return the error
type PaginateRawResources func(resList []json.RawMessage, meta ResponseMetadata) (bool, error)

// CreateOrReplaceResource creates new or replaces existing resources inside a given collection.
//
// If the ClientRequest.Method is not set, it will default to POST.
// If it is given, it must be PUT or POST; otherwise an error will be returned.
//
// For example, this can be used to create a new collection inside a database, a new document inside a collection, or update a document with new data.
func (c *Client) CreateOrReplaceResource(ctx context.Context, request ClientRequest) ([]byte, *ResponseMetadata, error) {
	request.Method = strings.ToUpper(request.Method)
	switch request.Method {
	case "":
		// default = POST
		request.Method = http.MethodPost
	case http.MethodPost, http.MethodPut:
		// valid
	default:
		return nil, nil, errors.Errorf("interstellar: Invalid request method '%s'; must be either PUT or POST", request.Method)
	}
	req, err := c.NewHTTPRequest(ctx, request)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.Requester.Do(req)
	if err != nil {
		return nil, nil, err
	}
	meta := GetResponseMetadata(resp)
	switch resp.StatusCode {
	case http.StatusOK:
		fallthrough
	case http.StatusCreated:
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		return body, &meta, err
	case http.StatusPreconditionFailed:
		return nil, &meta, ErrPreconditionFailed
	default:
		return nil, &meta, rest.NewErrorHTTPResponse(resp)
	}
}

// GetResource retrieves the body of a resource given by the request
// For example, this can be used to get the full body of a Collection resource, or a Document resource by it ID
func (c *Client) GetResource(ctx context.Context, request ClientRequest) ([]byte, *ResponseMetadata, error) {
	request.Method = http.MethodGet
	req, err := c.NewHTTPRequest(ctx, request)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.Requester.Do(req)
	if err != nil {
		return nil, nil, err
	}
	meta := GetResponseMetadata(resp)
	switch resp.StatusCode {
	case http.StatusOK:
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, &meta, err
		}
		return body, &meta, nil
	case http.StatusPreconditionFailed:
		return nil, &meta, ErrPreconditionFailed
	case http.StatusNotFound:
		resp.Body.Close()
		return nil, &meta, ErrResourceNotFound
	default:
		return nil, &meta, rest.NewErrorHTTPResponse(resp)
	}
}

func requestIsQuery(req *http.Request) {
	req.Header.Set(HeaderContentType, ContentTypeQueryJSON)
	req.Header.Set(HeaderDocDBIsQuery, "true")
}

// ListResources executes requests on a collection results to paginate through all the sub-resources.
// For example, this can be used to paginate through all documents within a collection.
//
// After each successful result list is returned, the array of objects are extracted from the given Key and passed to the supplied PaginateRawResources function.
// If PaginateRawResources function returns (true, nil), the next page will be requested
// If PaginateRawResources function returns (false, nil), then pagination will stop, and ListResults will return without error.
// If PaginateRawResources function returns a non-nil error, then pagination will stop, and ListResults will return that error.
// Pagination will also stop after the last page is returned from the API
func (c *Client) ListResources(ctx context.Context, key string, request ClientRequest, fn PaginateRawResources) error {
	prequest := &request
	prequest.Method = strings.ToUpper(request.Method)
	var body []byte
	switch request.Method {
	case "":
		// default = Get
		prequest.Method = http.MethodGet
	case http.MethodPost:
		// query

		// read entire query string
		data, err := prequest.readEntireBody()
		if err != nil {
			return err
		}
		body = data

		if request.Options == nil {
			request.Options = RequestOptionsFunc(requestIsQuery)
		}
		request.Options = RequestOptionsList{
			request.Options,
			RequestOptionsFunc(requestIsQuery),
		}
	default:
		return errors.Errorf("interstellar: Invalid request method '%s'; must be either GET or POST", request.Method)
	}
	req, err := c.NewHTTPRequest(ctx, request)
	if err != nil {
		return err
	}
	for {
		resp, err := c.Requester.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNotModified {
				return ErrResourceNotModified
			}
			return rest.NewErrorHTTPResponse(resp)
		}
		meta := GetResponseMetadata(resp)
		results, err := ParseArrayFromResponse(resp.Body, key)
		resp.Body.Close()
		if err != nil {
			return err
		}
		ok, err := fn(results, meta)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if cont := resp.Header.Get(HeaderContinuation); cont != "" {
			if body != nil {
				// reset query string body
				request.Body = bytes.NewBuffer(body)
			}
			req, err = c.NewHTTPRequest(ctx, request)
			if err != nil {
				return err
			}
			req.Header.Set(HeaderSessionToken, meta.SessionToken)
			req.Header.Set(HeaderContinuation, meta.Continuation)
			continue
		}
		break
	}
	return nil
}

// DeleteResource issues a delete command against a resource designate by the request
func (c *Client) DeleteResource(ctx context.Context, request ClientRequest) (bool, *ResponseMetadata, error) {
	request.Method = http.MethodDelete
	req, err := c.NewHTTPRequest(ctx, request)
	if err != nil {
		return false, nil, err
	}
	resp, err := c.Requester.Do(req)
	if err != nil {
		return false, nil, err
	}
	meta := GetResponseMetadata(resp)
	switch resp.StatusCode {
	case http.StatusNoContent:
		resp.Body.Close()
		return true, &meta, nil
	case http.StatusPreconditionFailed:
		return false, &meta, ErrPreconditionFailed
	case http.StatusNotFound:
		resp.Body.Close()
		return false, &meta, ErrResourceNotFound
	default:
		return false, &meta, rest.NewErrorHTTPResponse(resp)
	}
}
