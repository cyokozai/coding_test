package main

import "time"

type SeatStatus int

const (
	SeatAvailable SeatStatus = iota
	SeatOccupied
	SeatCleaning
)

type Seat struct {
	ID          int
	TypeID      int
	Status      SeatStatus
	UserID      int
	AvailableAt time.Time
}

// システム全体で `[]Seat` (サイズ n+1) として持つ。
// seats[1] ～ seats[n] を使用する。
