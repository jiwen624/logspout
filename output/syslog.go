package output

import (
	"fmt"
	slog "log/syslog"

	"github.com/jiwen624/logspout/log"
	"github.com/pkg/errors"
)

type Syslog struct {
	Protocol string `json:"protocol"`
	NetAddr  string `json:"netAddr"`
	Tag      string `json:"tag"`
	logger   *slog.Writer
}

func (s *Syslog) String() string {
	return fmt.Sprintf("Syslog{Protocl:%s,NetAddr:%s,Tag:%s}",
		s.Protocol, s.NetAddr, s.Tag)
}

func (s *Syslog) Write(p []byte) (n int, err error) {
	if s.logger == nil {
		return 0, fmt.Errorf("output is null: %s", s)
	}
	return s.logger.Write(p)
}

func (s *Syslog) ID() ID {
	return id(s.String())
}

func (s *Syslog) Type() Type {
	return syslog
}

func (s *Syslog) Activate() error {
	o := fmt.Sprintf("%s://%s", s.Protocol, s.NetAddr)
	log.Infof("Activating output %s", o)

	if err := s.buildSyslog(); err != nil {
		return errors.Wrap(err, "activate syslog")
	}
	return nil
}

func (s *Syslog) Deactivate() error {
	o := fmt.Sprintf("%s//%s", s.Protocol, s.NetAddr)
	log.Infof("Deactivating output %s", o)

	return errors.Wrap(s.logger.Close(), "deactivate syslog")
}

// buildSyslog extracts output parameters from the config file and build the
// syslog output
func (s *Syslog) buildSyslog() error {
	var (
		protocol = "udp"
		netaddr  = "localhost:514"
		level    = slog.LOG_INFO
		tag      = "logspout"
	)

	if s.Protocol != "" {
		protocol = s.Protocol
	}

	if s.NetAddr != "" {
		netaddr = s.NetAddr
	}
	if s.Tag != "" {
		tag = s.Tag
	}
	// TODO: The syslog default level is hardcoded for now.
	w, err := slog.Dial(protocol, netaddr, level, tag)
	if err != nil {
		return errors.Wrap(err, "build syslog")
	}
	s.logger = w
	return nil
}
