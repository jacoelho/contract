package contract

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
)

const (
	pactHeader      = "X-Pact-Mock-Service"
	pactHeaderValue = "true"
	contentType     = "Content-Type"
	contentTypeJSON = "application/json"
)

type options struct {
	backend   Backend
	client    *http.Client
	contracts []string
}

type Option func(*options)

func WithClient(c *http.Client) Option {
	return func(args *options) {
		args.client = c
	}
}

func WithContracts(filenames []string) Option {
	return func(args *options) {
		args.contracts = filenames
	}
}

func WithBackend(b Backend) Option {
	return func(args *options) {
		if b != nil {
			args.backend = b
		}
	}
}

type ContainerMockService struct {
	backend   Backend
	client    *http.Client
	contracts []string
	baseURL   string
}

func MockService(t *testing.T, settings ...Option) *ContainerMockService {
	args := &options{}

	for _, opt := range settings {
		opt(args)
	}

	m := &ContainerMockService{
		client:    args.client,
		contracts: args.contracts,
	}

	if m.client == nil {
		m.client = http.DefaultClient
	}

	if m.backend == nil {
		b, err := NewContractContainer(ContractContainerConfig{})
		if err != nil {
			t.Fatal(err)
		}

		m.backend = b
	}

	if err := m.backend.Run(); err != nil {
		t.Fatal(err)
	}

	baseURL, err := m.backend.BaseURL()
	if err != nil {
		t.Fatal(err)
	}

	m.baseURL = baseURL

	if err := m.createInteractions(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := m.Verify(context.Background()); err != nil {
			t.Error(err)
		}

		if err := m.backend.Stop(); err != nil {
			t.Error(err)
		}
	})

	return m
}

func (m *ContainerMockService) URL() string {
	return m.baseURL
}

func (m *ContainerMockService) createInteractions(ctx context.Context) error {
	if err := m.Delete(ctx); err != nil {
		return err
	}

	for _, contract := range m.contracts {
		f, err := os.Open(path.Clean(contract))
		if err != nil {
			return err
		}

		err = m.Create(ctx, f)
		_ = f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ContainerMockService) Create(ctx context.Context, r io.Reader) error {
	req, err := http.NewRequest(http.MethodPost, m.baseURL+"/interactions", r)
	if err != nil {
		return err
	}

	req.Header.Add(pactHeader, pactHeaderValue)
	req.Header.Add(contentType, contentTypeJSON)

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errors.New("request failed")
	}

	return nil
}

func (m *ContainerMockService) Delete(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodDelete, m.baseURL+"/interactions", nil)
	if err != nil {
		return err
	}

	req.Header.Add(pactHeader, pactHeaderValue)
	req.Header.Add(contentType, contentTypeJSON)

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errors.New("request failed")
	}

	return nil
}

func (m *ContainerMockService) Verify(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodGet, m.baseURL+"/interactions/verification", nil)
	if err != nil {
		return err
	}

	req.Header.Add(pactHeader, pactHeaderValue)
	req.Header.Add(contentType, contentTypeJSON)

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(data))
	}

	return nil
}
