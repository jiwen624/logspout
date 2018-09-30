package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/jiwen624/logspout/log"
	"github.com/pkg/errors"
)

// ExitOnErr prints the error and exits immediately
func ExitOnErr(wrapper string, e error) {
	if e != nil {
		log.Error(errors.Wrap(e, wrapper))
		os.Exit(1)
	}
}

// CombineErrs combines multiple errors
func CombineErrs(errs []error) error {
	var cmb []string
	for _, err := range errs {
		if err == nil {
			continue
		}
		cmb = append(cmb, err.Error())
	}
	if len(cmb) == 0 {
		return nil
	} else {
		return fmt.Errorf(strings.Join(cmb, "\n"))
	}
}
