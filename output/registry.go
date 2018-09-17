package output

import (
	"sync"

	"github.com/pkg/errors"
)

var (
	// Where the data is written to. There may be multiple destinations for a
	// specific output type.
	where map[Type]map[ID]Output
	// The mutex to protect the global map above
	whereMu sync.RWMutex
	// Make sure the global registry is initialized only once
	once sync.Once
)

var (
	ErrDuplicate = errors.New("duplicate ID found")
	ErrNotFound  = errors.New("output not found")
)

// The operation applies to the output
type Apply func(Output) error

// The predicate that filters the ouput
type Predicate func(Output) bool

// Register registers an output destination
func Register(output Output) error {
	whereMu.Lock()
	defer whereMu.Unlock()

	if err := output.Activate(); err != nil {
		return errors.Wrap(err, "register failed:")
	}

	once.Do(func() {
		where = make(map[Type]map[ID]Output)
	})

	typ := output.Type()
	tm, ok := where[typ]
	if !ok {
		tm = make(map[ID]Output, 1)
		where[typ] = tm
	}

	id := output.ID()
	if _, ok := tm[id]; ok {
		return ErrDuplicate
	}
	tm[id] = output
	return nil
}

// Unregister unregisters an output from the global registry
func Unregister(output Output) error {
	whereMu.Lock()
	defer whereMu.Unlock()

	if err := output.Deactivate(); err != nil {
		return errors.Wrap(err, "unregister failed:")
	}

	tm, ok := where[output.Type()]
	if !ok {
		return ErrNotFound
	}

	id := output.ID()
	if _, ok := tm[id]; ok {
		return ErrNotFound
	}
	delete(tm, id)
	return nil
}

// ForEach applies the operation to each output which matches the predicate
func ForEach(apply Apply, predicate Predicate) []error {
	// TODO: This function may hold the lock for too long
	whereMu.RLock()
	defer whereMu.RUnlock()

	var errs []error
	for _, tm := range where {
		for _, o := range tm {
			if !predicate(o) {
				continue
			}
			if err := apply(o); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// ForAll applies the operation to all the outputs
func ForAll(apply Apply) []error {
	return ForEach(apply, func(Output) bool {
		return true
	})
}

// Do applies the operation to the specified output
func Do(apply Apply, typ Type, id ID) error {
	whereMu.RLock()
	defer whereMu.RUnlock()

	if tm, ok := where[typ]; ok {
		if o, ok := tm[id]; ok {
			return apply(o)
		}
	}
	return ErrNotFound
}
