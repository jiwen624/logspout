package spout

type Jitter interface {
	// AddJitter accepts a value and returns a new value with a jitter applied.
	AddJitter(int) int
}

type interTransJitter struct{}

func InterTransJitter() *interTransJitter {
	return &interTransJitter{}
}

func (j *interTransJitter) AddJitter(val int) int {
	// TODO
	return val
}

type intraTransJitter struct{}

func IntraTransJitter() *intraTransJitter {
	return &intraTransJitter{}
}

func (j *intraTransJitter) AddJitter(val int) int {
	// TODO
	return val
}
