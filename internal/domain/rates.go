package domain

import "time"

type Rates struct {
	Timestamp 	time.Duration
	AskPrice	string
	BidPrice	string
}