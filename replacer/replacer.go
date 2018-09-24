package replacer

import "github.com/leesper/go_rng"

// Replacer is the interface which must be implemented by a particular replacement policy.
type Replacer interface {
	// ReplacedValue returns the new replaced value.
	ReplacedValue(*rng.GaussianGenerator) (string, error)
}
