package output

import (
	"fmt"
	"sync"

	"github.com/jiwen624/logspout/utils"

	"github.com/pkg/errors"
)

// Registry is a collection of outputs. Output can be registered to or unregistered
// from the registry. It can also accepts byte slice and write it to the selected
// output based on a filter.
type Registry struct {
	// Where the data is written to. There may be multiple destinations for a
	// specific output type.
	m map[Type]map[ID]Output
	// The mutex to protect the global map above
	sync.RWMutex
	// Make sure the global registry is initialized only once
	sync.Once
}

var (
	ErrDuplicate     = errors.New("duplicate ID found")
	ErrNotFound      = errors.New("output not found")
	ErrEmptyRegistry = errors.New("registry is empty")
)

func (r *Registry) Size() int {
	r.Lock()
	defer r.Unlock()
	return len(r.m)
}

func (r *Registry) String() string {
	r.Lock()
	defer r.Unlock()
	return fmt.Sprintf("%+v", r.m)

}

// Register registers an output destination
func (r *Registry) Register(output Output) error {
	r.Lock()
	defer r.Unlock()

	if err := output.Activate(); err != nil {
		return errors.Wrap(err, "register failed")
	}

	r.Do(func() {
		r.m = make(map[Type]map[ID]Output)
	})

	typ := output.Type()
	tm, ok := r.m[typ]
	if !ok {
		tm = make(map[ID]Output, 1)
		r.m[typ] = tm
	}

	id := output.ID()
	if _, ok := tm[id]; ok {
		return ErrDuplicate
	}
	tm[id] = output
	return nil
}

// Unregister unregisters an output from the global registry
func (r *Registry) Unregister(output Output) error {
	r.Lock()
	defer r.Unlock()

	if err := output.Deactivate(); err != nil {
		return errors.Wrap(err, "unregister failed:")
	}

	tm, ok := r.m[output.Type()]
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

// Get accepts and output ID and returns the output object
func (r *Registry) Get(id ID) (output Output, err error) {
	err = r.ForOne(func(o Output) error {
		output = o
		return nil
	}, unspecified, id)
	return
}

// ForEach applies the operation to each output which matches the predicate
func (r *Registry) ForEach(apply Apply, predicate Predicate) error {
	r.RLock()
	defer r.RUnlock()

	if r.m == nil {
		return ErrEmptyRegistry
	}

	var errs []error
	for _, tm := range r.m {
		for _, o := range tm {
			if !predicate(o) {
				continue
			}
			if err := apply(o); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return utils.CombineErrs(errs)
}

// ForAll applies the operation to all the outputs
func (r *Registry) ForAll(apply Apply) error {
	return r.ForEach(apply, func(Output) bool {
		return true
	})
}

// ForOne applies the operation to the specified output
func (r *Registry) ForOne(apply Apply, typ Type, id ID) error {
	r.RLock()
	defer r.RUnlock()

	if r.m == nil {
		return ErrEmptyRegistry
	}

	var typs []Type
	if typ == unspecified {
		typs = Types()
	} else {
		typs = append(typs, typ)
	}

	for _, tp := range typs {
		if tm, ok := r.m[tp]; ok {
			if o, ok := tm[id]; ok {
				return apply(o)
			}
		}
	}
	return ErrNotFound
}

// Write is a helper function to write the string to all the outputs
func (r *Registry) Write(str string) error {
	return r.ForAll(func(o Output) error {
		_, err := o.Write([]byte(str))
		return err
	})
}

func NewRegistry() *Registry {
	return &Registry{}
}

// RegistryFromConf creates and registers outputs from a JSON-based configuration
// byte slice.
func RegistryFromConf(ow map[string]Wrapper) (*Registry, error) {
	om := buildOutputMap(ow)
	r := &Registry{}

	var errs []error
	for _, o := range om {
		if err := r.Register(o); err != nil {
			errs = append(errs, err)
		}
	}

	return r, utils.CombineErrs(errs)
}
