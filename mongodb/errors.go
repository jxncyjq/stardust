package mongodb

import "errors"

var (
	ErrNotFound      = errors.New("mongodb: document not found")
	ErrInsertFailed  = errors.New("mongodb: insert failed")
	ErrUpdateFailed  = errors.New("mongodb: update failed")
	ErrDeleteFailed  = errors.New("mongodb: delete failed")
	ErrNilClient     = errors.New("mongodb: client is nil")
	ErrInvalidConfig = errors.New("mongodb: invalid config")
)
