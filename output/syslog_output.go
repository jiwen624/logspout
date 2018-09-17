package output

type Syslog struct {
	Protocol   string `json:"protocol"`
	NetAddr    string `json:"netAddr"`
	MaxBackups int    `json:"maxBackups"`
	Tag        string `json:"tag"`
}

// TODO: Write
func (f Syslog) Write(p []byte) (n int, err error) {
	// TODO: use bufio to avoid excessive I/O
	// TODO: flush the buffer when program exits
	return 0, nil
}
