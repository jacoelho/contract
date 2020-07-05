package contract

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/ory/dockertest/v3"
)

const (
	defaultRepository = "pactfoundation/pact-cli"
	defaultPort       = "1234"
)

var _ Backend = (*Container)(nil)

type ContainerConfig struct {
	Repository string
	Tag        string
}

type Container struct {
	repository string
	tag        string
	resource   *dockertest.Resource
	mtx        sync.Mutex
}

func NewContractContainer(cfg ContainerConfig) (*Container, error) {
	if cfg.Repository == "" {
		cfg.Repository = defaultRepository
	}

	return &Container{
		repository: cfg.Repository,
		tag:        cfg.Tag,
	}, nil
}

func (c *Container) Run() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.resource != nil {
		if err := c.resource.Close(); err != nil {
			return err
		}
	}

	opts := &dockertest.RunOptions{
		Repository:   c.repository,
		Tag:          c.tag,
		ExposedPorts: []string{defaultPort},
		Cmd:          []string{"mock-service", "--host=0.0.0.0", "-p", defaultPort},
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return err
	}

	resource, err := pool.RunWithOptions(opts)
	if err != nil {
		return err
	}

	c.resource = resource

	return nil
}

func (c *Container) Stop() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.resource == nil {
		return nil
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return err
	}

	if err := pool.Purge(c.resource); err != nil {
		return err
	}

	c.resource = nil
	return nil
}

func (c *Container) BaseURL() (string, error) {
	c.mtx.Lock()

	if c.resource == nil {
		c.mtx.Unlock()
		return "", ErrBackendNotRunning
	}

	baseURL := "http://" + c.resource.GetHostPort(defaultPort+"/tcp")
	c.mtx.Unlock()

	err := retry(func() error {
		var err error
		resp, err := http.Get(baseURL)
		if err != nil {
			return err
		}

		defer func() {
			err = resp.Body.Close()
		}()

		_, err = ioutil.ReadAll(resp.Body)
		return err
	})
	if err != nil {
		return "", err
	}

	return baseURL, nil
}

func retry(op func() error) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5
	bo.MaxElapsedTime = time.Minute
	return backoff.Retry(op, bo)
}
