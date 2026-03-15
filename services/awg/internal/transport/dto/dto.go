package dto

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

func CreatePeerResponse(publicKey, presharedKey string) Response {
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
