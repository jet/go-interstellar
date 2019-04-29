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

package interstellar_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jet/go-interstellar"
	"github.com/jet/go-interstellar/internal/testutil"
	"github.com/jet/go-interstellar/internal/testutil/deep"
	"github.com/jet/go-interstellar/internal/testutil/integration"
)

func TestIntegrationLoadData(t *testing.T) {
	integration.Mark(t)
	client := testutil.CreateTestClient(t)
	path := "./testdata/databases/"
	finfo, err := ioutil.ReadDir(path)
	if err != nil {
		t.Fatalf("could not read dir '%s': %v", path, err)
	}
	for _, info := range finfo {
		if info.IsDir() {
			defer integration.LoadDatabase(t, client, filepath.Join(path, info.Name()))()
		}
	}
}

func TestIntegrationOffers(t *testing.T) {
	integration.Mark(t)
	client := testutil.CreateTestClient(t)
	ctx := context.Background()
	defer integration.LoadDatabase(t, client, "./testdata/databases/db1")()

	// Enumerate collections
	var colls []interstellar.CollectionResource
	if err := client.WithDatabase("db1").ListCollections(ctx, nil, func(resList []interstellar.CollectionResource, meta interstellar.ResponseMetadata) (bool, error) {
		colls = append(colls, resList...)
		return true, nil
	}); err != nil {
		t.Errorf("list collections failed: %v", err)
		return
	}

	// Enumerate offers
	var offers []interstellar.OfferResource
	if err := client.ListOffers(ctx, nil, func(resList []interstellar.OfferResource, meta interstellar.ResponseMetadata) (bool, error) {
		offers = append(offers, resList...)
		return true, nil
	}); err != nil {
		t.Errorf("list offers failed: %v", err)
		return
	}
	if len(colls) != 1 {
		t.Errorf("expected 1 collection, got %d", len(colls))
		return
	}
	coll := colls[0]
	var found *interstellar.OfferResource
	for i := range offers {
		offer := offers[i]
		if offer.OfferResourceID == coll.ResourceID {
			found = &offer
			break
		}
	}
	if found == nil {
		t.Errorf("did not find matching offer for collection %#v", coll)
		return
	}
	if found.Content == nil || found.Content.V2 == nil {
		t.Errorf("offer content is empty: %s", testutil.ToJSON(found))
		return
	}
	offer, meta, err := client.WithOffer(found.ID).Get(ctx, nil)
	if err != nil {
		t.Errorf("offer '%s' could not be retrieved: %v", found.ID, err)
		return
	}
	t.Logf("Get Offer:\n%s", testutil.ToJSON(offer))
	t.Logf("Get Offer Metadata:\n%s", testutil.ToJSON(meta))
	found.Content.V2.OfferThroughput = 3000
	o, meta, err := client.ReplaceOffer(ctx, interstellar.ReplaceOfferRequest{
		Offer: found,
		Options: &interstellar.CommonRequestOptions{
			IfMatch: found.ETag,
		},
	})
	if err != nil {
		t.Errorf("offer could not be replaced: %v", err)
		return
	}
	t.Logf("Replaced Offer:\n%s", testutil.ToJSON(o))
	t.Logf("Replaced Offer Metadata:\n%s", testutil.ToJSON(meta))
}

type accountEvent struct {
	ID            string `json:"id"`
	AccountNumber string `json:"AccountNumber"`
	Delta         int    `json:"Delta"`
}

type Family struct {
	ID       string `json:"id"`
	LastName string `json:"lastName"`
	Parents  []struct {
		FirstName string `json:"firstName"`
	} `json:"parents"`
	Children []struct {
		FirstName string `json:"firstName"`
		Gender    string `json:"gender"`
		Grade     int    `json:"grade"`
		Pets      []struct {
			GivenName string `json:"givenName"`
		}
	} `json:"children"`
	Address struct {
		State  string `json:"state"`
		County string `json:"county"`
		City   string `json:"city"`
	} `json:"address"`
	CreationDate int64 `json:"creationDate"`
	IsRegistered bool  `json:"isRegistered"`
}

