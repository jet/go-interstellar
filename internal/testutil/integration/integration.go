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

package integration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jet/go-interstellar"
	"github.com/jet/go-interstellar/internal/testutil"
)

// Mark sets the test as an integration test
// which will only run when RUN_INTEGRATION_TESTS=Y
func Mark(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_INTEGRATION_TESTS") != "Y" {
		t.Skip("Skipped integration test. Set Environment variable RUN_INTEGRATION_TESTS=Y to run")
	}
}

// noop function does nothing
// returned by helper functions that error out and do not need a cleanup function to be run
func noop() {}

// LoadDatabase creates a new database named after the folder pointed at by `path`
// Then for each sub-directory, calls 'LoadCollection'
//
// Returns a function that will delete the database (for cleanup purposes)
func LoadDatabase(t *testing.T, client *interstellar.Client, path string) func() {
	t.Helper()
	dbid := filepath.Base(path)
	_, _, err := client.CreateDatabase(nil, dbid, nil)
	if err != nil {
		t.Errorf("error creating database: '%s': %v", dbid, err)
		return noop
	}
	var dbres *interstellar.DatabaseResource
	db := client.WithDatabase(dbid)

	if dbres, _, err = db.Get(nil, nil); err != nil {
		t.Errorf("error getting database: '%s': %v", dbid, err)
		return noop
	}
	testutil.DebugLog(t, "Database Created:\n%s", testutil.ToJSON(dbres))

	finfo, err := ioutil.ReadDir(path)
	if err != nil {
		t.Fatalf("could not read dir '%s': %v", path, err)
	}
	var cleanup []func()
	for _, info := range finfo {
		if info.IsDir() {
			cleanup = append(cleanup, LoadCollection(t, db, filepath.Join(path, info.Name())))
		}
	}
	return func() {
		for _, fn := range cleanup {
			fn()
		}
		ok, meta, err := db.Delete(nil, nil)
		if err != nil || !ok {
			t.Errorf("unable to delete db '%s': %v", dbid, err)
			return
		}
		testutil.DebugLog(t, "Database Deleted. Metadata:\n%s", testutil.ToJSON(meta))
	}
}

func readCollectionRequest(t *testing.T, path string) *interstellar.CreateCollectionRequest {
	t.Helper()
	testutil.DebugLog(t, "read collection request: %s", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("unable to open '%s': %v", path, err)
		return nil
	}
	var req interstellar.CreateCollectionRequest
	if err = json.Unmarshal(data, &req); err != nil {
		t.Fatalf("unable to decode CreateCollectionRequest: %v", err)
		return nil
	}
	return &req
}

// LoadCollection creates a new collection defined by the file 'col.json' in the given path.
// Then loads all of the documents in docs.json into the given collection
//
// Returns a function that will delete the collection (for cleanup purposes)
func LoadCollection(t *testing.T, client *interstellar.DatabaseClient, path string) func() {
	t.Helper()
	req := readCollectionRequest(t, filepath.Join(path, "col.json"))
	if req == nil {
		return noop
	}
	_, _, err := client.CreateCollection(nil, *req)
	if err != nil {
		t.Errorf("error creating database: '%s': %v", req.ID, err)
		return noop
	}
	var colres *interstellar.CollectionResource
	col := client.WithCollection(req.ID)
	if colres, _, err = col.Get(nil, nil); err != nil {
		t.Errorf("error getting collection: '%s': %v", req.ID, err)
		return noop
	}
	testutil.DebugLog(t, "Collection Created:\n%s", testutil.ToJSON(colres))
	if colres.PartitionKey != nil {
		LoadDocumentsPartitioned(t, col, filepath.Join(path, "pdocs.json"))
	} else {
		LoadDocuments(t, col, filepath.Join(path, "docs.json"))
	}
	return func() {
		ok, meta, err := col.Delete(nil, nil)
		if err != nil || !ok {
			t.Errorf("unable to delete collection '%s': %v", req.ID, err)
		}
		testutil.DebugLog(t, "Collection Deleted. Metadata:\n%s", testutil.ToJSON(meta))
	}
}

// PartitionedDocs is a document with a partition key
type partitionedDoc struct {
	PartitionKey []string        `json:"partitionKey"`
	Document     json.RawMessage `json:"doc"`
}

// LoadDocumentsPartitioned loads all of the documents in the json file 'path' into the given collection which have partition keys assigned
func LoadDocumentsPartitioned(t *testing.T, client *interstellar.CollectionClient, path string) {
	alldocs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file '%s': %v", path, err)
	}
	docslist, err := interstellar.ParseArrayResponse(bytes.NewReader(alldocs))
	if err != nil {
		t.Fatalf("could not parse file '%s': %v", path, err)
	}
	for dn, data := range docslist {
		var pdoc partitionedDoc
		if err = json.Unmarshal(data, &pdoc); err != nil {
			t.Errorf("error decoding paritioned document[%d]: %v", dn, err)
			continue
		}
		var props interstellar.DocumentProperties
		if err = json.Unmarshal(pdoc.Document, &props); err != nil {
			t.Errorf("error decoding paritioned document[%d] properties: %v", dn, err)
			continue
		}
		var req interstellar.CreateDocumentRequest
		req.Body = pdoc.Document
		req.PartitionKey = pdoc.PartitionKey
		if _, _, err = client.CreateDocument(nil, req); err != nil {
			t.Errorf("error creating document: '%s': %v", props.ID, err)
			continue
		}
		doc := client.WithDocument(props.ID, pdoc.PartitionKey)
		var docbs []byte
		docbs, _, err = doc.GetRaw(nil, nil)
		if err != nil {
			t.Errorf("error getting document: '%s': %v", props.ID, err)
			continue
		}
		testutil.DebugLog(t, "Document Created:\n%s", testutil.FormatJSON(docbs))
		if err = json.Unmarshal(docbs, &props); err != nil {
			t.Errorf("error decoding document[%d] properties: %v", dn, err)
			continue
		}
	}
}

// LoadDocuments loads all of the documents in the json file 'path' into the given collection
func LoadDocuments(t *testing.T, client *interstellar.CollectionClient, path string) {
	alldocs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file '%s': %v", path, err)
	}
	docslist, err := interstellar.ParseArrayResponse(bytes.NewReader(alldocs))
	if err != nil {
		t.Fatalf("could not parse file '%s': %v", path, err)
	}
	for dn, data := range docslist {
		var props interstellar.DocumentProperties
		if err = json.Unmarshal(data, &props); err != nil {
			t.Errorf("error decoding document[%d] properties: %v", dn, err)
			continue
		}
		var req interstellar.CreateDocumentRequest
		req.Body = data
		if _, _, err = client.CreateDocument(nil, req); err != nil {
			t.Errorf("error creating document: '%s': %v", props.ID, err)
			continue
		}
		doc := client.WithDocument(props.ID, nil)
		var docbs []byte
		docbs, _, err = doc.GetRaw(nil, nil)
		if err != nil {
			t.Errorf("error getting document: '%s': %v", props.ID, err)
			continue
		}
		testutil.DebugLog(t, "Document Created:\n%s", testutil.FormatJSON(docbs))
		if err = json.Unmarshal(docbs, &props); err != nil {
			t.Errorf("error decoding document[%d] properties: %v", dn, err)
			continue
		}
	}
}
