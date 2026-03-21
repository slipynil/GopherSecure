package handlers

import (
	"awg-service/internal/repository/model"
	"awg-service/internal/transport/dto"

	awgctrlgo "github.com/slipynil/awgctrl-go"
)

// It is an awg interface for interacting with the AWG service.
type awg interface {
	AddPeer(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error)
	DeletePeer(peerPublicKeyStr string) error
}

type repository interface {
	AddUser(id int64, peer *awgctrlgo.Peer) error
	DeleteUser(publicKey string) error
	DeleteUserEx(publicKey string) (*dto.DeleteResult, error)
	RestoreUser(user *model.User) error
	GetFile(id string) (string, error)
	GetUser(id int64) (*model.User, error)
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
