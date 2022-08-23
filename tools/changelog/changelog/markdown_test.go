package changelog

import (
	"math/rand"
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/assert"
)

func TestMarkdown(t *testing.T) {
	t.Parallel()

	changes := []*Entry{{
		Type: "improvement",
		Scope: Scope{
			Primary:   "cli",
			SubScopes: []string{"engine", "backend"},
		},
		Description: "Foo bar baz",
		GitHubMeta: GitHubMeta{
			PullRequestNumbers: []int{9001},
		},
	}, {
		Type: "improvement",
		Scope: Scope{
			Primary:   "sdk",
			SubScopes: []string{"go"},
		},
		Description: "Make SDK go brrrr",
		GitHubMeta: GitHubMeta{
			PullRequestNumbers: []int{10000},
		},
	}, {
		Type: "fix",
		Scope: Scope{
			Primary:   "sdkgen",
			SubScopes: []string{"nodejs"},
		},
		Description: "Fix Typescript SDK code generation.",
		GitHubMeta: GitHubMeta{
			PullRequestNumbers: []int{20000},
		},
	}, {
		Type: "fix",
		Scope: Scope{
			Primary:   "sdkgen",
			SubScopes: []string{"go"},
		},
		Description: "Fix Go SDK code generation",
		GitHubMeta: GitHubMeta{
			PullRequestNumbers: []int{20001},
		},
	}}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(changes), func(i1, i2 int) {
		changes[i1], changes[i2] = changes[i2], changes[i1]
	})

	changelog := Changelog{
		Entries: changes,
	}

	version := "3.39.0"
	date := "2022-08-19"
	config := MustParseConfig(t, []byte(`
types:
  improvement: Improvements
  fix: Bug Fixes
`))

	// Put improvements and fixes first:
	buf, err := changelog.Template(config, version, date)
	assert.NoError(t, err)

	want := autogold.Want("markdown_test", `## 3.39.0 (2022-08-19)


### Improvements

- [cli/{backend,engine}] Foo bar baz
  [#9001](https://github.com/pulumi/pulumi/pulls/9001)

- [sdk/go] Make SDK go brrrr
  [#10000](https://github.com/pulumi/pulumi/pulls/10000)


### Bug Fixes

- [sdkgen/go] Fix Go SDK code generation
  [#20001](https://github.com/pulumi/pulumi/pulls/20001)

- [sdkgen/nodejs] Fix Typescript SDK code generation.
  [#20000](https://github.com/pulumi/pulumi/pulls/20000)

`)
	want.Equal(t, buf.String())
}
