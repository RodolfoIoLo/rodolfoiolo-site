package requestctx

import "context"

type Key string

const (
	identityKey Key = "identity"
	metadataKey Key = "metadata"
	refreshKey  Key = "refresh"
)

type Identity struct {
	UserID int64
	Role   string
}

type Metadata struct {
	UserAgent string
	IPHash    string
	VisitorID string
}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityKey, identity)
}

func IdentityFrom(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityKey).(Identity)
	return identity, ok
}

func WithMetadata(ctx context.Context, metadata Metadata) context.Context {
	return context.WithValue(ctx, metadataKey, metadata)
}

func MetadataFrom(ctx context.Context) Metadata {
	metadata, _ := ctx.Value(metadataKey).(Metadata)
	return metadata
}

func WithRefreshToken(ctx context.Context, raw string) context.Context {
	return context.WithValue(ctx, refreshKey, raw)
}

func RefreshTokenFrom(ctx context.Context) string {
	raw, _ := ctx.Value(refreshKey).(string)
	return raw
}
