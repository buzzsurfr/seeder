package internal

import "io"

type Seed struct {
	Name   string
	source io.ReadCloser
	target io.WriteCloser
}
