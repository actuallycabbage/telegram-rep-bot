package db

import "errors"

var (
	ErrAccNotFound         = errors.New("Account not found")
	ErrNotImplemented      = errors.New("Not Implemented")
	ErrUnsupportedDatabase = errors.New("Unsupported database type")
	ErrNoRowsAffected      = errors.New("No rows were affected")
)
