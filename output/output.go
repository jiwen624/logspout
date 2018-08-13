package output

type Output interface {
	Write(string) error
}
