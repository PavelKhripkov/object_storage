package bucket

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/bucket"
)

type bucketStorage interface {
	Get(ctx context.Context, id string) (bucket.Bucket, error)
	Create(ctx context.Context, bucket bucket.Bucket) error
	Delete(ctx context.Context, bucket bucket.Bucket) error
}
