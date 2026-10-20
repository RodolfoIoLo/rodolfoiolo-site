package storage

import "context"

type ObjectStore interface {
	Put(context.Context, string, string, []byte) error
	Delete(context.Context, string) error
	PublicURL(string) string
}
