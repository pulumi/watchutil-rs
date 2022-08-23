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
  foo: [bar, baz, thwomp]
  quux: []
`
	config := MustParseConfig(t, []byte(strings.TrimSpace(testConfigText)))
	assert.ElementsMatch(t, config.Types.Keys(), []string{"feat", "fix", "chore"})
	assert.ElementsMatch(t, config.Scopes.Keys(), []string{"foo", "quux"})
	fooScopes, ok := config.Scopes.Get("foo")
	assert.True(t, ok)
	assert.ElementsMatch(t, fooScopes, []string{"bar", "baz", "thwomp"})

	// zero out the template as it will clobber the autogold test, we don't care about that here:
	config.Template.Template = nil
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
  - thwomp
  quux: []
template:
  template: null
`)
	want.Equal(t, string(output))
}

func MustParseConfig(t *testing.T, value []byte) Config {
	var config Config
	err := yaml.Unmarshal([]byte(value), &config)
	assert.NoError(t, err)

	if config.Template.Template == nil {
		config.Template.Template = DefaultTemplate
	}
	return config
}
