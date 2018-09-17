package output

type Syslog struct {
	Protocol   string `json:"protocol"`
	NetAddr    string `json:"netAddr"`
	MaxBackups int    `json:"maxBackups"`
	Tag        string `json:"tag"`
}

// TODO: Write
func (f *Syslog) Write(p []byte) (n int, err error) {
	// TODO: use bufio to avoid excessive I/O
	// TODO: flush the buffer when program exits
	return 0, nil
}

func (f *Syslog) ID() ID {
	return ID("") // TODO
}

func (f *Syslog) Type() Type {
	return syslog
}

func (f *Syslog) Activate() error {
	// TODO
	return nil
}

func (f *Syslog) Deactivate() error {
	// TODO
	return nil
}
