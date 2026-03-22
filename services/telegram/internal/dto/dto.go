package dto

type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func (resp *Response) GetKey() string {
	if dataMap, ok := resp.Data.(map[string]any); ok {
		if pubKey, ok := dataMap["public_key"].(string); ok {
			return pubKey
		}
	}
	return ""
}

func (resp *Response) GetPresharedKey() string {
	if dataMap, ok := resp.Data.(map[string]any); ok {
		if presharedKey, ok := dataMap["preshared_key"].(string); ok {
			return presharedKey
		}
	}
	return ""
}

type AddPeerRequest struct {
	ID              int64  `json:"id"`
	VirtualEndpoint string `json:"virtual_endpoint"`
	DNS             string `json:"dns,omitempty"`
}

type DelPeerRequest struct {
	PublicKey string `json:"public_key"`
}

type CallbackData struct {
	Action string `json:"action"`
}

type PaymentHandler struct {
	InvoicePayload string
	TotalAmount    int
	Currency       string
}
