package repository

import "sync"

type Repository struct {
	StoragePath string
	mu          sync.Mutex
}

func New(path string) *Repository {
	return &Repository{
		StoragePath: path,
	}
}
