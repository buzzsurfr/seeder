package internal

import (
	"io"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/buzzsurfr/seeder/internal/sources/aws/ssm"
	"github.com/buzzsurfr/seeder/internal/targets/local"
	"github.com/spf13/viper"
)

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
	s.source.Close()
	s.target.Close()
	s.closed = true

	return nil
}

type Seeds []Seed

func UnmarshalSeeds(sess *session.Session, key string) Seeds {
	items := viper.Get(key)
	var seeds Seeds

	for _, item := range items.([]interface{}) {
		seed := item.(map[interface{}]interface{})
		name := seed["name"].(string)
		var source io.ReadCloser
		var target io.WriteCloser

		// Source
		sourceConfig := seed["source"].(map[interface{}]interface{})
		switch sourceConfig["type"] {
		case "ssm-parameter":
			spec := sourceConfig["spec"].(map[interface{}]interface{})
			source = ssm.NewParameter(sess, spec["name"].(string))
		}

		// Target
		targetConfig := seed["target"].(map[interface{}]interface{})
		switch targetConfig["type"] {
		case "file":
			spec := targetConfig["spec"].(map[interface{}]interface{})
			target = local.NewFile(spec["path"].(string), spec["name"].(string))
		}

		// Add seed to seeds
		seeds = append(seeds, *NewSeed(name, source, target))
	}
	return seeds
}
