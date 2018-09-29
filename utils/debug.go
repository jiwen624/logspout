// +build debug

package utils

import "github.com/jiwen624/logspout/log"

// ExitOnErr will print the error and panic
func CheckErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
