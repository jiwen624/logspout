//go:generate go run codebot.go
package spout

var version string

func Version() string {
	return version
}
