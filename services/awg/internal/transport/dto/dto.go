package dto

import "awg-service/internal/repository/model"

type Response struct {
	Data  any
	Error string
}

type Request struct {
	ID              int64  `json:"id"`
	VirtualEndpoint string `json:"virtual_endpoint"`
	DNS             string `json:"dns,omitempty"`
}

type DelRequest struct {
	PublicKey string `json:"public_key"`
}

// DeleteResult contains information about the deletion operation result
type DeleteResult struct {
	Found   bool         // was the user found
	Deleted bool         // was the user successfully deleted
	User    *model.User  // deleted user (for potential rollback)
}

func CreatePeerResponse(publicKey string) Response {
	return Response{
		Data: struct {
			PublicKey string `json:"public_key"`
		}{
			PublicKey: publicKey,
		},
	}
}

func CreateKeysResponse(publicKey string, presharedKey string) Response {
	return Response{
		Data: struct {
			PublicKey    string `json:"public_key"`
			PresharedKey string `json:"preshared_key"`
		}{
			PublicKey:    publicKey,
			PresharedKey: presharedKey,
		},
	}
}
