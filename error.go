package inertia

import "errors"

var (
	ErrNotFound              = errors.New("inertia-fiber: context does not have 'Inertia'")
	ErrRendererNotRegistered = errors.New("inertia-fiber: renderer not registered")
)
