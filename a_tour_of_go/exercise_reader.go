package main

import "golang.org/x/tour/reader"

type MyReader struct {
	pos int
}

// TODO: Add a Read([]byte) (int, error) method to MyReader.
func (r MyReader) Read(data []byte) (num int, e error) {
	data[r.pos] = 'A'
	r.pos += 1
	return 1, nil
}

func main() {
	reader.Validate(MyReader{})
}
