//go:generate go run codebot.go
package spout

var (
	// the version
	version string
	// the commit short hash via `git rev-parse --short HEAD`
	commit string
)

func Version() string {
	return version
}

func Commit() string {
	return commit
}
