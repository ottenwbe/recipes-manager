package account

import "github.com/google/uuid"

type Account struct {
	Name string
	ID   AccID
}

type AccID uuid.UUID
