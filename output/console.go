package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jiwen624/logspout/utils"
)

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
		return 0, fmt.Errorf("output is null: %s", c)
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
	normalized := normalizeName(c.FileName)
	switch normalized {
	case "stdout":
		c.logger = os.Stdout
	case "stderr":
		c.logger = os.Stderr
	default:
		return fmt.Errorf("invalid console type: %s", c)
	}
	c.FileName = normalized
	return nil
}

func (c *Console) Deactivate() error {
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
