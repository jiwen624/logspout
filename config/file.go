package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/jiwen624/logspout/flag"
)

const (
	// The maximum configuration file size in bytes
	maxConfFileSize = 1048576
)

func readFile(path string, sizeLimit int64) ([]byte, error) {
	fi, e := os.Stat(path)
	if e != nil {
		return nil, fmt.Errorf("the file doesn't exist: %s", path)
	}

	size := fi.Size()
	if size >= sizeLimit {
		return nil, fmt.Errorf("the file is too big: > %d bytes", maxConfFileSize)
	}

	cf, err := ioutil.ReadFile(flag.ConfigPath)
	if err != nil {
		return nil, err
	}

	return cf, nil
}

// FromJsonFile loads a config file and parse it to s SpoutConfig object
func FromJsonFile(name string) (*SpoutConfig, error) {
	bs, err := readFile(name, maxConfFileSize)
	if err != nil {
		return nil, errors.Wrap(err, "Load config from file:")
	}
	return loadJson(bs)
}
