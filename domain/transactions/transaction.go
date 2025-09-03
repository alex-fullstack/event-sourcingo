package transactions

import "github.com/google/uuid"

type Transaction struct {
	ID          uuid.UUID
	SequenceID  int64
	AggregateID uuid.UUID
}

func NewTransaction(id, aggregateID uuid.UUID, sequenceID int64) *Transaction {
	return &Transaction{
		ID:          id,
		SequenceID:  sequenceID,
		AggregateID: aggregateID,
	}
}
