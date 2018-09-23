package output

import (
	"crypto/sha1"
	"encoding/json"
	"io"
	"sync"

	"github.com/jiwen624/logspout/utils"
)

// ID is the short identity of an output, which is usually calculated from the
// filename, IP+port, etc.
type ID string

// Output is the interface defines the operations an output can perform. All the
// output destinations must implement the methods defined here in order to be
// accpeted by the spout.
type Output interface {
	io.Writer
	// ID returns the short ID of the output destination
	ID() ID // TODO: []byte or string? md5 or sha1?
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

// outputMap is the map for the output types and their factory methods
var (
	outputMap map[Type]Initializer
	mu        sync.Mutex
)

func init() {
	initDefaultTypes()
}

// initDefaultTypes initializes the map of output types and their corresponding
// struct instances factory methods.
//
// This function is not concurrent-safe and should only be called in a init()
// function
func initDefaultTypes() {
	mu.Lock()
	defer mu.Unlock()
	outputMap = map[Type]Initializer{
		console: func() Output { return &Console{} },
		file:    func() Output { return &File{} },
		syslog:  func() Output { return &Syslog{} },
		kafka:   func() Output { return &Kafka{} },
	}
}

// RegisterType registers a new output type initializer. A new initializer will
// override the old one with the same type.
func RegisterType(t Type, init Initializer) {
	mu.Lock()
	defer mu.Unlock()
	outputMap[t] = init
}

// UnregisterType unregisters a specific type's initializer.
func UnregisterType(t Type) {
	mu.Lock()
	defer mu.Unlock()
	delete(outputMap, t)
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

// build builds a single output instance based on the wrapper.
func build(m Wrapper) Output {
	op := outputMap[m.T]()
	utils.CheckErr(json.Unmarshal(m.Raw, op))

	return op
}

func id(s string) ID {
	h := sha1.New()
	h.Write([]byte(s))
	id := h.Sum(nil)
	return ID(id)
}

// Initializer defines the initializer type which creates an empty output object
type Initializer func() Output

// The operation applies to the output
type Apply func(Output) error

// The predicate that filters the ouput
type Predicate func(Output) bool
