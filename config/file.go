package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

const (
	// The maximum configuration file size in bytes
	maxConfFileSize = 1048576
)

func readFile(path string, sizeLimit int64) ([]byte, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrap(err, "read file")
	}

	size := fi.Size()
	if size >= sizeLimit {
		return nil, fmt.Errorf("larger than %d bytes", sizeLimit)
	}

	cf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read file")
	}

	return cf, nil
}

// FromJsonFile loads a config file and parse it to s SpoutConfig object
func FromJsonFile(name string) (*SpoutConfig, error) {
	bs, err := readFile(name, maxConfFileSize)
	if err != nil {
		return nil, errors.Wrap(err, "load config from file")
	}
	return loadJson(bs)
}
