# Contract

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/mod/github.com/jacoelho/contract?tab=overview)

[Go](https://golang.org/) Contract testing using [Pact](https://docs.pact.io/) and [Mock Service](https://github.com/pact-foundation/pact-mock_service).

## Usage

```go
package integration_test

import (
	"net/http"
	"testing"

	"github.com/jacoelho/contract"
)

func TestIntegration(t *testing.T) {
	pact := contract.TestingMockService(t, contract.WithContracts([]string{"fixtures/simple.json"}))

	resp, err := http.Get(pact.URL() + "/path")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = resp.Body.Close() }()
}
```

Interactions will be verified when test is clean-up (requires go 1.14).

Every test runs in a different container making `t.Parallel()` safe.

## License

The MIT License (MIT)

See [LICENSE](LICENSE) to see the full text.
