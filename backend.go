package contract

import (
	"errors"
)

type Runnable interface {
	Run() error
}

type Stoppable interface {
	Stop() error
}

type Backend interface {
	Runnable
	Stoppable
	BaseURL() (string, error)
}

var ErrBackendNotRunning = errors.New("backend not running")
