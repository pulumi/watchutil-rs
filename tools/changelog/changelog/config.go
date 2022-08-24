package changelog

import (
	"github.com/goccy/go-yaml"
)

type Config struct {
	// Mapping of short names to long names.
	Types ConfigTypes `yaml:"types"`

	// List of valid scopes and subscopes
	Scopes ConfigScopes `yaml:"scopes"`
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

	c.Items = actualMap

	return nil
}

func (c OrderedMap[K, V]) Keys() []string {
	keys := []string{}
	for _, v := range c.Items {
		keys = append(keys, v.Key.(string))
	}
	return keys
}

func (c OrderedMap[K, V]) Get(key K) (V, bool) {
	for _, v := range c.Items {
		if v.Key.(K) == key {
			return v.Value.(V), true
		}
	}
	var defaultValue V
	return defaultValue, false
}

func (c OrderedMap[K, V]) ToMap() map[K]V {
	out := make(map[K]V)
	for _, v := range c.Items {
		out[v.Key.(K)] = v.Value.(V)
	}
	return out
}
