package model

type User struct {
	Id              int64  `json:"id"`
	PublicKey       string `json:"public_key"`
	PresharedKey    string `json:"preshared_key"`
	VirtualEndpoint string `json:"virtual_endpoint"`
}
