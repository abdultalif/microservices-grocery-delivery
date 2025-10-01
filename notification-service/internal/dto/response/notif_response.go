package response

import "github.com/google/uuid"

type ListResponse struct {
	ID      uuid.UUID `json:"id"`
	Subject string    `json:"subject"`
	Status  string    `json:"status"`
	SentAt  string    `json:"sent_at"`
}
