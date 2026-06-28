package objectstore

import (
	"context"
	"net/url"
	"time"
)

func (c *Client) PresignedGetObject(
	ctx context.Context,
	objectKey string,
	expiry time.Duration,
	reqParams url.Values,
) (*url.URL, error) {
	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return c.client.PresignedGetObject(ctx, c.bucket, objectKey, expiry, reqParams)
}
