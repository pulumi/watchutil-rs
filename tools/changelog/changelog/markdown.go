package changelog

import (
	"bytes"
	"html/template"
	"sort"
	"strings"
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
var markdownTemplate, _ = template.New("markdown").Funcs(funcs).Parse(markdownTemplateText)

type TemplateInputs struct {
	Version string
	Date    string
	Groups  []Group
	Titles  map[string]string
}

type Group struct {
	Type    string
	Changes []*Change
}

func (c *Changelog) Markdown(config Config, version string, date string) (*bytes.Buffer, error) {
	grouped := make(map[string][]*Change)
	for _, c := range c.Changes {
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

	var groups []Group

	for _, groupName := range config.Types.Keys() {
		groupName := groupName
		if _, has := grouped[groupName]; has {
			groups = append(groups, takeGroup(grouped, groupName))
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err := markdownTemplate.Execute(buf, TemplateInputs{
		Version: version,
		Date:    date,
		Groups:  groups,
		Titles:  config.Types.ToMap(),
	})

	return buf, err
}

func takeGroup(grouped map[string][]*Change, group string) Group {
	improvements := grouped[group]
	delete(grouped, group)
	return Group{
		Type:    group,
		Changes: improvements,
	}
}
