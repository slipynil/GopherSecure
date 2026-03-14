package handlers

import (
	awgctrlgo "github.com/slipynil/awgctrl-go"
)

// It is an awg interface for interacting with the AWG service.
type awg interface {
	AddPeer(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error)
	DeletePeer(peerPublicKeyStr string) error
}

type repository interface {
	GetFile(id string) (string, error)
}

// It is a handlers struct that contains the AWG service and handles HTTP requests.
type handlers struct {
	awg        awg
	repository repository
}

func New(awg awg, r repository) *handlers {
	return &handlers{
		awg:        awg,
		repository: r,
	}
}
