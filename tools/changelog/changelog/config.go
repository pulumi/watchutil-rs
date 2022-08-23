package changelog

import (
	"fmt"
	"text/template"

	"github.com/goccy/go-yaml"
)

type Config struct {
	// Types is a map of short types to descriptive names to be used in rendering.
	Types ConfigTypes `yaml:"types"`

	// Scopes is a map of valid scopes to subscopes
	Scopes ConfigScopes `yaml:"scopes"`

	// Template is the template to use for rendering the changelog.
	Template RenderTemplate `yaml:"template"`
}

type ConfigTypes = OrderedMap[string, string]
type ConfigScopes = OrderedMap[string, []string]

type OrderedMap[K comparable, V any] struct {
	Items yaml.MapSlice `yaml:",inline"`
}

func (c *OrderedMap[K, V]) UnmarshalYAML(unmarshal func(any) error) error {
	// Validate the type of the map:
	var validMap map[K]V
	if err := unmarshal(&validMap); err != nil {
		return err
	}

	var actualMap yaml.MapSlice
	if err := unmarshal(&actualMap); err != nil {
		return err
	}

	var resultMap yaml.MapSlice
	for _, v := range actualMap {
		resultMap = append(resultMap, yaml.MapItem{
			Key:   v.Key.(K),
			Value: validMap[v.Key.(K)],
		})
	}
	c.Items = resultMap

	return nil
}

func (c *OrderedMap[K, V]) Keys() []string {
	keys := []string{}
	for _, v := range c.Items {
		keys = append(keys, v.Key.(string))
	}
	return keys
}

func (c *OrderedMap[K, V]) Get(key K) (V, bool) {
	for _, v := range c.Items {
		if v.Key.(K) == key {
			return v.Value.(V), true
		}
	}
	var defaultValue V
	return defaultValue, false
}

func (c *OrderedMap[K, V]) ToMap() map[K]V {
	out := make(map[K]V)
	for _, v := range c.Items {
		out[v.Key.(K)] = v.Value.(V)
	}
	return out
}

type RenderTemplate struct {
	*template.Template
}

func (t *RenderTemplate) UnmarshalYAML(unmarshal func(any) error) error {
	var text string
	if err := unmarshal(&text); err != nil {
		return err
	}

	tmpl, err := NewTemplate(text)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	t.Template = tmpl

	return nil
}
