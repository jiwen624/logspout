// +build !debug

package utils

import "github.com/jiwen624/logspout/log"

// CheckErr prints the error and do nothing else
func CheckErr(e error) {
	if e != nil {
		log.Error(e)
	}
}
