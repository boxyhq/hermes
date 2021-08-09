package auth

import "context"

type authMetadataCtxKey string

var ctxKey = authMetadataCtxKey("authMetadataCtxKey")

func withAuthMetadata(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, ctxKey, md)
}

func MetadataFromCtx(ctx context.Context) Metadata {
	if v, ok := ctx.Value(ctxKey).(Metadata); ok {
		return v
	}
	return Metadata{}
}
