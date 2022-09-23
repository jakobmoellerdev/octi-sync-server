package service

import (
	"io"
)

//nolint:lll
//go:generate mockgen -package mock -destination mock/module.go github.com/jakob-moeller-cloud/octi-sync-server/service Module
type Module interface {
	Raw() io.Reader
	Size() int
}
