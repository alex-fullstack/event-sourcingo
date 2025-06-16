package transactions

import "github.com/google/uuid"

type Transaction struct {
	Id          uuid.UUID
	SequenceId  int64
	AggregateId uuid.UUID
}

func NewTransaction(id, aggregateId uuid.UUID, sequenceId int64) *Transaction {
	return &Transaction{
		Id:          id,
		SequenceId:  sequenceId,
		AggregateId: aggregateId,
	}
}
