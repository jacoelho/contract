package contract_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/jacoelho/contract"
)

func TestContract(t *testing.T) {
	mock, err := contract.MockService(contract.WithContracts([]string{"fixtures/simple.json"}))
	if err != nil {
		t.Fatal(err)
	}

	defer mock.Cancel()

	resp, err := http.Get(mock.URL() + "/bla")
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if err := mock.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}
}