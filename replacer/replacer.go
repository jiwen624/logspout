// Package replacer defines the interface and the implementations of replacers.
// The replacers are responsible of generating new values based on pre-defined
// rules. The generated values will be concatenated into an entry of machine log
// which is then sent to the output destinations.
package replacer

// Replacer should be implemented by a replacement policy.
type Replacer interface {
	// ReplacedValue returns the new replaced value. It may need a random
	// number generator provided by the caller.
	ReplacedValue(RandomGenerator) (string, error)

	// Copy returns a deep copy of the replacer
	Copy() Replacer
}

type Replacers map[string]Replacer

func (r *Replacers) Copy() Replacers {
	n := Replacers{}
	for k, v := range *r {
		n[k] = v.Copy()
	}
	return n
}
