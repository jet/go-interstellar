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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// ContentTypeJSON is the MIME-Type for generic JSON content
	ContentTypeJSON = "application/json"
	// ContentTypeQueryJSON is the MIME-Type for CosmosDB Queries (json)
	ContentTypeQueryJSON = "application/query+json"
)

// Common Request Headers
// https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
const (
	// HeaderUserAgent used to differentiate different client applications.
	// Recommended format is {user agent name}/{version}
	// For example: MyApplication/1.0.0
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderUserAgent = "User-Agent"
	// HeaderContentType is used mostly for POST query operations
	// where the Content-Type header must be application/query+json
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderContentType = "Content-Type"
	// HeaderIfMatch used for optimistic concurrency based off ETag
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderIfMatch = "If-Match"
	// HeaderIfNoneMatch used for optimistic concurrency based off ETag
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderIfNoneMatch = "If-None-Match"
	// HeaderIfModifiedSince used for optimistic concurrency based off Modified Date (RFC1123 Date/Time Format)
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderIfModifiedSince = "If-Modified-Since"
	// HeaderActivityID "x-ms-activity-id"
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderActivityID = "x-ms-activity-id"
	// HeaderConsistencyLevel  Strong, Bounded, Session, or Eventual (in order of strongest to weakest)
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderConsistencyLevel = "x-ms-consistency-level"
	// HeaderMaxItemCount is supplied in list/query operations to limit the number of results per page.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderMaxItemCount = "x-ms-max-item-count"
	// HeaderDocDBPartitionKey the partition key value for the requested document or attachment.
	// The format of this header is a JSON array of values
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderDocDBPartitionKey = "x-ms-documentdb-partitionkey"
	// HeaderDocDBQueryEnableCrossPartition is set to true for queries which should span multiple partitions, and a partition key is not supplied.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderDocDBQueryEnableCrossPartition = "x-ms-documentdb-query-enablecrosspartition"
	// HeaderMSAPIVersion is used to specify which version of the REST API is being used by the request
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderMSAPIVersion = "x-ms-version"
	// HeaderAIM is sued to indicate the Change Feed request. It must be set to "incremental feed" or otherwise omitted.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
	HeaderAIM = "A-IM"

	// HeaderDocDBPartitionKeyRangeID Used in change feed requests. This is a number which is the Parittion Key Range ID used for reading data.
	HeaderDocDBPartitionKeyRangeID = "x-ms-documentdb-partitionkeyrangeid"
)

// HeaderDocDBIsQuery is used to indicate the POST request is a query, not a Create. Must be set to "true".
const HeaderDocDBIsQuery = "x-ms-documentdb-isquery"

// Common Response Headers
// https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
const (
	// HeaderDate is the date time of the response. This date time format conforms to the RFC 1123 date time format expressed in Coordinated Universal Time.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderDate = "Date"
	// HeaderETag indicates the etage of the resource. the same as the `_etag` property on the resource body.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderETag = "ETag"
	// HeaderAltContentPath is the alternate content path of the resource.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderAltContentPath = "x-ms-alt-content-path"
	// HeaderContinuation represents the intermediate state of query (or read-feed) execution, and is returned when there are additional results aside from what was returned in the response. Clients can resubmitted the request with a request header containing the value of x-ms-continuation.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderContinuation = "x-ms-continuation"
	// HeaderItemCount is the number of items returned for a query or read-feed request.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderItemCount = "x-ms-item-count"
	// HeaderRequestCharge is the number of normalized requests a.k.a. request units (RU) for the operation. For more information, see Request units in Azure Cosmos DB.
	// See: https://docs.microsoft.com/en-us/azure/cosmos-db/request-units/
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderRequestCharge = "x-ms-request-charge"
	// HeaderResourceQuota is the allotted quota for a resource in an account.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderResourceQuota = "x-ms-resource-quota"
	// HeaderResourceUsage is the current usage count of a resource in an account. When deleting a resource, this shows the number of resources after the deletion.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderResourceUsage = "x-ms-resource-usage"
	// HeaderRetryAfterMS is the number of milliseconds to wait to retry the operation after an initial operation received HTTP status code 429 and was throttled.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderRetryAfterMS = "x-ms-retry-after-ms"
	// HeaderSchemaVersion Shows the resource schema version number.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderSchemaVersion = "x-ms-schemaversion"
	// HeaderServiceVersion is the service version number.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderServiceVersion = "x-ms-serviceversion"
	// HeaderSessionToken is the session token of the request.
	// For session consistency, clients must echo this request via the x-ms-session-token request header for subsequent operations made to the corresponding collection.
	// See: https://docs.microsoft.com/azure/cosmos-db/consistency-levels
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
	HeaderSessionToken = "x-ms-session-token"
)

