# Contract

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

GNU General Public License v3.0 or later

See [LICENSE](LICENSE) to see the full text.
