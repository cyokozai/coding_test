package main

import "time"

type UserStatus int

const (
	UserInUse UserStatus = iota
	UserCheckedOut
)

type UserSession struct {
	ID         int
	SeatID     int
	CheckInTime time.Time
	Status     UserStatus

	// 料金関連
	FoodOrders     map[int]int // フードID -> 注文回数
	TotalShowerFee int         // シャワー料金の累計

	// シャワー利用状況
	CurrentShowerID int       // 利用中のシャワーID (0なら不使用)
	ShowerStartTime time.Time // 現在のシャワー開始時刻
}

// システム全体で以下を持つ:
// nextUserID int = 1
// users map[int]*UserSession // キー: 利用客ID
// システム全体で `map[int]int` (キー: フードID, 値: 価格) として持つ。
// foodPrices[1] ～ foodPrices[f]