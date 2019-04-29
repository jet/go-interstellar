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
	"encoding/json"
	"time"
)

// OfferResource represents a performance-level offering on a resource
type OfferResource struct {
	ID              string
	ResourceID      string
	Timestamp       time.Time
	Self            string
	ETag            string
	OfferVersion    OfferVersion
	OfferType       OfferType
	Content         *OfferContent
	Resource        string
	OfferResourceID string
}

type offerJSON struct {
	ID              string          `json:"id"`
	ResourceID      string          `json:"_rid"`
	Timestamp       int64           `json:"_ts"`
	Self            string          `json:"_self"`
	ETag            string          `json:"_etag"`
	OfferVersion    string          `json:"offerVersion,omitempty"`
	OfferType       string          `json:"offerType"`
	Content         json.RawMessage `json:"content,omitempty"`
	Resource        string          `json:"resource"`
	OfferResourceID string          `json:"offerResourceId"`
}

// OfferContentV2 is the content of the OfferVersion "V2" for user-defined throughput
type OfferContentV2 struct {
	OfferThroughput int `json:"offerThroughput"`

	// RUPMEnabled is Request Units(RU)/Minute throughput is enabled/disabled for collection in the Azure Cosmos DB service.
	RUPMEnabled *bool `json:"offerIsRUPerMinuteThroughputEnabled,omitempty"`
}

// OfferVersion differentiates different offer schemas
type OfferVersion string

const (
	// OfferV1 uses a set list of OfferTypes
	OfferV1 OfferVersion = "V1"
	// OfferV2 allows tunable request unit throughput
	OfferV2 OfferVersion = "V2"
)

// OfferType is used to specify the pre-defined performance levels for the CosmosDB container
type OfferType string

const (
	// OfferTypeInvalid indicates the performance level is user-defined
	OfferTypeInvalid = OfferType("Invalid")
	// OfferTypeS1 maps to the S1 performance level
	OfferTypeS1 = OfferType("S1")
	// OfferTypeS2 maps to the S3 performance level
	OfferTypeS2 = OfferType("S2")
	// OfferTypeS3 maps to the S3 performance level
	OfferTypeS3 = OfferType("S3")
)

// OfferContent encapsulates the different offer schemas depending on OfferVersion
type OfferContent struct {
	// V2 is the OfferContentV2 version of the schema for user-defined performance parameters
	V2 *OfferContentV2
}

// ErrOfferInvalidVersion is returned when an invalid offer version is read
const ErrOfferInvalidVersion = Error("insterstellar: invalid offer version")

// UnmarshalJSON implementes json.Unmarshaler for OfferResource
func (oc *OfferResource) UnmarshalJSON(data []byte) error {
	var offerjs offerJSON
	if err := json.Unmarshal(data, &offerjs); err != nil {
		return err
	}
	oc.ID = offerjs.ID
	oc.ResourceID = offerjs.ResourceID
	oc.Self = offerjs.Self
	oc.ETag = offerjs.ETag
	oc.OfferVersion = OfferVersion(offerjs.OfferVersion)
	oc.OfferType = OfferType(offerjs.OfferType)
	oc.Timestamp = time.Unix(offerjs.Timestamp, 0)
	oc.Resource = offerjs.Resource
	oc.OfferResourceID = offerjs.OfferResourceID
	if oc.OfferVersion == OfferV2 {
		var content OfferContent
		if err := json.Unmarshal(offerjs.Content, &content.V2); err != nil {
			return err
		}
		oc.Content = &content
	}
	return nil
}

// MarshalJSON implementes json.Marshaler for OfferResource
func (oc *OfferResource) MarshalJSON() ([]byte, error) {
	var offerjs offerJSON
	offerjs.ID = oc.ID
	offerjs.ResourceID = oc.ResourceID
	offerjs.Self = oc.Self
	offerjs.ETag = oc.ETag
	offerjs.Timestamp = oc.Timestamp.Unix()
	offerjs.OfferVersion = string(oc.OfferVersion)
	offerjs.OfferType = string(oc.OfferType)
	offerjs.Resource = oc.Resource
	offerjs.OfferResourceID = oc.OfferResourceID
	if oc.OfferVersion == OfferV2 {
		oc.OfferType = OfferTypeInvalid
		content, err := json.Marshal(oc.Content.V2)
		if err != nil {
			return nil, err
		}
		offerjs.Content = json.RawMessage(content)
	}
	return json.Marshal(&offerjs)
}
