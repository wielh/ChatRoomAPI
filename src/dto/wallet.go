package dto

type GetStateRequest struct {
	UserID uint64
}

type GetStateResponse struct {
	Money uint32
}

type ChargeRequest struct {
	UserID uint64
	Money  uint32
}

type ChargeResponse struct {
	OK         bool
	MinAccount int
	MaxAccount int
}
