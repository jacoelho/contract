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
	"time"
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
	timeout   time.Duration
}

type Option func(*options)

// WithClient allows to specify a custom *http.Client
func WithClient(client *http.Client) Option {
	return func(args *options) {
		args.client = client
	}
}

// WithContracts allows to specify which interactions should be loaded
func WithContracts(filenames []string) Option {
	return func(args *options) {
		args.contracts = filenames
	}
}

// WithBackend allows to use a custom backend
// by default a container will be used
func WithBackend(b Backend) Option {
	return func(args *options) {
		args.backend = b
	}
}

// WithTimeout specifies how long to wait for an http request to the mock service
func WithTimeout(duration time.Duration) Option {
	return func(args *options) {
		args.timeout = duration
	}
}

// MockService represents a HTTP mock/stub implementation of Pact
type MockService struct {
	backend   Backend
	contracts []string
	client    *http.Client
	baseURL   string
	timeout   time.Duration
}

// TestingMockService creates a testing mock service
// test will fail if a call to verification return errors
func TestingMockService(t *testing.T, settings ...Option) *MockService {
	args := &options{
		client:  http.DefaultClient,
		timeout: 5 * time.Second,
	}

	for _, opt := range settings {
		opt(args)
	}

	mock := &MockService{
		contracts: args.contracts,
		timeout:   args.timeout,
		client:    args.client,
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

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), mock.timeout)
		defer cancel()

		if err := mock.Verify(ctx); err != nil {
			t.Error(err)
		}

		if err := mock.backend.Stop(); err != nil {
			t.Error(err)
		}
	})

	deleteCtx, deleteCancel := context.WithTimeout(context.Background(), mock.timeout)
	defer deleteCancel()

	// clean-up any health check interaction
	if err := mock.Delete(deleteCtx); err != nil {
		t.Fatal(err)
	}

	createCtx, createCancel := context.WithTimeout(context.Background(), mock.timeout)
	defer createCancel()
	if err := mock.createInteractions(createCtx); err != nil {
		t.Fatal(err)
	}

	return mock
}

// URL returns a URL ready to be interacted, URL will be have the form `localhost:port`
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

// Create creates an interaction to be tested
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

// Delete deletes all interactions
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

// Verify verifies interactions
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
