package memory

import (
	"bytes"
	"io"
)

type Module struct {
	data io.Reader
	size int
}

func (m *Module) Raw() io.Reader {
	return m.data
}

func (m *Module) Size() int {
	return m.size
}

func ModuleFromBytes(data []byte) *Module {
	return &Module{
		bytes.NewReader(data),
		len(data),
	}
}
