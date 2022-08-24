package changelog

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/hexops/autogold"
	"github.com/stretchr/testify/assert"
)

func TestConfigRemarshaling(t *testing.T) {
	testConfigText := `
types:
  feat: Features
  fix: Bug Fixes
  chore: Miscellaneous Tasks
scopes:
  foo: [bar, baz]
`
	config := MustParseConfig(t, []byte(strings.TrimSpace(testConfigText)))
	assert.ElementsMatch(t, config.Types.Keys(), []string{"feat", "fix", "chore"})
	assert.ElementsMatch(t, config.Scopes.Keys(), []string{"foo"})

	output, err := yaml.Marshal(config)
	assert.NoError(t, err)

	want := autogold.Want("config_roundtrip", `types:
  feat: Features
  fix: Bug Fixes
  chore: Miscellaneous Tasks
scopes:
  foo:
  - bar
  - baz
`)
	want.Equal(t, string(output))
}

func MustParseConfig(t *testing.T, value []byte) Config {
	var config Config
	err := yaml.Unmarshal([]byte(value), &config)
	assert.NoError(t, err)
	return config
}
