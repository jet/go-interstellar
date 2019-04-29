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
	"strconv"
)

// HeaderIndexingDirective is used to enable or disable indexing on the resource.
const HeaderIndexingDirective = "x-ms-indexing-directive"

// HeaderDocDBIsUpsert is set to true if the document should be created if it does not exist, or updated in-place if it does.
const HeaderDocDBIsUpsert = "x-ms-documentdb-is-upsert"

// DocumentClient is a client scoped to a single document
// Used to perform API calls within the scope of a single Document resource
type DocumentClient struct {
	Client       *Client
	DatabaseID   string
	CollectionID string
	DocumentID   string
	PartitionKey []string
}

// WithDocument creates a DocumentClient for the given Document ID and PartitionKey within this Collection
func (c *CollectionClient) WithDocument(id string, partitionKey []string) *DocumentClient {
	return &DocumentClient{
		Client:       c.Client,
		DatabaseID:   c.DatabaseID,
		CollectionID: c.CollectionID,
		DocumentID:   id,
		PartitionKey: partitionKey,
	}
}

// ResourceLink gets the resource link for the document
func (c *DocumentClient) ResourceLink() string {
	return fmt.Sprintf("dbs/%s/colls/%s/docs/%s", url.PathEscape(c.DatabaseID), url.PathEscape(c.CollectionID), url.PathEscape(c.DocumentID))
}

func (c *DocumentClient) addPartitionKey(opts RequestOptions) RequestOptions {
	if len(c.PartitionKey) == 0 {
		return opts
	}
	fn := RequestOptionsFunc(func(req *http.Request) {
		b, _ := json.Marshal(c.PartitionKey)
		req.Header.Set(HeaderDocDBPartitionKey, string(b))
	})
	if opts == nil {
		return fn
	}
	return RequestOptionsList{opts, fn}
}

// CreateDocumentRequest are parameters for CreateDocument
type CreateDocumentRequest struct {
	// Partition Key for partitioned collections
	PartitionKey []string

	// Upsert indicates if the request should replace the existing document
	Upsert bool

	// IndexingDirective determines if the document will be indexed
	IndexingDirective *DocumentIndexingDirective

	// Document is the replacement document. This will be marshalled into JSON
	// Either this or Document must be non-nil.
	Document interface{}

	// Body is the document body as JSON bytes. Either this or Document must be non-nil.
	Body []byte

	// Options are any additional request options to add to the request
	Options RequestOptions

	// Unmarshaler is an optional Unmarshaler that will be called with the response body
	Unmarshaler json.Unmarshaler
}

func (r CreateDocumentRequest) json() ([]byte, error) {
	if r.Body == nil && r.Document == nil {
		return nil, Error("interstellar: must set either a Document or a Body for CreateDocumentRequest")
	}
	if len(r.Body) == 0 {
		return json.Marshal(r.Document)
	}
	return r.Body, nil
}

// ApplyOptions applies the request options to the api request
func (r CreateDocumentRequest) ApplyOptions(req *http.Request) {
	if r.Upsert {
		req.Header.Set(HeaderDocDBIsUpsert, strconv.FormatBool(r.Upsert))
	}
	if len(r.PartitionKey) > 0 {
		pkey, _ := json.Marshal(r.PartitionKey)
		req.Header.Set(HeaderDocDBPartitionKey, string(pkey))
	}
	if r.IndexingDirective != nil {
		req.Header.Set(HeaderIndexingDirective, string(*r.IndexingDirective))
	}
	if r.Options != nil {
		r.Options.ApplyOptions(req)
	}
}

// CreateDocument creates or updates a document in the collection
func (c *CollectionClient) CreateDocument(ctx context.Context, req CreateDocumentRequest) ([]byte, *ResponseMetadata, error) {
	body, err := req.json()
	if err != nil {
		return nil, nil, err
	}
	rl := c.ResourceLink()
	data, meta, err := c.Client.CreateOrReplaceResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s/docs", rl),
		ResourceLink: rl,
		ResourceType: ResourceDocuments,
		Body:         bytes.NewBuffer(body),
		Options:      req,
	})
	if err != nil {
		return nil, meta, err
	}
	if req.Unmarshaler != nil {
		if err = req.Unmarshaler.UnmarshalJSON(data); err != nil {
			return nil, meta, err
		}
	}
	return data, meta, nil
}

// ListDocumentsRaw lists each document in the collection as raw JSON objects
func (c *CollectionClient) ListDocumentsRaw(ctx context.Context, opts RequestOptions, fn PaginateRawResources) error {
	rl := c.ResourceLink()
	return c.Client.ListResources(ctx, "Documents", ClientRequest{
		Path:         fmt.Sprintf("/%s/docs", rl),
		ResourceLink: rl,
		ResourceType: ResourceDocuments,
		Options:      opts,
	}, fn)
}

