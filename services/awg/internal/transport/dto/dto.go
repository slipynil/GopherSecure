package dto

type Response struct {
	Data  any
	Error string
}

type Request struct {
	DNS             string `json:"dns,omitempty"`
	VirtualEndpoint string `json:"virtual_endpoint"`
	ID              int64  `json:"id"`
}

type DelRequest struct {
	PublicKey string `json:"public_key"`
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
