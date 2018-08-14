// generated by jsonenums -type=Type; DO NOT EDIT

package output

import (
	"encoding/json"
	"fmt"
)

var (
	_TypeNameToValue = map[string]Type{
		"console": console,
		"file":    file,
		"syslog":  syslog,
		"kafka":   kafka,
		"es":      es,
	}

	_TypeValueToName = map[Type]string{
		console: "console",
		file:    "file",
		syslog:  "syslog",
		kafka:   "kafka",
		es:      "es",
	}
)

func init() {
	var v Type
	if _, ok := interface{}(v).(fmt.Stringer); ok {
		_TypeNameToValue = map[string]Type{
			interface{}(console).(fmt.Stringer).String(): console,
			interface{}(file).(fmt.Stringer).String():    file,
			interface{}(syslog).(fmt.Stringer).String():  syslog,
			interface{}(kafka).(fmt.Stringer).String():   kafka,
			interface{}(es).(fmt.Stringer).String():      es,
		}
	}
}

// MarshalJSON is generated so Type satisfies json.Marshaler.
func (r Type) MarshalJSON() ([]byte, error) {
	if s, ok := interface{}(r).(fmt.Stringer); ok {
		return json.Marshal(s.String())
	}
	s, ok := _TypeValueToName[r]
	if !ok {
		return nil, fmt.Errorf("invalid Type: %d", r)
	}
	return json.Marshal(s)
}

// UnmarshalJSON is generated so Type satisfies json.Unmarshaler.
func (r *Type) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Type should be a string, got %s", data)
	}
	v, ok := _TypeNameToValue[s]
	if !ok {
		return fmt.Errorf("invalid Type %q", s)
	}
	*r = v
	return nil
}
