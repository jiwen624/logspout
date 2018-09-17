package output

type Kafka struct {
	// TODO
}

func (k Kafka) Write(p []byte) (n int, err error) {
	// TODO
	return 0, nil
}

func (k Kafka) ID() ID {
	return ID("") // TODO
}

func (k Kafka) Type() Type {
	return kafka
}
