package redis

import (
	"bytes"
	"io"
)

type Module struct {
	data io.Reader
	size int
}

func (r Module) Raw() io.Reader {
	return r.data
}

func (r Module) Size() int {
	return r.size
}

func ModuleFromBytes(data []byte) *Module {
	return &Module{
		bytes.NewReader(data),
		len(data),
	}
}

func ModuleFromReader(data io.Reader, size int) *Module {
	return &Module{
		data,
		size,
	}
}
