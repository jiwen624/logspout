// +build !debug

package utils

// PanicOnErr is noop in release mode
func PanicOnErr(e error) {}
