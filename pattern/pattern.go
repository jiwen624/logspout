package pattern

type Pattern interface {
	Matches(string) []string
	Names() []string
}
