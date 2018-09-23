package output

import "fmt"

type Console struct {
	FileName string
}

func (c *Console) Write(p []byte) (n int, err error) {
	// TODO: use bufio to avoid excessive I/O
	// TODO: flush the buffer when program exits
	fmt.Println("a placeholder for console output")

	return 0, nil // TODO
}

func (c *Console) String() string {
	return fmt.Sprintf("Console{FileName:%s}", c.FileName)
}

func (c *Console) ID() ID {
	return ID("") // TODO
}

func (c *Console) Type() Type {
	return console
}

func (c *Console) Activate() error {
	// TODO
	return nil
}

func (c *Console) Deactivate() error {
	// TODO
	return nil
}
