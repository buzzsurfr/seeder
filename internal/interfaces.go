package internal

import "io"

type Source interface {
	io.ReadCloser
}

type Target interface {
	io.WriteCloser
}
