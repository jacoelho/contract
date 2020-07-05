package contract_test

import (
	"net/http"
	"testing"

	"github.com/jacoelho/contract"
)

func TestContractSimple(t *testing.T) {
	t.Parallel()
	pact := contract.TestingMockService(t, contract.WithContracts([]string{"fixtures/simple.json"}))

	resp, err := http.Get(pact.URL() + "/path_one")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = resp.Body.Close() }()
}

func TestContractMultiple(t *testing.T) {
	t.Parallel()
	pact := contract.TestingMockService(t, contract.WithContracts([]string{
		"fixtures/pact-1.json",
		"fixtures/pact-2.json",
	}))

	{
		resp, err := http.Get(pact.URL() + "/path1")
		if err != nil {
			t.Error(err)
		}
		defer func() { _ = resp.Body.Close() }()
	}
	{
		resp, err := http.Get(pact.URL() + "/path2")
		if err != nil {
			t.Error(err)
		}
		defer func() { _ = resp.Body.Close() }()
	}

}
