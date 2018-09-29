// +build !debug

package utils

import (
	"os"

	"github.com/jiwen624/logspout/log"
	"github.com/pkg/errors"
)

// ExitOnErr prints the error and do nothing else
func ExitOnErr(wrapper string, e error) {
	if e != nil {
		log.Error(errors.Wrap(e, wrapper))
		os.Exit(1)
	}
}
