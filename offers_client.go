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
	"strings"
)

// ListOffersRaw lists each offer in the CosmosDB account as raw JSON objects
func (c *Client) ListOffersRaw(ctx context.Context, opts RequestOptions, fn PaginateRawResources) error {
	return c.ListResources(ctx, "Offers", ClientRequest{
		Path:         "/offers",
		ResourceLink: "",
		ResourceType: ResourceOffers,
		Options:      opts,
	}, fn)
}

// QueryOffersRaw executes the given OfferQuery and paginates through the offers
func (c *Client) QueryOffersRaw(ctx context.Context, query *Query, fn PaginateRawResources) error {
	if query == nil {
		return Error("interstellar: query cannot be nil")
	}
	qjson, err := json.Marshal(query)
	if err != nil {
		return err
	}
	return c.ListResources(ctx, "Offers", ClientRequest{
		Method:       http.MethodPost,
		Path:         "/offers",
		ResourceLink: "",
		ResourceType: ResourceOffers,
		Options:      query,
		Body:         bytes.NewBuffer(qjson),
	}, fn)
}

// PaginateOfferResource pagination function for a list of OfferResource
type PaginateOfferResource func(resList []OfferResource, meta ResponseMetadata) (bool, error)

// ListOffers lists each collection in the CosmosDB account
func (c *Client) ListOffers(ctx context.Context, opts RequestOptions, fn PaginateOfferResource) error {
	return c.ListOffersRaw(ctx, opts, func(resList []json.RawMessage, meta ResponseMetadata) (bool, error) {
		collections := make([]OfferResource, len(resList))
		for i, res := range resList {
			var db OfferResource
			if err := json.Unmarshal(res, &db); err != nil {
				return false, err
			}
			collections[i] = db
		}
		return fn(collections, meta)
	})
}

// OfferClient is a client scoped to a single offer
// Used to perform API calls within the scope of the Offer resource
type OfferClient struct {
	Client  *Client
	OfferID string
}

// WithOffer creates a OfferClient for the given Offer within this account
func (c *Client) WithOffer(id string) *OfferClient {
	return &OfferClient{
		Client:  c,
		OfferID: id,
	}
}

// GetRaw retrieves the raw offer JSON
func (c *OfferClient) GetRaw(ctx context.Context, opts RequestOptions) ([]byte, *ResponseMetadata, error) {
	return c.Client.GetResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/offers/%s", c.OfferID),
		ResourceLink: strings.ToLower(c.OfferID),
		ResourceType: ResourceOffers,
		Options:      opts,
	})
}

// Get retrieves the OfferResource
func (c *OfferClient) Get(ctx context.Context, opts RequestOptions) (*OfferResource, *ResponseMetadata, error) {
	body, meta, err := c.GetRaw(ctx, opts)
	if err != nil {
		return nil, meta, err
	}
	var offer OfferResource
	if err = json.Unmarshal(body, &offer); err != nil {
		return nil, meta, err
	}
	return &offer, meta, nil
}

// ReplaceOfferRequest encapsulates the offer to replace
type ReplaceOfferRequest struct {
	Offer   *OfferResource
	Options RequestOptions
}

// ApplyOptions applies the request options to the api request
func (r ReplaceOfferRequest) ApplyOptions(req *http.Request) {
	if r.Options != nil {
		r.Options.ApplyOptions(req)
	}
}

// ReplaceOffer replaces an existing offer with new parameters
func (c *Client) ReplaceOffer(ctx context.Context, req ReplaceOfferRequest) (*OfferResource, *ResponseMetadata, error) {
	rl := fmt.Sprintf("offers/%s", url.PathEscape(req.Offer.ResourceID))
	body, err := req.Offer.MarshalJSON()
	if err != nil {
		return nil, nil, err
	}
	resp, meta, err := c.CreateOrReplaceResource(ctx, ClientRequest{
		Method:       http.MethodPut,
		Path:         fmt.Sprintf("/%s", rl),
		ResourceLink: strings.ToLower(req.Offer.ResourceID),
		ResourceType: ResourceOffers,
		Body:         bytes.NewBuffer(body),
		Options:      req,
	})
	if err != nil {
		return nil, meta, err
	}
	var result OfferResource
	if err = (&result).UnmarshalJSON(resp); err != nil {
		return nil, meta, err
	}
	return &result, meta, err
}
