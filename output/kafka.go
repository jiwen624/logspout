package output

type Kafka struct {
	// TODO
}

func (k *Kafka) Write(p []byte) (n int, err error) {
	// TODO
	return 0, nil
}

func (k *Kafka) ID() ID {
	return ID("") // TODO
}

func (k *Kafka) String() string {
	return "Kafka"
}

func (k *Kafka) Type() Type {
	return kafka
}

func (k *Kafka) Activate() error {
	// TODO
	return nil
}

func (k *Kafka) Deactivate() error {
	// TODO
	return nil
}