// ConsistencyLevel specifies the consistency level of the operation
// See https://docs.microsoft.com/azure/cosmos-db/consistency-levels for more information
type ConsistencyLevel string

const (
	// ConsistencyStrong denotes the operation should be strongly consistent. (Limited Linearizability guarantee)
	ConsistencyStrong = ConsistencyLevel("Strong")
	// ConsistencyBounded denotes a slightly weaker consistency than strong. Reads may lag behind writes by a constant number or time frame.
	ConsistencyBounded = ConsistencyLevel("Bounded")
	// ConsistencySession is consistency scoped to a single client/session. Read your writes, monotonic reads, monotonic writes.
	ConsistencySession = ConsistencyLevel("Session")
	// ConsistencyEventual is the weakest consistency level.
	ConsistencyEventual = ConsistencyLevel("Eventual")
)

// ClientRequest encapsulates the CosmosDB API request parameters
type ClientRequest struct {
	// Method is the HTTP Method/Verb used for the request
	Method string
	// Path is the URL path portion of the request
	Path string
	// ResourceType is the type of resource being requested, such as "dbs" or "colls"
	ResourceType ResourceType
	// ResourceLink is the identity property of the resource that the request is directed at.
	// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/access-control-on-cosmosdb-resources
	// For example, the format for a collection is: "dbs/{db.id}/colls/{coll.id}" like "dbs/MyDatabase/colls/MyCollection".
	// The resource links is CASE SENSITIVE.
	// Note: It is NOT the same as the '_self' property of a resource (in most cases).
	ResourceLink string
	// Options allow for applying additional headers and other request options to the HTTP request
	Options RequestOptions
	// Body is a reader which should be sent as the body of the request
	Body io.Reader
	// GetBody is used to set the body of the request for retrys and resubmissions
	GetBody func() (io.ReadCloser, error)
}

func (req *ClientRequest) readEntireBody() ([]byte, error) {
	if req.Body != nil {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = bytes.NewBuffer(data)
		return data, nil
	}
	if req.GetBody != nil {
		rc, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return ioutil.ReadAll(rc)
	}
	return nil, nil
}

func (req *ClientRequest) reusableBody() error {
	if req.Body == nil && req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return err
		}
		defer body.Close()
		data, err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		req.Body = bytes.NewBuffer(data)
	} else if req.Body != nil && req.GetBody == nil {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}
		req.Body = bytes.NewBuffer(data)
		req.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewBuffer(data)), nil
		}
	}
	return nil
}

// NewHTTPRequest creates a new authorized http request that can be run against a Requester such as *http.Client or anything implementing the Requester interface
func (c *Client) NewHTTPRequest(ctx context.Context, req ClientRequest) (*http.Request, error) {
	if err := (&req).reusableBody(); err != nil {
		return nil, err
	}
	hreq, err := http.NewRequest(req.Method, fmt.Sprintf("%s/%s", strings.TrimSuffix(c.Endpoint, "/"), strings.TrimPrefix(req.Path, "/")), req.Body)
	if err != nil {
		return nil, err
	}
	hreq.GetBody = req.GetBody
	if c.UserAgent != "" {
		hreq.Header.Set(HeaderUserAgent, c.UserAgent)
	} else {
		hreq.Header.Set(HeaderUserAgent, DefaultUserAgent)
	}
	if ctx != nil {
		hreq = hreq.WithContext(ctx)
	}
	if req.Options != nil {
		req.Options.ApplyOptions(hreq)
	}
	hreq, err = c.Authorizer.Authorize(hreq, req.ResourceType, req.ResourceLink)
	return hreq, err
}

// RequestOptions augments the request, such as adding headers, or query parameter to an existing http.Request
type RequestOptions interface {
	ApplyOptions(req *http.Request)
}

// RequestOptionsFunc implements RequestOptions for a pure function
// Can be used to apply options with an anonymous function such as RequestOptionsFunc(func(req *http.Request) { ... })
type RequestOptionsFunc func(req *http.Request)

// ApplyOptions implementation for RequestOptions interface
func (fn RequestOptionsFunc) ApplyOptions(req *http.Request) {
	fn(req)
}

// RequestOptionsList implements RequestOptions for a list/slice of RequestOptionsListRequestOptionsList
type RequestOptionsList []RequestOptions

// ApplyOptions implementation for RequestOptions interface
func (l RequestOptionsList) ApplyOptions(req *http.Request) {
	for _, opt := range l {
		if opt != nil {
			opt.ApplyOptions(req)
		}
	}
}