func TestIntegrationCreateAndReplaceDocument(t *testing.T) {
	integration.Mark(t)
	client := testutil.CreateTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	defer integration.LoadDatabase(t, client, "./testdata/databases/db2")()

	cc := client.WithDatabase("db2").WithCollection("families")
	_, _, err := cc.Get(ctx, nil)
	if err != nil {
		t.Fatalf("unable to get collection: %v", err)
	}
	var griffin Family
	griffin.ID = "GriffinFamily"
	griffin.LastName = "Griffin"
	griffin.CreationDate = 0
	griffin.Parents = []struct {
		FirstName string `json:"firstName"`
	}{
		{FirstName: "Peter"},
		{FirstName: "Louis"},
	}
	griffin.Children = []struct {
		FirstName string `json:"firstName"`
		Gender    string `json:"gender"`
		Grade     int    `json:"grade"`
		Pets      []struct {
			GivenName string `json:"givenName"`
		}
	}{
		{FirstName: "Meg", Gender: "female"},
		{FirstName: "Cris", Gender: "male"},
		{FirstName: "Stewey", Gender: "male", Pets: []struct {
			GivenName string `json:"givenName"`
		}{
			{GivenName: "Brian"},
		}},
	}
	if _, _, err := cc.CreateDocument(ctx, interstellar.CreateDocumentRequest{
		Document:     &griffin,
		PartitionKey: []string{griffin.ID},
	}); err != nil {
		t.Fatalf("unable to create document: %v", err)
	}
	initial := griffin.CreationDate

	dc := cc.WithDocument("GriffinFamily", []string{"GriffinFamily"})
	latch := make(chan struct{})
	var wg sync.WaitGroup
	incrementFn := func(i int) {
		<-latch
		defer wg.Done()
		var andersen Family
		var etag string
		for { // retry until success
			select {
			case <-ctx.Done():
				t.Errorf("%d: timeout", i)
				return
			default:
			}
			if meta, err := dc.Get(ctx, nil, &andersen); err != nil {
				t.Errorf("%d: unable to get document: %v", i, err)
				return
			} else {
				etag = meta.ETag
			}
			andersen.CreationDate += 100
			_, _, err := dc.ReplaceDocument(ctx, interstellar.ReplaceDocumentRequest{
				ETag:     etag,
				Document: andersen,
			})
			if err == nil {
				return
			}
			if err == interstellar.ErrPreconditionFailed {
				t.Logf("%d: precondition failed, try again", i)
			} else {
				t.Errorf("%d: ReplaceDocument Err: %v", i, err)
				return
			}

		}
	}
	num := 5
	wg.Add(num)
	for i := 0; i < num; i++ {
		go incrementFn(i)
	}
	close(latch)
	wg.Wait()
	if _, err := dc.Get(ctx, nil, &griffin); err != nil {
		t.Fatalf("unable to get document: %v", err)
	}
	if griffin.CreationDate != initial+int64(num*100) {
		t.Fatal("optimistic concurrency error: did not have expected value after all goroutines exited")
	}
}

func TestIntegrationListDocuments(t *testing.T) {
	integration.Mark(t)
	client := testutil.CreateTestClient(t)
	ctx := context.Background()
	defer integration.LoadDatabase(t, client, "./testdata/databases/db1")()

	// Enumerate events
	var events []accountEvent
	opts := &interstellar.CommonRequestOptions{
		ActivityID: "foo",
	}
	if err := client.WithDatabase("db1").WithCollection("col1").ListDocumentsRaw(ctx, opts, func(resList []json.RawMessage, meta interstellar.ResponseMetadata) (bool, error) {
		for _, raw := range resList {
			var event accountEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				return false, err
			}
			events = append(events, event)
		}
		return true, nil
	}); err != nil {
		t.Errorf("query documents failed: %v", err)
		return
	}
	if len(events) != 101 {
		t.Errorf("expected 101 documents, got %d", len(events))
		return
	}
	expected := map[string]int{
		"100": 2547,
		"200": 1984,
		"300": 416,
		"400": 436,
		"500": 1826,
	}
	balances := make(map[string]int)
	for _, e := range events {
		balances[e.AccountNumber] = balances[e.AccountNumber] + e.Delta
	}
	if diffs := deep.Equal(expected, balances); diffs != nil {
		t.Errorf("account balances do not match: %v", diffs)
		return
	}
}

func TestIntegrationQueryDocuments(t *testing.T) {
	integration.Mark(t)
	client := testutil.CreateTestClient(t)
	ctx := context.Background()
	defer integration.LoadDatabase(t, client, "./testdata/databases/db1")()

	// Enumerate collections
	var colls []interstellar.CollectionResource
	if err := client.WithDatabase("db1").ListCollections(ctx, nil, func(resList []interstellar.CollectionResource, meta interstellar.ResponseMetadata) (bool, error) {
		colls = append(colls, resList...)
		return true, nil
	}); err != nil {
		t.Errorf("list collections failed: %v", err)
		return
	}

	// Enumerate events
	var events []accountEvent
	query := &interstellar.Query{
		Query: `SELECT * FROM Events e WHERE e.AccountNumber = @acct`,
		Parameters: []interstellar.QueryParameter{
			interstellar.QueryParameter{Name: "@acct", Value: "100"},
		},
		MaxItemCount: 5,
	}
	if err := client.WithDatabase("db1").WithCollection("col1").QueryDocumentsRaw(ctx, query, func(resList []json.RawMessage, meta interstellar.ResponseMetadata) (bool, error) {
		for _, raw := range resList {
			var event accountEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				return false, err
			}
			events = append(events, event)
		}
		return true, nil
	}); err != nil {
		t.Errorf("query documents failed: %v", err)
		return
	}
	if len(events) != 18 {
		t.Errorf("expected 18 documents, got %d", len(events))
		return
	}
	balance := 0
	for _, e := range events {
		balance += e.Delta
	}
	if balance != 2547 {
		t.Errorf("expected events account balance=2547, got %d", balance)
		return
	}
}
