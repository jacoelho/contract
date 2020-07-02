package contract

import (
	"io/ioutil"
	"net/http"

	"github.com/ory/dockertest/v3"
)

// Cancelable cancels
type Cancelable func() error

// TestContract allows to runs contract testing in containers
func TestContract() (string, Cancelable, error) {
	pool, err := dockertest.NewPool("")

	opts := &dockertest.RunOptions{
		Repository:   "pactfoundation/pact-cli",
		Tag:          "latest",
		ExposedPorts: []string{"1234"},
		Cmd:          []string{"mock-service", "--host=0.0.0.0", "-p=1234"},
	}

	resource, err := pool.RunWithOptions(opts)
	if err != nil {
		return "", nil, err
	}

	url := "http://localhost:" + resource.GetPort("1234/tcp")

	if err := pool.Retry(func() error {
		var err error
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		_, err = ioutil.ReadAll(resp.Body)
		return err
	}); err != nil {
		return "", nil, err
	}

	cancel := func() error {
		return pool.Purge(resource)
	}

	return url, cancel, nil
}
