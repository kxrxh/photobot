package ownership

import "context"

// NoopTransfer satisfies auth.OwnershipTransferClient when no downstream merge URLs are configured.
type NoopTransfer struct{}

func (NoopTransfer) TransferOwnership(context.Context, int32, int32) error { return nil }
