// Package media stores uploaded files (product/section images) in an S3-compatible object store
// (MinIO in dev/self-host, any S3 in prod). It is a thin wrapper over minio-go: the one new dependency
// the storefront builder genuinely needs — a real merchant cannot sell anything without uploading their
// own photos, and signing S3 PUTs by hand is exactly the kind of fiddly code a small, audited client
// avoids. Objects are written under a per-tenant prefix so one store can never overwrite another's media.
package media

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client uploads objects to a single bucket and hands back a public URL the storefront can <img src>.
type Client struct {
	mc         *minio.Client
	bucket     string
	publicBase string // e.g. http://localhost:9000 — the bucket is public-read, so URLs need no signing.
}

// Config wires the object store. Endpoint is host:port (no scheme); UseSSL toggles https on the wire.
type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicBase string
	UseSSL     bool
}

// New connects to the object store and ensures the bucket exists. It returns (nil, nil) when no endpoint
// is configured, so a deployment without MinIO simply has uploads disabled rather than failing to boot.
func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, nil
	}
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("media: connect: %w", err)
	}
	exists, err := mc.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("media: bucket check: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("media: make bucket: %w", err)
		}
	}
	return &Client{
		mc:         mc,
		bucket:     cfg.Bucket,
		publicBase: strings.TrimRight(cfg.PublicBase, "/"),
	}, nil
}

// Upload writes r (size bytes) under the given tenant's prefix with a random name + ext, and returns the
// public URL. The random name avoids collisions and stops a caller from guessing/overwriting another
// object; the tenant prefix keeps stores isolated even within the shared bucket.
func (c *Client) Upload(ctx context.Context, tenantID, ext, contentType string, r io.Reader, size int64) (string, error) {
	name, err := randomHex()
	if err != nil {
		return "", err
	}
	key := sanitizePrefix(tenantID) + "/" + name + ext
	_, err = c.mc.PutObject(ctx, c.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("media: put object: %w", err)
	}
	return c.publicBase + "/" + c.bucket + "/" + key, nil
}

// sanitizePrefix keeps only safe characters so a tampered tenant id can't escape its prefix (path
// traversal) or break the key. Tenant ids are ObjectID hex in practice, so this is belt-and-suspenders.
func sanitizePrefix(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "shared"
	}
	return b.String()
}

func randomHex() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
