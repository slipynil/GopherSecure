package dto

import (
	"encoding/base64"
	"encoding/json"
)

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
		if psk, ok := dataMap["preshared_key"].(string); ok {
			return psk
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

func DecodeCallbackData(raw string) (*CallbackData, error) {
	bs, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var data CallbackData
	if err := json.Unmarshal(bs, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// base64 кодирование для безопасной передачи действий в Inline кнопках
func EncodeCallbackData(action string) string {
	data := CallbackData{Action: action}
	bs, _ := json.Marshal(data)
	// можно добавить base64 encoding, если бояться спецсимволов
	return base64.RawURLEncoding.EncodeToString(bs)
}
