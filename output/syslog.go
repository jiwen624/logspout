package output

// For output-syslog
const (
	PROTOCOL  = "protocol"
	NETADDR   = "netaddr"
	SYSLOGTAG = "tag"
)

type Syslog struct {
	Protocol   string `json:"protocol"`
	NetAddr    string `json:"netAddr"`
	MaxBackups int    `json:"maxBackups"`
	Tag        string `json:"tag"`
}

func (s *Syslog) String() string {
	return "Syslog"
}

// TODO: Write
func (s *Syslog) Write(p []byte) (n int, err error) {
	// TODO: use bufio to avoid excessive I/O
	// TODO: flush the buffer when program exits
	// fmt.Println("a dumb placeholder for syslog output")
	return 0, nil
}

func (s *Syslog) ID() ID {
	return ID("") // TODO
}

func (s *Syslog) Type() Type {
	return syslog
}

func (s *Syslog) Activate() error {
	// TODO
	return nil
}

func (s *Syslog) Deactivate() error {
	// TODO
	return nil
}
