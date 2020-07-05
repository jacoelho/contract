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

func WithClient(client *http.Client) Option {
	return func(args *options) {
		args.client = client
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

type MockService struct {
	backend   Backend
	contracts []string
	client    *http.Client
	baseURL   string
}

func TestingMockService(t *testing.T, settings ...Option) *MockService {
	args := &options{}

	for _, opt := range settings {
		opt(args)
	}

	mock := &MockService{
		contracts: args.contracts,
	}

	if mock.client == nil {
		mock.client = http.DefaultClient
	}

	if mock.backend == nil {
		b, err := NewContractContainer(ContainerConfig{})
		if err != nil {
			t.Fatal(err)
		}

		mock.backend = b
	}

	if err := mock.backend.Run(); err != nil {
		t.Fatal(err)
	}

	baseURL, err := mock.backend.BaseURL()
	if err != nil {
		t.Fatal(err)
	}

	mock.baseURL = baseURL

	// clean-up any health check interaction
	if err := mock.Delete(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := mock.createInteractions(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := mock.Verify(context.Background()); err != nil {
			t.Error(err)
		}

		if err := mock.backend.Stop(); err != nil {
			t.Error(err)
		}
	})

	return mock
}

func (m *MockService) URL() string {
	return m.baseURL
}

func (m *MockService) createInteractions(ctx context.Context) error {
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

func (m *MockService) Create(ctx context.Context, r io.Reader) error {
	req, err := m.newHttpRequest(http.MethodPost, "/interactions", r)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	return checkHttpStatus(resp)
}

func (m *MockService) Delete(ctx context.Context) error {
	req, err := m.newHttpRequest(http.MethodDelete, "/interactions", nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	return checkHttpStatus(resp)
}

func (m *MockService) Verify(ctx context.Context) error {
	req, err := m.newHttpRequest(http.MethodGet, "/interactions/verification", nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	return checkHttpStatus(resp)
}

func (m *MockService) newHttpRequest(httpMethod, relativePath string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(httpMethod, m.baseURL+relativePath, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add(pactHeader, pactHeaderValue)
	req.Header.Add(contentType, contentTypeJSON)

	return req, nil
}

func checkHttpStatus(resp *http.Response) error {
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
