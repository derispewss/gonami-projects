package domain

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrNoAmount       = errors.New("tidak ada nominal yang bisa dideteksi")
	ErrAlreadyHandled = errors.New("draft sudah diproses sebelumnya")
)
