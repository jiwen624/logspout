package output

import (
	"io/ioutil"

	"github.com/jiwen624/logspout/log"
)

type Discard struct{}

func (d Discard) Write(p []byte) (n int, err error) {
	return ioutil.Discard.Write(p)
}

func (d Discard) ID() ID {
	return id("discard")
}

func (d Discard) Type() Type {
	return discard
}

func (d Discard) String() string {
	return "discard"
}

func (d Discard) Activate() error {
	log.Info("Activating output discard.")
	return nil
}

func (d Discard) Deactivate() error {
	log.Info("Deactivating output discard.")
	return nil
}
