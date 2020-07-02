package contract_test

import (
	"testing"
	"time"

	"github.com/jacoelho/contract"
)

func TestContract(t *testing.T) {
	mock, err := contract.MockService(contract.WithContracts([]string{"fixture/a.json"})
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Cancel()

	resp, err := http.Get(mock.URL() + "/bla")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
}