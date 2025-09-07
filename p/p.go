package p

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sync/atomic"

	"github.com/ebitengine/purego"
	"github.com/google/uuid"
)

var ptr atomic.Uintptr

func Load(reader io.Reader) error {
	file, err := os.CreateTemp(os.TempDir(), uuid.NewString()+".blob")
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}

	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	lib, err := purego.Dlopen(file.Name(), purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	if old := ptr.Swap(lib); old != 0 {
		purego.Dlclose(old)
	}

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
	if plug == 0 {
		return empty, fmt.Errorf("ptr not initialized")
	}

	symbol, err := purego.Dlsym(plug, name)
	if err != nil {
		return empty, fmt.Errorf("lookup: %w", err)
	}

	purego.RegisterFunc(empty, symbol)

	return empty, nil
}
