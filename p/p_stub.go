//go:build !timetzdata

package p

import (
	"errors"
	"io"
)

var errDisabled = errors.New("disabled")

func load(reader io.Reader) error               { return errDisabled }
func collect(name string) (s string, err error) { return "", errDisabled }
