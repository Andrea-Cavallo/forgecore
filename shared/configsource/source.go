package configsource

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Source interface {
	Load(ctx context.Context) (map[string]string, error)
}

type Map struct {
	values map[string]string
}

func NewMap(values map[string]string) Map {
	return Map{values: values}
}

func (m Map) Load(context.Context) (map[string]string, error) {
	values := make(map[string]string, len(m.values))
	for key, value := range m.values {
		values[key] = value
	}
	return values, nil
}

type Env struct {
	keys []string
}

func NewEnv(keys []string) Env {
	return Env{keys: keys}
}

func (e Env) Load(context.Context) (map[string]string, error) {
	values := make(map[string]string, len(e.keys))
	for _, key := range e.keys {
		if value, ok := os.LookupEnv(key); ok {
			values[key] = value
		}
	}
	return values, nil
}

type YAMLFile struct {
	path     string
	optional bool
}

func NewYAMLFile(path string, optional bool) YAMLFile {
	return YAMLFile{path: path, optional: optional}
}

func (y YAMLFile) Load(context.Context) (map[string]string, error) {
	if y.path == "" {
		return map[string]string{}, nil
	}
	data, err := os.ReadFile(y.path)
	if err != nil && y.optional {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lettura YAML configurazione: %w", err)
	}
	return parseYAML(data)
}

func parseYAML(data []byte) (map[string]string, error) {
	raw := map[string]any{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing YAML configurazione: %w", err)
	}
	values := map[string]string{}
	flattenYAML("", raw, values)
	return values, nil
}

func flattenYAML(prefix string, raw map[string]any, values map[string]string) {
	for key, value := range raw {
		flatKey := joinKey(prefix, key)
		if nested, ok := value.(map[string]any); ok {
			flattenYAML(flatKey, nested, values)
			continue
		}
		values[strings.ToUpper(flatKey)] = fmt.Sprint(value)
	}
}

func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "_" + key
}
