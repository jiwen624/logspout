package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	// The maximum configuration file size in bytes
	maxConfFileSize = 1048576
)

const (
	extJson = ".json"
	extYml  = ".yml"
	extYaml = ".yaml"
)

// config file errors
var (
	errUnsupportedFileType = errors.New("unsupported file extension")
	errFileTooLarge        = errors.New("file too large")
)

func readFile(path string, sizeLimit int64) ([]byte, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrap(err, "read file")
	}

	size := fi.Size()
	if size >= sizeLimit {
		m := fmt.Sprintf("size: %d, expect: <=%d", size, sizeLimit)
		return nil, errors.Wrap(errFileTooLarge, m)
	}

	cf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read file")
	}

	return cf, nil
}

// FromJsonFile loads a json config file and parse it to s SpoutConfig object
func FromJsonFile(name string) (*SpoutConfig, error) {
	bs, err := readFile(name, maxConfFileSize)
	if err != nil {
		return nil, errors.Wrap(err, "from json")
	}
	return loadJson(bs)
}

// FromYamlFile loads a yaml config file and parse it to s SpoutConfig object
func FromYamlFile(name string) (*SpoutConfig, error) {
	return nil, errors.Wrap(errUnsupportedFileType, "Yaml")
}

// FromFile build a SpoutConfig object from a file. It calls different builders
// based on the file type.
func FromFile(path string) (*SpoutConfig, error) {
	switch ext := filepath.Ext(path); ext {
	case extJson:
		return FromJsonFile(path)
	case extYml, extYaml:
		return FromYamlFile(path)
	default:
		return nil, errors.Wrap(errUnsupportedFileType, ext)
	}
}
