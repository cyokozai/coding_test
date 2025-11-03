package main

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// shower-start: {時刻} {座席 ID}
func handleShowerStart(fields []string, currentTime time.Time) {
	seatID, _ := strconv.Atoi(fields[2])

	// 1. エラーチェック (seat not used)
	_, session := findSessionBySeatID(seatID)
	if session == nil {
		fmt.Println("shower-start: seat not used")
		return
	}

	// 2. エラーチェック (already started)
	if session.CurrentShowerID != 0 {
		fmt.Println("shower-start: already started")
		return
	}

	// 3. 空きシャワーを探す (ID最小)
	foundShowerID := 0
	for i := 1; i < len(showers); i++ {
		// updateAllStates 実行済み
		if showers[i].Status == ShowerAvailable {
			foundShowerID = i
			break
		}
	}

	// 4. エラーチェック (fully occupied)
	if foundShowerID == 0 {
		fmt.Println("shower-start: fully occupied")
		return
	}

	// 5. 状態更新
	showers[foundShowerID].Status = ShowerOccupied
	showers[foundShowerID].UserID = session.ID
	session.CurrentShowerID = foundShowerID
	session.ShowerStartTime = currentTime

	fmt.Printf("shower-start: %d\n", foundShowerID)
}

// shower-end: {時刻} {座席 ID}
func handleShowerEnd(fields []string, currentTime time.Time) {
	seatID, _ := strconv.Atoi(fields[2])

	// 1. エラーチェック (seat not used)
	_, session := findSessionBySeatID(seatID)
	if session == nil {
		fmt.Println("shower-end: seat not used")
		return
	}

	// 2. エラーチェック (not started)
	if session.CurrentShowerID == 0 {
		fmt.Println("shower-end: not started")
		return
	}

	// 3. 正常処理
	showerID := session.CurrentShowerID
	duration := currentTime.Sub(session.ShowerStartTime)
	durationSec := duration.Seconds()

	// 4. 料金計算
	// 15分 = 900秒
	fee := calculateFee(durationSec, showerCharge, 15*60)
	session.TotalShowerFee += fee

	// 5. 利用時間計算 (分単位・切り上げ)
	minutes := int(math.Ceil(durationSec / 60.0))

	// 6. 状態更新
	session.CurrentShowerID = 0
	showers[showerID].Status = ShowerAvailable // 基本実装
	showers[showerID].UserID = 0
	// (応用実装: showers[showerID].Status = ShowerCleaning, ...)

	fmt.Printf("shower-end: %d\n", minutes)
}
