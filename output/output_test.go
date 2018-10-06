package output

import "github.com/pkg/errors"

type alwaysSuccessfulWriter struct{}

func (s *alwaysSuccessfulWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (s *alwaysSuccessfulWriter) Close() error {
	return nil
}

type alwaysFailedWriter struct{}

func (f *alwaysFailedWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("failed to write bytes to")
}

func (f *alwaysFailedWriter) Close() error {
	return nil
}
