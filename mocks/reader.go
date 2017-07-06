package mocks

import "fmt"

// A Mock io.reader to trigger errors on reading
type Reader struct {
}

func (f Reader) Read(bytes []byte) (int, error) {
	return 0, fmt.Errorf("Reader failed")
}
