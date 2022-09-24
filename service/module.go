package service

import (
	"io"
)

//go:generate mockgen -source module.go -package mock -destination mock/module.go Module
type Module interface {
	Raw() io.Reader
	Size() int
}
