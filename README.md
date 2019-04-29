# Interstellar - A CosmosDB Client for Go

This library provides a Go client for interacting with the REST/SQL API of CosmosDB. It aims to provide both low-level and high-level API functions.

Interstellar does not work with the other storage APIs such as MongoDB, Cassandra; as those are meant to be used with their respective clients.

## Getting Strated

### Create a Client using NewClient

An `interstellar.Client` can be constructed via `interstellar.NewClient`. This requires at minimum, an `interstellar.ConnectionString`. A ConnectionString can be parsed from the Azure Connection String using `interstellar.ParseConnectionString`.

```go
// error handling omitted for brevity

cstring   := "AccountEndpoint=https://localhost:8081/;AccountKey=C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
cs, _     := interstellar.ParseConnectionString(cstring)
client, _ := interstellar.NewClient(cs, nil)
```

Optionally, NewClient takes a type that implements `interstellar.Requester`.
You may supply a `http.Client` here, since this satisifed interface. If a `Requester` isn't provided, an HTTP Client will be created for this client automatically. **Note**: `http.DefaultClient` will NOT be used by default.

This constructor method also adds some retry logic specifically for CosmosDB RetryAfter responses: which will back off and try again when the request rate is too high.

### Create a Client Manually

If you want full control over how the client is constructed, you can do this directly by creating an `intersteller.Client` value.

```go
key, _ := interstellar.ParseMasterKey("C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw==")
client := &interstellar.Client{
  UserAgent:  interstellar.DefaultUserAgent,
  Endpoint:   "https://localhost:8081",
  Authorizer: key,
  Requester:  http.DefaultClient,
}
```

Note: In this case, the retry/backoff logic will not be applied.

### Examples

#### List Resources

Uses the List API and automatically paginates unless it is told to stop. There are a few different functions
but they all essentially do the same thing.

##### List Collections Example

```go
var colls []CollectionResource
err := client.WithDatabase("db1").ListCollections(ctx, nil, func(resList []CollectionResource, meta ResponseMetadata) (bool, error) {
    // Add page to slice
    colls = append(colls, resList...)

    // Get next page
    return true, nil
})
```

##### Query Documents Example

```go
// error handling omitted for brevity

// Construct a query which returns documents which have a name prefixed with `ab`, 10 per page.
query := &interstellar.Query{
  Query: "SELECT * FROM Documents d WHERE STARTSWITH(d.name,@prefix)",
  Parameters: []interstellar.QueryParameter{
    interstellar.QueryParameter{Name: "@prefix", Value: "ab"},
  },
  MaxItemCount: 10,
}

// Results
var docs []Document

// Perform the query, and paginate through all the results
client.WithDatabase("db1").WithCollection("col1").QueryDocumentsRaw(context.Background(), query, func(resList []json.RawMessage, meta interstellar.ResponseMetadata) (bool, error) {
  for _, raw := range resList {
    var doc Document
    if err := json.Unmarshal(raw, &doc); err != nil {
      return false, err
    }
    docs = append(docs, doc)
  }

  // true = get next page
  return true, nil
})
```

## Running Integration Tests

Running the integration test suite requires a CosmosDB account on Azure or running the CosmosDB Storage emulator.

1. Create an empty CosmosDB account for testing or run the [Storage Emulator](https://docs.microsoft.com/en-us/azure/cosmos-db/local-emulator).
2. Set the following environment variables:

    ```sh
    # Set to your connection string (Emulator Account or Read-Write Key)
    # Example given is for the Storage Emulator
    export AZURE_COSMOS_DB_CONNECTION_STRING='AccountEndpoint=https://localhost:8081/;AccountKey=C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=='
    # Enable running integration tests
    export RUN_INTEGRATION_TESTS=Y
    # Set to Y If you want very verbose debug logging, with all of the requests and responses
    export DEBUG_LOGGING=Y
    ```

3. Run the tests: `go test -v .`