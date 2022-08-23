package changelog

import (
	"fmt"
	"strings"
)

type Changelog struct {
	// One or more change entries
	Entries []*Entry `yaml:"changes,omitempty"`
}

func (cl *Changelog) Merge(other Changelog) *Changelog {
	var changes []*Entry
	changes = append(changes, cl.Entries...)
	changes = append(changes, other.Entries...)
	return &Changelog{
		Entries: changes,
	}
}

// Entry records a change by its category, the scope of the change, and metadata from GitHub needed to render a
// changelog.
//
// Follows the conventional commits specification: https://www.conventionalcommits.org/en/v1.0.0/#specification
type Entry struct {
	// Type is a noun describing the category of change: improvement, fix, and so on.
	Type string `yaml:"type,omitempty"`
	// Scope is a noun describing the part of the system improved, such as: cli, codegen, sdk, automation. Subscopes can
	// be represented with a slash after a scope, as in: cli/about, cli/display, auto/go, codegen/dotnet, sdk/nodejs.
	Scope Scope `yaml:"scope,omitempty"`
	// Description is a description of the change.
	Description string `yaml:"description,omitempty"`
	// Metadata contains additional, optional data for rendering the changelog, which may be set to override inferred
	// values.
	GitHubMeta GitHubMeta `yaml:"github,inline"`
}

type GitHubMeta struct {
	// PullRequestNumbers are the GitHub Pull Requests used to implementing this change, typically a single PR.
	PullRequestNumbers []int `yaml:"prs,omitempty"`
}

func (v *Entry) Conventional() string {
	var message string

	typ := v.Type
	message += fmt.Sprintf("%v(%v): %v.", typ, v.Scope.String(), v.Description)

	var prs []string
	for _, prNumber := range v.GitHubMeta.PullRequestNumbers {
		prs = append(prs, fmt.Sprintf("[#%v](https://github.com/pulumi/pulumi/pull/%v)", prNumber, prNumber))
	}

	if len(prs) > 0 {
		message += " " + strings.Join(prs, ", ")
	}

	return message
}
