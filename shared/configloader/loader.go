package configloader

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/Andrea-Cavallo/golang-modules/shared/configsource"
)

const ConfigFileEnv = "FORGECORE_CONFIG_FILE"

type Values struct {
	values map[string]string
}

func (v Values) String(key string) string {
	return v.values[key]
}

func (v Values) Bool(key string) bool {
	value, err := strconv.ParseBool(v.values[key])
	if err != nil {
		return false
	}
	return value
}

func (v Values) Int(key string) int {
	value, err := strconv.Atoi(v.values[key])
	if err != nil {
		return 0
	}
	return value
}

func (v Values) Secret(key string) configschema.Secret {
	return configschema.Secret(v.values[key])
}

type Loader struct {
	schema  configschema.Schema
	sources []configsource.Source
}

func New(schema configschema.Schema, sources ...configsource.Source) Loader {
	return Loader{schema: schema, sources: sources}
}

func NewDefault(schema configschema.Schema) Loader {
	yamlPath := os.Getenv(ConfigFileEnv)
	sources := []configsource.Source{
		configsource.NewMap(schema.Defaults()),
		configsource.NewYAMLFile(yamlPath, true),
		configsource.NewEnv(schema.Keys()),
	}
	return New(schema, sources...)
}

func (l Loader) Load(ctx context.Context) (Values, error) {
	merged := map[string]string{}
	for _, source := range l.sources {
		values, err := source.Load(ctx)
		if err != nil {
			return Values{}, fmt.Errorf("caricamento sorgente config: %w", err)
		}
		mergeValues(merged, values)
	}
	if err := l.schema.Validate(merged); err != nil {
		return Values{}, err
	}
	return Values{values: merged}, nil
}

func mergeValues(target, source map[string]string) {
	for key, value := range source {
		target[key] = value
	}
}
