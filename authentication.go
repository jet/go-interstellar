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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// TokenVersion is the version of the token that is implemented
	// See https://docs.microsoft.com/en-us/rest/api/documentdb/access-control-on-documentdb-resources
	// for more information
	TokenVersion = "1.0"
	// MasterTokenAuthType specifies that the type of authentication used is 'master' when computing the hash signature for an authenticated REST API call.
	// See https://docs.microsoft.com/en-us/rest/api/cosmos-db/access-control-on-cosmosdb-resources#constructkeytoken for the specification.
	MasterTokenAuthType = "master"
)

// MasterKey is the shared key for the storage account
type MasterKey []byte

// ConnectionString is the connection string for accessing a storage account
type ConnectionString struct {
	Endpoint   string
	AccountKey MasterKey
}

// ParseConnectionString parses a connection string to the storage account
// The format of the connection string is as follows:
//
//     AccountEndpoint=https://accountname.documents.azure.com:443/;AccountKey=BASE64KEY;
//
func ParseConnectionString(connectionString string) (ConnectionString, error) {
	var cs ConnectionString
	for _, cmp := range strings.Split(connectionString, ";") {
		kv := strings.SplitN(cmp, "=", 2)
		switch kv[0] {
		case "AccountEndpoint":
			cs.Endpoint = kv[1]
		case "AccountKey":
			kb, err := ParseMasterKey(kv[1])
			if err != nil {
				return cs, err
			}
			cs.AccountKey = MasterKey(kb)
		}
	}
	return cs, nil
}

// ParseMasterKey parses a base-64 encoded shared access key
func ParseMasterKey(key string) (MasterKey, error) {
	kb, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return MasterKey(kb), nil
}

// HeaderMSDate is the date/time of the request.
// This date time format conforms to the RFC 1123 date time format expressed in Coordinated Universal Time.
// This is passed in the request and much match the date in the hash for the Authorization header
// See: https://docs.microsoft.com/en-us/rest/api/cosmos-db/access-control-on-cosmosdb-resources#constructkeytoken
const HeaderMSDate = "x-ms-date"

// HeaderAuthorization is the header used to pass the Authorization token to the API
const HeaderAuthorization = "Authorization"

// ResourceType indicates the type of resource being requested. Used mostly in Authentication.
type ResourceType string

func (rt ResourceType) String() string {
	return strings.ToLower(string(rt))
}

const (
	// ResourceDatabases is the resource type of a Database
	ResourceDatabases ResourceType = "dbs"
	// ResourceCollections is the resource type of a Collection
	ResourceCollections ResourceType = "colls"
	// ResourceDocuments is the resource type of a Document
	ResourceDocuments ResourceType = "docs"
	// ResourceAttachments is the resource type of an Attachment
	ResourceAttachments ResourceType = "attachments"
	// ResourceStoredProcedures is the resource type of a Stored Procedure (sproc)
	ResourceStoredProcedures ResourceType = "sprocs"
	// ResourceUserDefinedFunctions is the resource type of a User-Defined Function (udf)
	ResourceUserDefinedFunctions ResourceType = "udfs"
	// ResourceTriggers is the resource type of a Trigger
	ResourceTriggers ResourceType = "triggers"
	// ResourceUsers is the resource type of a User
	ResourceUsers ResourceType = "users"
	// ResourcePermissions is the resource type of a Permission
	ResourcePermissions ResourceType = "permissions"
	// ResourceOffers is the resource type of an Offer
	ResourceOffers ResourceType = "offers"
)

// Authorize implements the authorization header for Microsoft Azure Storage services
// See https://docs.microsoft.com/en-us/rest/api/cosmos-db/access-control-on-cosmosdb-resources#constructkeytoken
// for implementation details.
// This implementation assumes the latest version of the API is 2017-04-17
func (k MasterKey) Authorize(r *http.Request, resourceType ResourceType, resourceLink string) (*http.Request, error) {
	if k == nil {
		return r, nil
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	cs := strings.Join([]string{
		strings.ToLower(r.Method),
		resourceType.String(),
		resourceLink, // case sensitive
		strings.ToLower(date),
		"", "",
	}, "\n")
	sig := k.Sign(cs)
	token := url.QueryEscape(fmt.Sprintf("type=%s&ver=%s&sig=%s", MasterTokenAuthType, TokenVersion, sig))
	r.Header.Set(HeaderAuthorization, token)
	if r.Header.Get(HeaderMSAPIVersion) == "" {
		r.Header.Set(HeaderMSAPIVersion, APIVersion)
	}
	r.Header.Set(HeaderMSDate, date)
	return r, nil
}

// Sign will sign the message with HMAC-SHA256
// returns the base-64 encoded result
func (k MasterKey) Sign(message string) string {
	h := hmac.New(sha256.New, []byte(k))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
