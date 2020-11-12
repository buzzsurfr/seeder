package internal

import "io"

type Seed struct {
	Name   string
	source io.ReadCloser
	target io.WriteCloser
	closed bool
}

func NewSeed(name string, source io.ReadCloser, target io.WriteCloser) *Seed {
	return &Seed{
		Name:   name,
		source: source,
		target: target,
		closed: false,
	}
}

func (s *Seed) Copy() (written int64, err error) {
	return io.Copy(s.target, s.source)
}

func (s *Seed) Close() error {
	source.Close()
	target.Close()
	s.closed = true
}

type Seeds []Seed
