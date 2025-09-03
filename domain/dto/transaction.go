package dto

type TransactionHandle struct {
	ID          string `json:"id"`
	SequenceID  string `json:"sequence_id"`
	AggregateID string `json:"aggregate_id"`
}
