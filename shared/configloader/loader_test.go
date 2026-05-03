package configloader

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
)

const testKey = "FORGECORE_TEST_PORT"

func TestDefaultYAMLEnvPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte("forgecore_test_port: ':9000'\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("scrittura yaml: %v", err)
	}
	t.Setenv(ConfigFileEnv, path)
	t.Setenv(testKey, ":9100")

	schema := configschema.Schema{
		{Key: testKey, Default: ":8080", Kind: configschema.String},
	}
	values, err := NewDefault(schema).Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := values.String(testKey); got != ":9100" {
		t.Fatalf("precedenza errata: %s", got)
	}
}

func TestInvalidBool(t *testing.T) {
	const key = "FORGECORE_TEST_BOOL"
	t.Setenv(key, "forse")
	schema := configschema.Schema{{Key: key, Kind: configschema.Bool}}
	if _, err := NewDefault(schema).Load(context.Background()); err == nil {
		t.Fatal("atteso errore per bool non valido")
	}
}
