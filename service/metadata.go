package service

import (
	"context"
	"errors"
	"time"
)

type MetadataID string

type ModifiedAt time.Time

//go:generate mockgen -source metadata.go -package mock -destination mock/metadata.go MetadataProvider
type MetadataProvider interface {
	Get(ctx context.Context, id MetadataID) (Metadata, error)
	Set(ctx context.Context, meta Metadata) error
}

type Metadata interface {
	GetID() MetadataID
	GetModifiedAt() ModifiedAt
}

var ErrNoMetadata = errors.New("no metadata found")

type BaseMetadata struct {
	ID         MetadataID `yaml:"id" json:"id"`
	ModifiedAt time.Time  `yaml:"modifiedAt" json:"modifiedAt"`
}

func (r *BaseMetadata) GetID() MetadataID {
	return r.ID
}

func (r *BaseMetadata) GetModifiedAt() ModifiedAt {
	return ModifiedAt(r.ModifiedAt.UTC())
}

func NewBaseMetadata(id string, modifiedAt time.Time) *BaseMetadata {
	return &BaseMetadata{MetadataID(id), modifiedAt}
}
