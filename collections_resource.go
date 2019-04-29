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

// CollectionResource represents a Collection container in Cosmos DB
// Documentation adapted from adapted from docs.microsoft.com
// See https://docs.microsoft.com/en-us/rest/api/cosmos-db/collections for the latest documentation
type CollectionResource struct {
	// ID is the unique user generated name for the collection.
	ID string `json:"id"`
	// ResourceID is a unique identifier that is also hierarchical per the resource stack on the resource model. It is used internally for placement of and navigation to the collection resource.
	ResourceID string `json:"_rid,omitempty"`
	// Timestamp is a system generated property. It denotes the last updated timestamp of the resource.
	Timestamp int64 `json:"_ts,omitempty"`
	// Self is the unique addressable URI for the resource.
	Self string `json:"_self,omitempty"`
	// ETag value required for optimistic concurrency control.
	ETag string `json:"_etag,omitempty"`
	// Docs specifies the addressable path of the documents resource.
	Docs string `json:"_docs,omitempty"`
	// Sprocs specifies the addressable path of the stored procedures (sprocs) resource.
	Sprocs string `json:"_sprocs,omitempty"`
	// Triggers specifies the addressable path of the triggers resource.
	Triggers string `json:"_triggers,omitempty"`
	// UDFs specifies the addressable path of the user-defined functions (udfs) resource.
	UDFs string `json:"_udfs,omitempty"`
	// Conflicts specifies the addressable path of the conflicts resource.
	// During an operation on a resource within a collection, if a conflict occurs, users can inspect the conflicting resources by performing a GET on the conflicts URI path.
	//
	Conflicts string `json:"_conflicts,omitempty"`
	// IndexingPolicy settings for collection.
	IndexingPolicy *CollectionIndexingPolicy `json:"indexingPolicy,omitempty"`
	// PartitionKey is the partitioning configuration settings for collection.
	PartitionKey *CollectionPartitionKey `json:"partitionKey,omitempty"`
}

// CollectionIndexingPolicy represents the indexing policy configuration for a Collection
type CollectionIndexingPolicy struct {
	// Automatic indicates whether automatic indexing is on or off.
	// The default value is True, thus all documents are indexed.
	// Setting the value to False would allow manual configuration of indexing paths.
	Automatic *bool `json:"automatic,omitempty"`
	// IndexingMode indicates the consistency of indexing with respect to document modifications.
	// By default, the indexing mode is Consistent. This means that indexing occurs synchronously during insertion, replacment or deletion of documents.
	// To have indexing occur asynchronously, set the indexing mode to lazy.
	IndexingMode *IndexingMode `json:"indexingMode,omitempty"`
	// IncludedPaths specifies which paths must be included in indexing
	IncludedPaths []*CollectionIncludedPath `json:"includedPaths,omitempty"`
	// ExcludedPaths specifies Which paths must be excluded from indexing
	ExcludedPaths []*CollectionExcludedPath `json:"excludedPaths,omitempty"`
}

// CollectionExcludedPath represents a JSON Path to exclude from indexing inside a CollectionIndexingPolicy
type CollectionExcludedPath struct {
	Path string `json:"path"`
}

// CollectionIncludedPath represents a JSON Path and the type of data to include when indexing. This is part of a CollectionIndexingPolicy
type CollectionIncludedPath struct {
	Path    string             `json:"path"`
	Indexes []*CollectionIndex `json:"indexes"`
}

// CollectionIndex describes the type of data and precision that an included indexing path should used when being indexed.
// From [Microsoft Documentation](https://docs.microsoft.com/en-us/rest/api/cosmos-db/collections#indexing-policy)
// > The type or scheme used for index entries has a direct impact on index storage and performance.
// > For a scheme using higher precision, queries are typically faster. However, there is also a higher storage overhead for the index.
// > Choosing a lower precision means that more documents might have to be processed during query execution, but the storage overhead will be lower.
type CollectionIndex struct {
	DataType  DataType      `json:"dataType"`
	Precision int           `json:"precision,omitempty"`
	Kind      PartitionKind `json:"kind"`
}

// IndexingMode specifies how indexing will be performed on each insertion, replacement, or deletion action.
// You can choose between synchronous (Consistent), asynchronous (Lazy) index updates, and no indexing (None).
type IndexingMode string

const (
	// IndexingModeNone means no indexing is performed at all
	IndexingModeNone = IndexingMode("None")
	// IndexingModeConsistent means the index is updated synchronously on each insertion, replacement, or deletion action taken on a document in the collection (MSFT)
	IndexingModeConsistent = IndexingMode("Consistent")
	// IndexingModeLazy means the index is updated asynchronously and may be out of date, eliminating the ability to consistently "read your writes".
	IndexingModeLazy = IndexingMode("Lazy")
)

// DataType is the data type used for indexing.
type DataType string

const (
	// DataTypeString denotes a string type
	DataTypeString = DataType("String")
	// DataTypeNumber denotes a number type (int or float)
	DataTypeNumber = DataType("Number")
	// DataTypePoint denotes a GeoJSON Point. (RFC7946 § 3.1.2)
	DataTypePoint = DataType("Point")
	// DataTypePolygon denotes a GeoJSON Polygon. (RFC7946 § 3.1.6)
	DataTypePolygon = DataType("Polygon")
	// DataTypeLineString a GeoJSON LineString. (RFC7946 § 3.1.4)
	DataTypeLineString = DataType("LineString")
)

// PartitionKind is the kind of index used for equality or range comparison
type PartitionKind string

const (
	// PartititionKindHash is used to enable equality comparisons
	PartititionKindHash = PartitionKind("Hash")
	// PartititionKindRange is used to enable sorting and range comparisons
	PartititionKindRange = PartitionKind("Range")
	// PartititionKindSpatial is used to enable spacial queries (such as geographic coordinates)
	PartititionKindSpatial = PartitionKind("Spatial")
)

// CollectionPartitionKey specifies the partition key indexing config
type CollectionPartitionKey struct {
	// Paths lists the document attributes to include in the parition scheme
	Paths []string `json:"paths"`
	// Kind is the algorithm used for partitioning.
	// Note: Only PartititionKindHash is supported.
	Kind PartitionKind `json:"kind"`
}
