package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/utils"
)

type UnsupportedConsoleTypeError string

func (e UnsupportedConsoleTypeError) Error() string {
	return fmt.Sprintf("Unsupported console type: %s", e)
}

type Console struct {
	FileName string `json:"defaultFileName"`
	logger   io.Writer
}

// console types
var supported = []string{
	"stdout",
	"stderr",
}

func (c *Console) Write(p []byte) (n int, err error) {
	if c.logger == nil {
		return 0, errors.Wrap(errOutputNull, c.String())
	}
	return c.logger.Write(p)
}

func (c *Console) String() string {
	return fmt.Sprintf("Console{FileName:%s}", c.FileName)
}

func (c *Console) ID() ID {
	return id(c.String())
}

func (c *Console) Type() Type {
	return console
}

func (c *Console) Activate() error {
	log.Infof("Activating output %s", c.FileName)

	normalized := normalizeName(c.FileName)
	switch normalized {
	case "stdout":
		c.logger = os.Stdout
	case "stderr":
		c.logger = os.Stderr
	default:
		return UnsupportedConsoleTypeError(c.String())
	}
	c.FileName = normalized
	return nil
}

func (c *Console) Deactivate() error {
	if c.logger == nil {
		return errors.Wrap(errOutputNull, c.String())
	}

	log.Infof("Deactivating output %s", c.FileName)

	c.logger = nil
	return nil
}

// normalizeName tries its best to normalize the console name, it returns
// the original name if fails to do that.
func normalizeName(c string) string {
	n := strings.ToLower(strings.TrimSpace(c))
	if utils.StrIndex(supported, n) >= 0 {
		return n
	}
	return c
}
