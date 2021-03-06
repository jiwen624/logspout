// Package output contains the interface and implementations of the output. The
// output provides a Write method to send the generated logs to the destination.
// It can also be activated and deactivated.
package output

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"sync"

	"github.com/pkg/errors"

	"github.com/jiwen624/logspout/utils"
)

// ID is the short identity of an output, which is usually calculated from the
// filename, IP+port, etc.
type ID string

// Output is the interface defines the operations an output can perform. All the
// output destinations must implement the methods defined here in order to be
// accepted by the spout.
type Output interface {
	io.Writer
	// ID returns the short ID of the output destination
	ID() ID
	// Type returns the type of this output
	Type() Type
	// String defines the string representation of the output
	String() string
	// Activate enables the output
	Activate() error
	// Deactivate disables the output and releases the resources
	Deactivate() error
}

// Wrapper is a wrapper struct that contains the output type and a byte slice
// which represents the configurations of that type.
type Wrapper struct {
	T   Type            `json:"type"`
	Raw json.RawMessage `json:"attrs"`
}

// ClosableWriter defines a writer who also can be closed.
type ClosableWriter interface {
	io.Writer
	Close() error
}

// DumbClosableWriter is used for test only. It never returns error
type DumbClosableWriter struct{}

// Write implements the Write method of interface ClosableWriter
func (d DumbClosableWriter) Write(p []byte) (n int, err error) { return len(p), nil }

// Close implements the Close method of the interface ClosableWriter
func (d DumbClosableWriter) Close() error { return nil }

// errors
var (
	errOutputNull = errors.New("output is null")
)

// initializers is the map for the output types and their factory methods
var (
	initializers map[Type]Initializer
	mu           sync.Mutex
)

func init() {
	setupInitializers()
}

// setupInitializers initializes the map of output types and their corresponding
// struct instances factory methods.
//
// This function is not concurrent-safe and should only be called in a init()
// function
func setupInitializers() {
	mu.Lock()
	defer mu.Unlock()
	initializers = map[Type]Initializer{
		console: func() Output { return &Console{} },
		file:    func() Output { return &File{} },
		syslog:  func() Output { return &Syslog{} },
		kafka:   func() Output { return &Kafka{} },
		discard: func() Output { return &Discard{} },
	}
}

// RegisterType registers a new output type initializer. A new initializer will
// override the old one with the same type.
func RegisterType(t Type, init Initializer) {
	mu.Lock()
	defer mu.Unlock()
	initializers[t] = init
}

// UnregisterType unregisters a specific type's initializer.
func UnregisterType(t Type) {
	mu.Lock()
	defer mu.Unlock()
	delete(initializers, t)
}

func GetInitializer(t Type) (Initializer, bool) {
	mu.Lock()
	defer mu.Unlock()
	i, ok := initializers[t]
	return i, ok
}

// buildOutputMap builds the outputs based on the configurations wrapped by
// Wrapper. It iterates all the configurations and creates output instances
// of various types.
func buildOutputMap(ow map[string]Wrapper) map[string]Output {
	om := map[string]Output{}
	for k, v := range ow {
		ov := build(v)
		om[k] = ov
	}
	return om
}

// buildFile builds a single output instance based on the wrapper.
func build(m Wrapper) Output {
	op := initializers[m.T]()
	utils.ExitOnErr("build", json.Unmarshal(m.Raw, op))

	return op
}

// id creates a short checksum of the string
func id(s string) ID {
	h := sha1.New()
	h.Write([]byte(s))
	id := h.Sum(nil)
	return ID(base64.StdEncoding.EncodeToString(id))
}

// Initializer defines the initializer type which creates an empty output object
type Initializer func() Output

// The operation applies to the output
type Apply func(Output) error

// The predicate that filters the output
type Predicate func(Output) bool
