package dto

import "errors"

var ErrUserExist error = errors.New("user already exists")
var ErrNotFound error = errors.New("not found")