// QueryDocumentsRaw posts the query to the collection and paginates through the results using the supplied paginate function
func (c *CollectionClient) QueryDocumentsRaw(ctx context.Context, query *Query, fn PaginateRawResources) error {
	if query == nil {
		return Error("interstellar: query cannot be nil")
	}
	rl := fmt.Sprintf("dbs/%s/colls/%s", url.PathEscape(c.DatabaseID), url.PathEscape(c.CollectionID))
	qjson, err := json.Marshal(&query)
	if err != nil {
		return err
	}
	return c.Client.ListResources(ctx, "Documents", ClientRequest{
		Method:       http.MethodPost,
		Path:         fmt.Sprintf("/%s/docs", rl),
		ResourceLink: rl,
		ResourceType: ResourceDocuments,
		Options:      query,
		Body:         bytes.NewBuffer(qjson),
	}, fn)
}

// GetRaw retrieves the raw document
func (c *DocumentClient) GetRaw(ctx context.Context, opts RequestOptions) ([]byte, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	return c.Client.GetResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s", rl),
		ResourceLink: rl,
		ResourceType: ResourceDocuments,
		Options:      c.addPartitionKey(opts),
	})
}

// Get retrieves the raw document and unmarshalls the content into the given value
func (c *DocumentClient) Get(ctx context.Context, opts RequestOptions, v interface{}) (*ResponseMetadata, error) {
	body, meta, err := c.GetRaw(ctx, opts)
	if err != nil {
		return meta, err
	}
	if err = json.Unmarshal(body, v); err != nil {
		return meta, err
	}
	return meta, nil
}

// Delete removes the document from the collection
func (c *DocumentClient) Delete(ctx context.Context, opts RequestOptions) (bool, *ResponseMetadata, error) {
	rl := c.ResourceLink()
	return c.Client.DeleteResource(ctx, ClientRequest{
		Path:         fmt.Sprintf("/%s", rl),
		ResourceLink: rl,
		ResourceType: ResourceDocuments,
		Options:      c.addPartitionKey(opts),
	})
}

// ReplaceDocumentRequest are parameters for CreateDocument
type ReplaceDocumentRequest struct {
	// ETag is used for optimistic concurrency. If set, the ETag value of the existing document must match this in order for the operation to complete.
	ETag string

	// IndexingDirective determines if the document will be indexed
	IndexingDirective *DocumentIndexingDirective

	// Document is the replacement document. This will be marshalled into JSON
	// Either this or Document must be non-nil.
	Document interface{}

	// Body is the document body as JSON bytes. Either this or Document must be non-nil.
	Body []byte

	// Options are any additional request options to add to the request
	Options RequestOptions

	// Unmarshaler is an optional Unmarshaler that will be called with the response body
	Unmarshaler json.Unmarshaler
}

func (r ReplaceDocumentRequest) json() ([]byte, error) {
	if r.Body == nil && r.Document == nil {
		return nil, Error("interstellar: must set either a Document or a Body for ReplaceDocumentRequest")
	}
	if len(r.Body) == 0 {
		return json.Marshal(r.Document)
	}
	return r.Body, nil
}

// ApplyOptions applies the request options to the api request
func (r ReplaceDocumentRequest) ApplyOptions(req *http.Request) {
	if r.ETag != "" {
		req.Header.Set(HeaderIfMatch, r.ETag)
	}
	if r.IndexingDirective != nil {
		req.Header.Set(HeaderIndexingDirective, string(*r.IndexingDirective))
	}
	if r.Options != nil {
		r.Options.ApplyOptions(req)
	}
}

// ReplaceDocument replaces this document
func (c *DocumentClient) ReplaceDocument(ctx context.Context, req ReplaceDocumentRequest) ([]byte, *ResponseMetadata, error) {
	body, err := req.json()
	if err != nil {
		return nil, nil, err
	}
	rl := c.ResourceLink()
	data, meta, err := c.Client.CreateOrReplaceResource(ctx, ClientRequest{
		Method:       http.MethodPut,
		Path:         fmt.Sprintf("/%s", rl),
		ResourceLink: rl,
		ResourceType: ResourceDocuments,
		Options:      c.addPartitionKey(req),
		Body:         bytes.NewBuffer(body),
	})
	if err != nil {
		return nil, meta, err
	}
	if req.Unmarshaler != nil {
		if err = req.Unmarshaler.UnmarshalJSON(data); err != nil {
			return nil, meta, err
		}
	}
	return data, meta, nil
}
