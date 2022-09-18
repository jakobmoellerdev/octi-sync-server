package service

import (
	"bytes"
	"io"
)

type RedisModule struct {
	data io.Reader
	size int
}

func (r RedisModule) Raw() io.Reader {
	return r.data
}

func (r RedisModule) Size() int {
	return r.size
}

func RedisModuleFromBytes(data []byte) *RedisModule {
	return &RedisModule{
		bytes.NewReader(data),
		int(len(data)),
	}
}

func RedisModuleFromReader(data io.Reader, size int) *RedisModule {
	return &RedisModule{
		data,
		size,
	}
}
