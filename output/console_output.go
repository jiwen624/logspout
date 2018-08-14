package output

type Console struct {
	FileName string
}

func (c Console) Write(s string) error {
	// TODO: use bufio to avoid excessive I/O
	// TODO: flush the buffer when program exits
	return nil
}
