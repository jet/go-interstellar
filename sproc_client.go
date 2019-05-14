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

// StoredProcedureResource represents a Stored Procedure in Cosmos DB
// Documentation adapted from adapted from docs.microsoft.com
// See https://docs.microsoft.com/en-us/rest/api/cosmos-db/collections for the latest documentation
type StoredProcedureResource struct {
	// ID is the unique user generated name for the Stored Procedure. The id must not exceed 255 characters.
	ID string `json:"id"`
	// ResourceID is a unique identifier that is also hierarchical per the resource stack on the resource model. It is used internally for placement of and navigation to the collection resource.
	ResourceID string `json:"_rid,omitempty"`
	// Timestamp is a system generated property. It denotes the last updated timestamp of the resource.
	Timestamp int64 `json:"_ts,omitempty"`
	// Self is the unique addressable URI for the resource.
	Self string `json:"_self,omitempty"`
	// ETag value required for optimistic concurrency control.
	ETag string `json:"_etag,omitempty"`
	// Body is the body of the Stored Procedure (javascript)
	Body string `json:"body"`
}

// CreateStoredProcedureRequest are parameters for CreateStoredProcedure
type CreateStoredProcedureRequest struct {
	ID      string         `json:"id"`
	Body    string         `json:"body"`
	Options RequestOptions `json:"-"`
}

// ApplyOptions applies the request options to the api request
func (r CreateStoredProcedureRequest) ApplyOptions(req *http.Request) {
	if r.Options != nil {
		r.Options.ApplyOptions(req)
	}
}

// CreateStoredProcedure creates a Stored Procedure
func (c *CollectionClient) CreateStoredProcedure(ctx context.Context, req CreateStoredProcedureRequest) (*StoredProcedureResource, *ResponseMetadata, error) {
	body, meta, err := c.createStoredProcedureRaw(ctx, req)
	if err != nil {
		return nil, meta, err
	}
	var sproc StoredProcedureResource
	if err = json.Unmarshal(body, &sproc); err != nil {
		return nil, meta, err
	}
	return &sproc, meta, nil
}

func (c *CollectionClient) createStoredProcedureRaw(ctx context.Context, req CreateStoredProcedureRequest) ([]byte, *ResponseMetadata, error) {
	rl := fmt.Sprintf("dbs/%s/colls/%s", url.PathEscape(c.DatabaseID), url.PathEscape(c.CollectionID))
	body, err := json.Marshal(&req)
	if err != nil {
		return nil, nil, err
	}
	return c.Client.CreateOrReplaceResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s/sprocs", rl),
		ResourceType: ResourceStoredProcedures,
		ResourceLink: rl,
		Options:      req,
		Body:         bytes.NewBuffer(body),
	})
}

// PaginateSProcResource pagination function for a list of StoredProcedureResource
type PaginateSProcResource func(resList []StoredProcedureResource, meta ResponseMetadata) (bool, error)

// ListStoredProcedures lists each stored procedure in the collection
func (c *CollectionClient) ListStoredProcedures(ctx context.Context, opts RequestOptions, fn PaginateSProcResource) error {
	return c.listStoredProcedures(ctx, opts, func(resList []json.RawMessage, meta ResponseMetadata) (bool, error) {
		sprocs := make([]StoredProcedureResource, len(resList))
		for i, res := range resList {
			var sproc StoredProcedureResource
			if err := json.Unmarshal(res, &sproc); err != nil {
				return false, err
			}
			sprocs[i] = sproc
		}
		return fn(sprocs, meta)
	})
}

func (c *CollectionClient) listStoredProcedures(ctx context.Context, opts RequestOptions, fn PaginateRawResources) error {
	rl := c.ResourceLink()
	return c.Client.ListResources(ctx, "StoredProcedures", ClientRequest{
		Path:         fmt.Sprintf("/%s/sprocs", rl),
		ResourceLink: rl,
		ResourceType: ResourceStoredProcedures,
		Options:      opts,
	}, fn)
}

// SProcClient is a client scoped to a single stored procedure
// Used to perform API calls within the scope of the Stored Procedure resource
type SProcClient struct {
	Client       *Client
	DatabaseID   string
	CollectionID string
	SProcID      string
}

// WithStoredProcedure creates a SProcClient for the given Stored Procedure within this Collection
func (c *CollectionClient) WithStoredProcedure(id string) *SProcClient {
	return &SProcClient{
		Client:       c.Client,
		DatabaseID:   c.DatabaseID,
		CollectionID: c.CollectionID,
		SProcID:      id,
	}
}

// ResourceLink gets the resource link for the stored procedure
func (c *SProcClient) ResourceLink() string {
	return fmt.Sprintf("dbs/%s/colls/%s/sprocs/%s", url.PathEscape(c.DatabaseID), url.PathEscape(c.CollectionID), url.PathEscape(c.SProcID))
}

// Replace replaces a Stored Procedure Body with the new one
func (c *SProcClient) Replace(ctx context.Context, body string, opts RequestOptions) (*StoredProcedureResource, *ResponseMetadata, error) {
	resp, meta, err := c.replaceRaw(ctx, body, opts)
	if err != nil {
		return nil, meta, err
	}
	var nproc StoredProcedureResource
	if err = json.Unmarshal(resp, &nproc); err != nil {
		return nil, meta, err
	}
	return &nproc, meta, err
}

func (c *SProcClient) replaceRaw(ctx context.Context, body string, opts RequestOptions) ([]byte, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	bs, err := json.Marshal(&StoredProcedureResource{
		ID:   c.SProcID,
		Body: body,
	})
	if err != nil {
		return nil, nil, err
	}
	return c.Client.CreateOrReplaceResource(ctx, ClientRequest{
		Method:       http.MethodPut,
		Path:         fmt.Sprintf("/%s", rl),
		ResourceType: ResourceStoredProcedures,
		ResourceLink: rl,
		Options:      opts,
		Body:         bytes.NewBuffer(bs),
	})
}

// Execute the stored procedure and return the raw result body
func (c *SProcClient) Execute(ctx context.Context, opts RequestOptions, args ...interface{}) ([]byte, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	bs, err := json.Marshal(args)
	if err != nil {
		return nil, nil, err
	}
	// reuse CreateOrReplaceResource since it will call a POST
	return c.Client.CreateOrReplaceResource(ctx, ClientRequest{
		Method:       http.MethodPost,
		Path:         fmt.Sprintf("/%s", rl),
		ResourceType: ResourceStoredProcedures,
		ResourceLink: rl,
		Options:      opts,
		Body:         bytes.NewBuffer(bs),
	})
}

// Func returns a function that can be called with with the stored procedures expected arguments, and returns the raw body
// The returned function takes a context object as its first parameter for cancellation/deadline
// The rest of the parameters are passed directly to the stored procedure (after being marshalled to JSON)
func (c *SProcClient) Func(opts RequestOptions) func(context.Context, ...interface{}) ([]byte, *ResponseMetadata, error) {
	return func(ctx context.Context, args ...interface{}) ([]byte, *ResponseMetadata, error) {
		return c.Execute(ctx, opts, args...)
	}
}

// Delete deletes the stored procedure
func (c *SProcClient) Delete(ctx context.Context, opts RequestOptions) (bool, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	return c.Client.DeleteResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s", rl),
		ResourceType: ResourceStoredProcedures,
		ResourceLink: rl,
		Options:      opts,
	})
}
