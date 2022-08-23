package changelog

import (
	"bytes"
	"sort"
	"strings"
	"text/template"
)

var markdownTemplateText = `
{{- /* Sub-template for rendering scope of change, e.g.: cli/{backend,engine} */ -}}

{{- define "scope" -}}
{{- if .Primary -}}
	[
	{{- .Primary -}}
		{{- if eq (len .SubScopes) 0 }}{{ end -}}
		{{- if eq (len .SubScopes) 1 }}{{ "/" }}{{ index .SubScopes 0 }}{{ end -}}
		{{- if ge (len .SubScopes) 2 }}{{ "/{" }}{{ join .SubScopes "," }}{{ "}" }}{{ end -}}
	] {{ end -}}
{{- end -}}

{{- /* Begin main template */ -}}

## {{ .Version }} ({{ .Date }})

{{ range $group := .Groups }}
### {{ index $.Titles $group.Type }}
{{ range $change := $group.Changes }}
- {{ template "scope" $change.Scope }}{{ $change.Description }}
{{- if $change.GitHubMeta.PullRequestNumbers }}
{{- range $pr := $change.GitHubMeta.PullRequestNumbers }}
  [#{{ $pr }}](https://github.com/pulumi/pulumi/pulls/{{ $pr }})
{{- end }}
{{- end }}
{{ end }}
{{ end -}}
`

var funcs = template.FuncMap{"join": strings.Join}
var DefaultTemplate = template.Must(template.New("markdown").Funcs(funcs).Parse(markdownTemplateText))

func NewTemplate(text string) (*template.Template, error) {
	return template.New("markdown").Funcs(funcs).Parse(text)
}

type TemplateInputs struct {
	Version string
	Date    string
	Groups  []Group
	Titles  map[string]string
}

type Group struct {
	Type    string
	Changes []*Entry
}

func (c *Changelog) Template(config Config, version string, date string) (*bytes.Buffer, error) {
	grouped := GroupEntries(c)
	var groups []Group

	for _, groupName := range config.Types.Keys() {
		groupName := groupName
		if _, has := grouped[groupName]; has {
			groups = append(groups, takeGroup(grouped, groupName))
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err := config.Template.Execute(buf, TemplateInputs{
		Version: version,
		Date:    date,
		Groups:  groups,
		Titles:  config.Types.ToMap(),
	})

	return buf, err
}

// GroupEntries takes a changelog and returns a map of entry types to changes.
func GroupEntries(c *Changelog) map[string][]*Entry {
	grouped := make(map[string][]*Entry)
	for _, c := range c.Entries {
		sort.Strings(c.Scope.SubScopes)
		grouped[c.Type] = append(grouped[c.Type], c)
	}

	for _, v := range grouped {
		sort.SliceStable(v, func(i, j int) bool {
			first := v[i]
			second := v[j]
			if strings.Compare(first.Scope.Primary, second.Scope.Primary) < 0 {
				return true
			}

			c1Subs := strings.Join(first.Scope.SubScopes, ",")
			c2Subs := strings.Join(second.Scope.SubScopes, ",")

			return strings.Compare(c1Subs, c2Subs) < 0
		})
	}
	return grouped
}

func takeGroup(grouped map[string][]*Entry, group string) Group {
	improvements := grouped[group]
	delete(grouped, group)
	return Group{
		Type:    group,
		Changes: improvements,
	}
}
