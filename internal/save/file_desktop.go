//go:build !(js && wasm)

package save

import (
	"os"
	"path/filepath"
)

func path(key string) string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "particle-accelerator", key+".json")
}

func Write(key, value string) {
	p := path(key)
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(value), 0o644)
}

func Read(key string) (string, bool) {
	b, err := os.ReadFile(path(key))
	if err != nil {
		return "", false
	}
	return string(b), true
}
