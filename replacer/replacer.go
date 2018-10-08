package replacer

// Replacer is the interface which must be implemented by a particular replacement policy.
type Replacer interface {
	// ReplacedValue returns the new replaced value. It may need a random
	// number generator provided by the caller.
	ReplacedValue(RandomGenerator) (string, error)

	// Copy returns a deep copy of the replacer
	Copy() Replacer
}

type Replacers map[string]Replacer

func (r Replacers) Copy() Replacers {
	n := Replacers{}
	for k, v := range r {
		n[k] = v.Copy()
	}
	return n
}
