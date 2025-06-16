package dto

type TransactionHandle struct {
	Id          string `json:"id"`
	SequenceId  string `json:"sequence_id"`
	AggregateId string `json:"aggregate_id"`
}
