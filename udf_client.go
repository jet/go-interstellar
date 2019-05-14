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
	"fmt"
	"net/http"
	"net/url"
)

// UserDefinedFunctionResource represents a User Defined Function in Cosmos DB
// Documentation adapted from adapted from docs.microsoft.com
// See https://docs.microsoft.com/en-us/rest/api/cosmos-db/collections for the latest documentation
type UserDefinedFunctionResource struct {
	// ID is the unique user generated name for the UDF. The id must not exceed 255 characters.
	ID string `json:"id"`
	// ResourceID is a unique identifier that is also hierarchical per the resource stack on the resource model. It is used internally for placement of and navigation to the collection resource.
	ResourceID string `json:"_rid,omitempty"`
	// Timestamp is a system generated property. It denotes the last updated timestamp of the resource.
	Timestamp int64 `json:"_ts,omitempty"`
	// Self is the unique addressable URI for the resource.
	Self string `json:"_self,omitempty"`
	// ETag value required for optimistic concurrency control.
	ETag string `json:"_etag,omitempty"`
	// Body is the body of the User Defined Function (javascript)
	Body string `json:"body"`
}

// CreateUserDefinedFunctionRequest are parameters for CreateUserDefinedFunction
type CreateUserDefinedFunctionRequest struct {
	ID      string         `json:"id"`
	Body    string         `json:"body"`
	Options RequestOptions `json:"-"`
}

// ApplyOptions applies the request options to the api request
func (r CreateUserDefinedFunctionRequest) ApplyOptions(req *http.Request) {
	if r.Options != nil {
		r.Options.ApplyOptions(req)
	}
}

// CreateUserDefinedFunction creates a UDF
func (c *CollectionClient) CreateUserDefinedFunction(ctx context.Context, req CreateUserDefinedFunctionRequest) (*UserDefinedFunctionResource, *ResponseMetadata, error) {
	body, meta, err := c.createUserDefinedFunctionRaw(ctx, req)
	if err != nil {
		return nil, meta, err
	}
	var udf UserDefinedFunctionResource
	if err = json.Unmarshal(body, &udf); err != nil {
		return nil, meta, err
	}
	return &udf, meta, nil
}

func (c *CollectionClient) createUserDefinedFunctionRaw(ctx context.Context, req CreateUserDefinedFunctionRequest) ([]byte, *ResponseMetadata, error) {
	rl := fmt.Sprintf("dbs/%s/colls/%s", url.PathEscape(c.DatabaseID), url.PathEscape(c.CollectionID))
	body, err := json.Marshal(&req)
	if err != nil {
		return nil, nil, err
	}
	return c.Client.CreateOrReplaceResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s/udfs", rl),
		ResourceType: ResourceUserDefinedFunctions,
		ResourceLink: rl,
		Options:      req,
		Body:         bytes.NewBuffer(body),
	})
}

// PaginateUDFResource pagination function for a list of UserDefiendFunctions
type PaginateUDFResource func(resList []UserDefinedFunctionResource, meta ResponseMetadata) (bool, error)

// ListUserDefinedFunctions lists each User-defined Function in the collection
func (c *CollectionClient) ListUserDefinedFunctions(ctx context.Context, opts RequestOptions, fn PaginateUDFResource) error {
	return c.listUserDefinedFunctionsRaw(ctx, opts, func(resList []json.RawMessage, meta ResponseMetadata) (bool, error) {
		udfs := make([]UserDefinedFunctionResource, len(resList))
		for i, res := range resList {
			var udf UserDefinedFunctionResource
			if err := json.Unmarshal(res, &udf); err != nil {
				return false, err
			}
			udfs[i] = udf
		}
		return fn(udfs, meta)
	})
}

func (c *CollectionClient) listUserDefinedFunctionsRaw(ctx context.Context, opts RequestOptions, fn PaginateRawResources) error {
	rl := c.ResourceLink()
	return c.Client.ListResources(ctx, "UserDefinedFunctions", ClientRequest{
		Path:         fmt.Sprintf("/%s/udfs", rl),
		ResourceLink: rl,
		ResourceType: ResourceUserDefinedFunctions,
		Options:      opts,
	}, fn)
}

// UDFClient is a client scoped to a single user-defined function
// Used to perform API calls within the scope of the UDF resource
type UDFClient struct {
	Client       *Client
	DatabaseID   string
	CollectionID string
	UDFID        string
}

// WithUDF creates a UDFClient for the given UDF within this Collection
func (c *CollectionClient) WithUDF(id string) *UDFClient {
	return &UDFClient{
		Client:       c.Client,
		DatabaseID:   c.DatabaseID,
		CollectionID: c.CollectionID,
		UDFID:        id,
	}
}

// ResourceLink gets the resource link for the user-defined function
func (c *UDFClient) ResourceLink() string {
	return fmt.Sprintf("dbs/%s/colls/%s/udfs/%s", url.PathEscape(c.DatabaseID), url.PathEscape(c.CollectionID), url.PathEscape(c.UDFID))
}

// Replace replaces a UDF Body with the new one
func (c *UDFClient) Replace(ctx context.Context, body string, opts RequestOptions) (*UserDefinedFunctionResource, *ResponseMetadata, error) {
	bs, meta, err := c.replaceRaw(ctx, body, opts)
	if err != nil {
		return nil, meta, err
	}
	var udf UserDefinedFunctionResource
	if err = json.Unmarshal(bs, &udf); err != nil {
		return nil, meta, err
	}
	return &udf, meta, nil
}

func (c *UDFClient) replaceRaw(ctx context.Context, body string, opts RequestOptions) ([]byte, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	udf := UserDefinedFunctionResource{
		ID:   c.UDFID,
		Body: body,
	}
	bs, err := json.Marshal(&udf)
	if err != nil {
		return nil, nil, err
	}
	return c.Client.CreateOrReplaceResource(ctx, ClientRequest{
		Method:       http.MethodPut,
		Path:         fmt.Sprintf("/%s", rl),
		ResourceType: ResourceUserDefinedFunctions,
		ResourceLink: rl,
		Options:      opts,
		Body:         bytes.NewBuffer(bs),
	})
}

// Delete deletes the user-defined function
func (c *UDFClient) Delete(ctx context.Context, opts RequestOptions) (bool, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	return c.Client.DeleteResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s", rl),
		ResourceType: ResourceUserDefinedFunctions,
		ResourceLink: rl,
		Options:      opts,
	})
}
