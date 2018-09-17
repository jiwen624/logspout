package output

import (
	"encoding/json"
	"io"

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
}

// Wrapper is a wrapper struct that contains the output type and a byte slice
// which represents the configurations of that type.
type Wrapper struct {
	T   Type            `json:"type"`
	Raw json.RawMessage `json:"attrs"`
}

// outputMap is the map for the output types and their factory methods
var outputMap map[Type]func() Output

func init() {
	registerOutput()
}

// registerOutput initializes the map of output types and their corresponding
// struct instances factory methods.
//
// This function is not concurrent-safe and should only be called in a init()
// function
func registerOutput() {
	outputMap = map[Type]func() Output{
		console: func() Output { return &Console{} },
		file:    func() Output { return &File{} },
		syslog:  func() Output { return &Syslog{} },
		kafka:   func() Output { return &Kafka{} },
	}
}

// BuildOutputMap builds the outputs based on the configurations wrapped by
// Wrapper. It iterates all the configurations and creates output instances
// of various types.
func BuildOutputMap(ow map[string]Wrapper) map[string]Output {
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
	utils.PanicOnErr(json.Unmarshal(m.Raw, op))

	return op
}
