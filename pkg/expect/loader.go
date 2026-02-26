package expect

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func unmarshalYAML(data []byte) (expectFile, error) {
	var f expectFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return expectFile{}, fmt.Errorf("go-expect: parse yaml: %w", err)
	}
	return f, nil
}

func unmarshalJSON(data []byte) (expectFile, error) {
	var f expectFile
	if err := json.Unmarshal(data, &f); err != nil {
		return expectFile{}, fmt.Errorf("go-expect: parse json: %w", err)
	}
	return f, nil
}

// LoadYAML parses YAML bytes and returns a Suite ready to run.
func LoadYAML(data []byte) (*Suite, error) {
	f, err := unmarshalYAML(data)
	if err != nil {
		return nil, err
	}
	return buildSuite([]expectFile{f})
}

// LoadJSON parses JSON bytes and returns a Suite ready to run.
func LoadJSON(data []byte) (*Suite, error) {
	f, err := unmarshalJSON(data)
	if err != nil {
		return nil, err
	}
	return buildSuite([]expectFile{f})
}

// LoadFile parses a YAML or JSON file from the OS filesystem, detected by extension.
func LoadFile(fpath string) (*Suite, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("go-expect: read file %q: %w", fpath, err)
	}
	if filepath.Ext(fpath) == ".json" {
		return LoadJSON(data)
	}
	return LoadYAML(data)
}

// LoadDir loads all *.yaml, *.yml, and *.json files in dir from the OS filesystem.
func LoadDir(dir string) (*Suite, error) {
	var files []expectFile
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(p)
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("go-expect: read %q: %w", p, err)
		}
		var f expectFile
		if ext == ".json" {
			f, err = unmarshalJSON(data)
		} else {
			f, err = unmarshalYAML(data)
		}
		if err != nil {
			return err
		}
		files = append(files, f)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return buildSuite(files)
}

// LoadFS loads all *.yaml, *.yml, and *.json files from fsys and returns a Suite ready to run.
// Useful with //go:embed directories. Use fs.Sub to scope to a subdirectory if needed.
func LoadFS(fsys fs.FS) (*Suite, error) {
	var files []expectFile
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := path.Ext(p)
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("go-expect: read %q: %w", p, err)
		}
		var f expectFile
		if ext == ".json" {
			f, err = unmarshalJSON(data)
		} else {
			f, err = unmarshalYAML(data)
		}
		if err != nil {
			return err
		}
		files = append(files, f)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return buildSuite(files)
}
