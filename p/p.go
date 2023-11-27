package p

import (
	"fmt"
	"io"
	"os"
	node "plugin"
	"runtime/debug"
	"sync/atomic"

	"github.com/google/uuid"
)

var ptr atomic.Pointer[node.Plugin]

func Load(reader io.Reader) error {
	file, err := os.CreateTemp(os.TempDir(), uuid.NewString()+".blob")
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}

	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	plug, err := node.Open(file.Name())
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	ptr.Store(plug)

	return nil
}

func Collect(name string) (s string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\nstack: %s", r, string(debug.Stack()))
		}
	}()

	call, err := lookup[func() string](name)
	if err != nil {
		return "", err
	}

	if call != nil {
		return call(), nil
	}

	return "<nil>", nil
}

func lookup[T any](name string) (T, error) {
	var empty T

	plug := ptr.Load()

	if plug == nil {
		return empty, fmt.Errorf("ptr not initialized")
	}

	symbol, err := plug.Lookup(name)
	if err != nil {
		return empty, fmt.Errorf("lookup: %w", err)
	}

	casted, ok := symbol.(T)
	if !ok {
		return empty, fmt.Errorf("type %T cant be casted to %T", symbol, empty)
	}

	return casted, nil
}
