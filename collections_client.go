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

const (
	// HeaderOfferType is used to set the Offer type on the Collection at creation time. This is one of the pre-defined levels: S1,S2,S3.
	HeaderOfferType = "x-ms-offer-type"
	// HeaderOfferThroughput is used to set the provisioned RU Throughput on the collection at creation time.
	HeaderOfferThroughput = "x-ms-offer-throughput"
)

// CollectionClient is a client scoped to a single collection
// Used to perform API calls within the scope of the Collection resource
type CollectionClient struct {
	Client       *Client
	DatabaseID   string
	CollectionID string
}

// WithCollection creates a CollectionClient for the given Collection within this Database
func (c *DatabaseClient) WithCollection(id string) *CollectionClient {
	return &CollectionClient{
		Client:       c.Client,
		DatabaseID:   c.DatabaseID,
		CollectionID: id,
	}
}

// ResourceLink gets the resource link for the collection
func (c *CollectionClient) ResourceLink() string {
	return fmt.Sprintf("dbs/%s/colls/%s", url.PathEscape(c.DatabaseID), url.PathEscape(c.CollectionID))
}

// ListCollectionsRaw lists each collection in the database as raw JSON objects
func (c *DatabaseClient) ListCollectionsRaw(ctx context.Context, opts RequestOptions, fn PaginateRawResources) error {
	rl := c.ResourceLink()
	return c.Client.ListResources(ctx, "DocumentCollections", ClientRequest{
		Path:         fmt.Sprintf("/%s/colls", rl),
		ResourceLink: rl,
		ResourceType: ResourceCollections,
		Options:      opts,
	}, fn)
}

// PaginateCollectionResource pagination function for a list of CollectionResource
type PaginateCollectionResource func(resList []CollectionResource, meta ResponseMetadata) (bool, error)

// ListCollections lists each collection in the database
func (c *DatabaseClient) ListCollections(ctx context.Context, opts RequestOptions, fn PaginateCollectionResource) error {
	return c.ListCollectionsRaw(ctx, opts, func(resList []json.RawMessage, meta ResponseMetadata) (bool, error) {
		collections := make([]CollectionResource, len(resList))
		for i, res := range resList {
			var db CollectionResource
			if err := json.Unmarshal(res, &db); err != nil {
				return false, err
			}
			collections[i] = db
		}
		return fn(collections, meta)
	})
}

// CreateCollectionRequest captures the request options for creating a new Collection
type CreateCollectionRequest struct {
	OfferThroughput int                       `json:"-"`
	OfferType       OfferType                 `json:"-"`
	Options         RequestOptions            `json:"-"`
	ID              string                    `json:"id"`
	IndexingPolicy  *CollectionIndexingPolicy `json:"indexingPolicy,omitempty"`
	PartitionKey    *CollectionPartitionKey   `json:"partitionKey,omitempty"`
}

// ApplyOptions applies additional headers necessary to complete a CreateCollection request
func (c CreateCollectionRequest) ApplyOptions(req *http.Request) {
	if c.OfferThroughput != 0 {
		req.Header.Set(HeaderOfferThroughput, fmt.Sprintf("%d", c.OfferThroughput))
	} else if c.OfferType != "" {
		req.Header.Set(HeaderOfferType, string(c.OfferType))
	}
	if c.Options != nil {
		c.Options.ApplyOptions(req)
	}
}

// CreateCollectionRaw creates a new collection and returns the raw response
func (c *DatabaseClient) CreateCollectionRaw(ctx context.Context, req CreateCollectionRequest) ([]byte, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	body, err := json.Marshal(&req)
	if err != nil {
		return nil, nil, err
	}
	return c.Client.CreateOrReplaceResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s/colls", rl),
		ResourceType: ResourceCollections,
		ResourceLink: rl,
		Options:      req,
		Body:         bytes.NewBuffer(body),
	})
}

// CreateCollection creates a new collection
func (c *DatabaseClient) CreateCollection(ctx context.Context, req CreateCollectionRequest) (*CollectionResource, *ResponseMetadata, error) {
	body, meta, err := c.CreateCollectionRaw(ctx, req)
	if err != nil {
		return nil, meta, err
	}
	var coll CollectionResource
	if err = json.Unmarshal(body, &coll); err != nil {
		return nil, meta, err
	}
	return &coll, meta, err
}

// GetRaw retrieves the raw collection
func (c *CollectionClient) GetRaw(ctx context.Context, opts RequestOptions) ([]byte, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	return c.Client.GetResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s", rl),
		ResourceLink: rl,
		ResourceType: ResourceCollections,
		Options:      opts,
	})
}

// Get retrieves the CollectionResource
func (c *CollectionClient) Get(ctx context.Context, opts RequestOptions) (*CollectionResource, *ResponseMetadata, error) {
	body, meta, err := c.GetRaw(ctx, opts)
	if err != nil {
		return nil, meta, err
	}
	var coll CollectionResource
	if err = json.Unmarshal(body, &coll); err != nil {
		return nil, meta, err
	}
	return &coll, meta, err
}

// Delete will delete the collection
// See Client.DeleteResource for more information
func (c *CollectionClient) Delete(ctx context.Context, opts RequestOptions) (bool, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	return c.Client.DeleteResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s", rl),
		ResourceLink: rl,
		ResourceType: ResourceCollections,
		Options:      opts,
	})
}
