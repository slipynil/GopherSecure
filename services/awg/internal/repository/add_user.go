package repository

import (
	"awg-service/internal/repository/model"
	"encoding/json"
	"os"

	awgctrlgo "github.com/slipynil/awgctrl-go"
)

func (r *Repository) AddUser(id int64, peer *awgctrlgo.Peer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var users []model.User
	content, err := os.ReadFile(r.GlobalFilePath)
	if err == nil && len(content) > 0 {
		if err := json.Unmarshal(content, &users); err != nil {
			return err
		}
	}
	users = append(users, model.User{
		Id:              id,
		PublicKey:       peer.PublicKey,
		PresharedKey:    peer.PresharedKey,
		VirtualEndpoint: peer.VirtualSocket,
	})
	file, err := os.OpenFile(r.GlobalFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(users)
}
