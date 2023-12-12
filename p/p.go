package p

import "io"

func Load(reader io.Reader) error               { return load(reader) }
func Collect(name string) (s string, err error) { return collect(name) }
