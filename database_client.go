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
	"net/url"
)

// DatabaseResource represents a Database in Cosmos DB
// Documentation adapted from adapted from docs.microsoft.com
// See https://docs.microsoft.com/en-us/rest/api/cosmos-db/databases for the latest documentation
type DatabaseResource struct {
	// ID is the unique user generated name for the database.
	ID string `json:"id"`
	// ResourceID is a unique identifier that is also hierarchical per the resource stack on the resource model. It is used internally for placement of and navigation to the database resource.
	ResourceID string `json:"_rid,omitempty"`
	// Timestamp is a system generated property. It denotes the last updated timestamp of the resource.
	Timestamp int64 `json:"_ts,omitempty"`
	// Self is the unique addressable URI for the resource.
	Self string `json:"_self,omitempty"`
	// ETag value required for optimistic concurrency control.
	ETag string `json:"_etag,omitempty"`
	// Colls specifies the addressable path of the collections resource.
	Colls string `json:"_colls,omitempty"`
	// Users specifies the addressable path of the users resource.
	Users string `json:"_users,omitempty"`
}

// CreateDatabaseRaw creates a new database with the given ID and returns the raw response
func (c *Client) CreateDatabaseRaw(ctx context.Context, id string, opts RequestOptions) ([]byte, *ResponseMetadata, error) {
	body, err := json.Marshal(DatabaseResource{ID: id})
	if err != nil {
		return nil, nil, err
	}
	return c.CreateOrReplaceResource(ctx, ClientRequest{
		Path:         "/dbs",
		ResourceType: ResourceDatabases,
		ResourceLink: "",
		Options:      opts,
		Body:         bytes.NewBuffer(body),
	})
}

// CreateDatabase creates a new database with the given ID
func (c *Client) CreateDatabase(ctx context.Context, id string, opts RequestOptions) (*DatabaseResource, *ResponseMetadata, error) {
	body, meta, err := c.CreateDatabaseRaw(ctx, id, opts)
	if err != nil {
		return nil, meta, err
	}
	var db DatabaseResource
	if err = json.Unmarshal(body, &db); err != nil {
		return nil, meta, err
	}
	return &db, meta, err
}

// ListDatabasesRaw lists each database in the CosmosDB Account as raw JSON objects given to the pagination function
func (c *Client) ListDatabasesRaw(ctx context.Context, opts RequestOptions, fn PaginateRawResources) error {
	return c.ListResources(ctx, "Databases", ClientRequest{
		Path:         "/dbs",
		ResourceType: ResourceDatabases,
		ResourceLink: "",
		Options:      opts,
	}, fn)
}

// PaginateDatabaseResource pagination function for a list of DatabaseResources
type PaginateDatabaseResource func(resList []DatabaseResource, meta ResponseMetadata) (bool, error)

// DatabaseClient is a client scoped to a single database
// Used to perform API calls within the scope of the Database resource
type DatabaseClient struct {
	Client     *Client
	DatabaseID string
}

// ListDatabases lists each database in the CosmosDB Account given to the pagination function
func (c *Client) ListDatabases(ctx context.Context, opts RequestOptions, fn PaginateDatabaseResource) error {
	return c.ListDatabasesRaw(ctx, opts, func(resList []json.RawMessage, meta ResponseMetadata) (bool, error) {
		databases := make([]DatabaseResource, len(resList))
		for i, res := range resList {
			var db DatabaseResource
			if err := json.Unmarshal(res, &db); err != nil {
				return false, err
			}
			databases[i] = db
		}
		return fn(databases, meta)
	})
}

// WithDatabase creates a DatabaseClient for the given Database ID
func (c *Client) WithDatabase(id string) *DatabaseClient {
	return &DatabaseClient{
		Client:     c,
		DatabaseID: id,
	}
}

// ResourceLink gets the resource link for the database
func (c *DatabaseClient) ResourceLink() string {
	return fmt.Sprintf("dbs/%s", url.PathEscape(c.DatabaseID))
}

// GetRaw retrieves the raw database resource
func (c *DatabaseClient) GetRaw(ctx context.Context, opts RequestOptions) ([]byte, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	return c.Client.GetResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s", rl),
		ResourceLink: rl,
		ResourceType: ResourceDatabases,
		Options:      opts,
	})
}

// Get retrieves the DatabaseResource
func (c *DatabaseClient) Get(ctx context.Context, opts RequestOptions) (*DatabaseResource, *ResponseMetadata, error) {
	body, meta, err := c.GetRaw(ctx, opts)
	if err != nil {
		return nil, meta, err
	}
	var db DatabaseResource
	if err = json.Unmarshal(body, &db); err != nil {
		return nil, meta, err
	}
	return &db, meta, err
}

// Delete will delete the database
// See Client.DeleteResource for more information
func (c *DatabaseClient) Delete(ctx context.Context, opts RequestOptions) (bool, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	return c.Client.DeleteResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s", rl),
		ResourceLink: rl,
		ResourceType: ResourceDatabases,
		Options:      opts,
	})
}
