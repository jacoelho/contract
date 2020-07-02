package contract

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

type Interactor interface {
	Create(context.Context, io.Reader) error
	Delete(context.Context) error
	Verify(context.Context) error
}

const (
	pactHeader      = "X-Pact-Mock-Service"
	pactHeaderValue = "true"
	contentType     = "Content-Type"
	contentTypeJSON = "application/json"
)

type options struct {
	BaseURL   string
	Client    *http.Client
	Contracts []string
}

type Option func(*options)

func WithClient(c *http.Client) Option {
	return func(args *options) {
		args.Client = c
	}
}

func WithContracts(filename []string) Option {
	return func(args *options) {
		args.Contracts = filename
	}
}

func WithBaseURL(baseURL string) Option {
	return func(args *options) {
		args.BaseURL = baseURL
	}
}

type ContainerMockService struct {
	baseURL   string
	client    *http.Client
	contracts []string
	cancel    Cancelable
}

func MockService(settings ...Option) (*ContainerMockService, error) {
	url, cancel, err := TestContract()
	if err != nil {
		return nil, err
	}

	args := &options{
		BaseURL: url,
		Client:  http.DefaultClient,
	}

	for _, opt := range settings {
		opt(args)
	}

	m := &ContainerMockService{
		baseURL:   args.BaseURL,
		client:    args.Client,
		contracts: args.Contracts,
		cancel:    cancel,
	}

	if err := m.createInteractions(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *ContainerMockService) Cancel() error {
	return m.cancel()
}

func (m *ContainerMockService) URL() string {
	return m.baseURL
}

func (m *ContainerMockService) createInteractions() error {
	for _, contract := range m.contracts {
		f, err := os.Open(path.Clean(contract))
		if err != nil {
			return err
		}

		err = m.Create(context.Background(), f)
		f.Close()
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

	defer resp.Body.Close()

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

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("request failed")
	}

	return nil
}

func (m *ContainerMockService) Verify(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodGet, m.baseURL+"/interactions/verify", nil)
	if err != nil {
		return err
	}

	req.Header.Add(pactHeader, pactHeaderValue)

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	return errors.New("request failed")
	// }

	fmt.Println(resp.StatusCode) // 500

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
