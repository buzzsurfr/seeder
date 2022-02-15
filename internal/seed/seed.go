package seed

import (
	"io"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/buzzsurfr/seeder/internal"
	"github.com/buzzsurfr/seeder/internal/sources/aws/secretsmanager"
	"github.com/buzzsurfr/seeder/internal/sources/aws/ssm"
	"github.com/buzzsurfr/seeder/internal/targets/local"
	"github.com/spf13/viper"
)

// Seed is the atomic unit of seeder
type Seed struct {
	Name   string
	Source internal.Source
	Target internal.Target
}

// NewSeed creates a new Seed
func NewSeed(name string, source internal.Source, target internal.Target) *Seed {
	return &Seed{
		Name:   name,
		Source: source,
		Target: target,
	}
}

// Copy copies seed from source to target
func (s *Seed) Copy() (written int64, err error) {
	return io.Copy(s.Target, s.Source)
}

// Close closes all dependencies of the Seed
func (s *Seed) Close() error {
	s.Source.Close()
	s.Target.Close()

	return nil
}

// Seeds are a collection of Seed
type Seeds []Seed

// UnmarshalSeeds reads a key from viper and returns Seeds
func UnmarshalSeeds(sess *session.Session, key string) Seeds {
	items := viper.Get(key)
	var seeds Seeds

	for _, item := range items.([]interface{}) {
		seed := item.(map[interface{}]interface{})
		name := seed["name"].(string)
		var source internal.Source
		var target internal.Target

		// Source
		sourceConfig := seed["source"].(map[interface{}]interface{})
		switch sourceConfig["type"] {
		case "ssm-parameter":
			spec := sourceConfig["spec"].(map[interface{}]interface{})
			source = ssm.NewParameter(sess, spec["name"].(string))
		case "secretsmanager":
			spec := sourceConfig["spec"].(map[interface{}]interface{})
			source = secretsmanager.NewSecret(sess, spec["secretId"].(string))
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
