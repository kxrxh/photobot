package dto

type OwnershipTransferRequest struct {
	FromUserID int32 `json:"from_user_id"`
	ToUserID   int32 `json:"to_user_id"`
}

type OwnershipTransferResponse struct {
	Message string `json:"message"`
}
