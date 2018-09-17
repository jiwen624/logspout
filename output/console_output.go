package output

type Console struct {
	FileName string
}

func (c Console) Write(p []byte) (n int, err error) {
	// TODO: use bufio to avoid excessive I/O
	// TODO: flush the buffer when program exits
	return 0, nil // TODO
}
