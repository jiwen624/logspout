package output

import (
	"sync"

	"github.com/pkg/errors"
)

var where struct {
	// Where the data is written to. There may be multiple destinations for a
	// specific output type.
	m map[Type]map[ID]Output
	// The mutex to protect the global map above
	sync.RWMutex
	// Make sure the global registry is initialized only once
	sync.Once
}

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
	where.Lock()
	defer where.Unlock()

	if err := output.Activate(); err != nil {
		return errors.Wrap(err, "register failed:")
	}

	where.Do(func() {
		where.m = make(map[Type]map[ID]Output)
	})

	typ := output.Type()
	tm, ok := where.m[typ]
	if !ok {
		tm = make(map[ID]Output, 1)
		where.m[typ] = tm
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
	where.Lock()
	defer where.Unlock()

	if err := output.Deactivate(); err != nil {
		return errors.Wrap(err, "unregister failed:")
	}

	tm, ok := where.m[output.Type()]
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
	where.RLock()
	defer where.RUnlock()

	var errs []error
	for _, tm := range where.m {
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

// ForOne applies the operation to the specified output
func ForOne(apply Apply, typ Type, id ID) error {
	where.RLock()
	defer where.RUnlock()

	if tm, ok := where.m[typ]; ok {
		if o, ok := tm[id]; ok {
			return apply(o)
		}
	}
	return ErrNotFound
}

// Write is a helper function to write the string to all the outputs
func Write(str string) []error {
	return ForAll(func(o Output) error {
		_, err := o.Write([]byte(str))
		return err
	})
}
