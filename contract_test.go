package contract_test

import (
	"net/http"
	"testing"

	"github.com/jacoelho/contract"
)

func TestContract(t *testing.T) {
	pact := contract.MockService(t, contract.WithContracts([]string{"fixtures/simple.json"}))

	resp, err := http.Get(pact.URL() + "/path")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = resp.Body.Close() }()
}