// CommonRequestOptions is a helper which adds additional options to their appropriate headers in the CosmosDB HTTP request
// The specific options which are permitted varies depending on the request
// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-request-headers
type CommonRequestOptions struct {
	ActivityID                          string
	ContentType                         string
	IfMatch                             string
	IfNoneMatch                         string
	IfModifiedSince                     time.Time
	SessionToken                        string
	ConsistencytLevel                   ConsistencyLevel
	DocumentDBPartitionKey              string
	DocumentDBPartitionKeyRangeID       string
	DocumentDBQueryEnableCrossPartition bool
	ChangeFeed                          bool
	MaxItemCount                        int
	Continuation                        string
}

// ApplyOptions sets the common headers defined in the CommonRequestOptions struct on the given http request object
func (o *CommonRequestOptions) ApplyOptions(req *http.Request) {
	if o == nil {
		return
	}
	if req.Method == http.MethodPut || req.Method == http.MethodPost {
		if o.ContentType != "" {
			req.Header.Set(HeaderContentType, o.ContentType)
		}
	}
	if req.Method == http.MethodPut || req.Method == http.MethodDelete {
		if o.IfMatch != "" {
			req.Header.Set(HeaderIfMatch, o.IfMatch)
		}
	}
	if req.Method == http.MethodGet {
		if o.IfNoneMatch != "" {
			req.Header.Set(HeaderIfNoneMatch, o.IfNoneMatch)
		} else if !o.IfModifiedSince.IsZero() {
			req.Header.Set(HeaderIfModifiedSince, o.IfModifiedSince.Format(http.TimeFormat))
		}
	}
	if o.DocumentDBQueryEnableCrossPartition {
		req.Header.Set(HeaderDocDBQueryEnableCrossPartition, "true")
	}
	if o.ActivityID != "" {
		req.Header.Set(HeaderActivityID, o.ActivityID)
	}
	if o.SessionToken != "" {
		req.Header.Set(HeaderSessionToken, o.SessionToken)
	}
	if o.ConsistencytLevel != "" {
		req.Header.Set(HeaderConsistencyLevel, string(o.ConsistencytLevel))
	}
	if o.Continuation != "" {
		req.Header.Set(HeaderContinuation, o.Continuation)
	}
	if o.MaxItemCount != 0 {
		req.Header.Set(HeaderMaxItemCount, fmt.Sprintf("%d", o.MaxItemCount))
	}
	if o.DocumentDBPartitionKey != "" {
		req.Header.Set(HeaderDocDBPartitionKey, o.DocumentDBPartitionKey)
	}
	if o.DocumentDBPartitionKeyRangeID != "" {
		req.Header.Set(HeaderDocDBPartitionKeyRangeID, o.DocumentDBPartitionKeyRangeID)
	}
	if o.ChangeFeed {
		req.Header.Set(HeaderAIM, "Incremental feed")
	}
}

// ResponseMetadata is the parsed header values from the response
// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/common-cosmosdb-rest-response-headers
type ResponseMetadata struct {
	Date           time.Time
	ETag           string
	ActivityID     string
	AltContentPath string
	Continuation   string
	RequestCharge  string
	ResourceQuota  string
	RetryAfterMS   time.Duration
	ItemCount      int64
	ResourceUsage  string
	SchemaVersion  string
	ServiceVersion string
	SessionToken   string
}

// GetResponseMetadata extracts response metadata from the http headers
// And parses them into native types where applicable (such as time or numbers)
func GetResponseMetadata(resp *http.Response) (m ResponseMetadata) {
	if resp == nil || resp.Header == nil {
		return
	}
	hdr := resp.Header
	if dhdr := hdr.Get(HeaderDate); dhdr != "" {
		if date, err := time.Parse(time.RFC1123, dhdr); err == nil {
			m.Date = date
		}
	}
	m.ETag = hdr.Get(HeaderETag)
	m.ActivityID = hdr.Get(HeaderActivityID)
	m.AltContentPath = hdr.Get(HeaderAltContentPath)
	m.Continuation = hdr.Get(HeaderContinuation)
	m.RequestCharge = hdr.Get(HeaderRequestCharge)
	m.ResourceQuota = hdr.Get(HeaderResourceQuota)
	m.ResourceUsage = hdr.Get(HeaderResourceUsage)
	m.SchemaVersion = hdr.Get(HeaderSchemaVersion)
	m.ServiceVersion = hdr.Get(HeaderServiceVersion)
	m.SessionToken = hdr.Get(HeaderSessionToken)
	if hv := hdr.Get(HeaderItemCount); hv != "" {
		i, err := strconv.ParseInt(hv, 10, 64)
		if err == nil {
			m.ItemCount = i
		}
	}
	return
}
