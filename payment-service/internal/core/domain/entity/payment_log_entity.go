package entity

import "github.com/google/uuid"

type PaymentLogEntity struct {
	ID        uuid.UUID
	PaymentID uuid.UUID
	Status    string
}
