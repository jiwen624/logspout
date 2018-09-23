package output

import (
	"fmt"
	"io"
	"os"
)

type Console struct {
	FileName string `json:"fileName"`
	file     io.Writer
}

func (c *Console) Write(p []byte) (n int, err error) {
	if c.file == nil {
		return 0, fmt.Errorf("output is null: %s", c)
	}
	return c.file.Write(p)
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
	switch c.FileName {
	case "stdout":
		c.file = os.Stdout
	case "stderr":
		c.file = os.Stderr
	default:
		return fmt.Errorf("invalid console type: %s", c)
	}
	return nil
}

func (c *Console) Deactivate() error {
	c.file = nil
	return nil
}
