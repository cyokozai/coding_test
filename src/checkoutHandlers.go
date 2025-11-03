package main

import (
	"fmt"
	"math"
	"strconv"
	"time"
)


// calculateSeatFee は座席利用料金を計算する（最小料金制）
func calculateSeatFee(session *UserSession, currentTime time.Time) int {
	// 利用時間（秒）を計算
	durationSec := currentTime.Sub(session.CheckInTime).Seconds()
	totalMinutes := int(math.Ceil(durationSec / 60.0))

	// 座席タイプを取得
	seatTypeID := seats[session.SeatID].TypeID
	seatType := seatTypes[seatTypeID]

	// 1. 基本料金を計算（パックなし）
	// 仕様: 1秒後に初回加算、以降10分(600秒)ごと
	basicFee := calculateFee(durationSec, seatType.BasicRate, 600)
	
	// 最小料金の初期値は基本料金
	minFee := basicFee

	// 2. 各パック適用時の料金を計算し、最小値を探す
	for _, pack := range seatType.Packs {
		var packFee int
		
		if totalMinutes <= pack.TimeMinutes {
			// パック時間内に収まる場合はパック料金のみ
			packFee = pack.Price
		} else {
			// パック時間を超過した場合
			overTimeMinutes := totalMinutes - pack.TimeMinutes
			overTimeSeconds := float64(overTimeMinutes * 60)
			
			// 超過分を基本料金で計算
			overageFee := calculateFee(overTimeSeconds, seatType.BasicRate, 600)
			packFee = pack.Price + overageFee
		}
		
		// 最小値を更新
		if packFee < minFee {
			minFee = packFee
		}
	}

	return minFee
}

// calculateFoodFee はフード料金を計算する
func calculateFoodFee(session *UserSession) int {
	totalFoodFee := 0
	for foodID, count := range session.FoodOrders {
		price := foodPrices[foodID]
		totalFoodFee += price * count
	}
	return totalFoodFee
}

// checkout: {時刻} {利用客 ID} {クーポンの情報}
func handleCheckout(fields []string, currentTime time.Time) {
	userID, _ := strconv.Atoi(fields[2])
	// ... (エラーチェック部分は変更なし) ...

	// --- 2. エラーチェック ---
	// 2a. user存在チェック -> invalid user
	session, exists := users[userID]
	if !exists {
		fmt.Println("checkout: invalid user")
		return
	}

	// 2b. 精算済みチェック -> already done
	if session.Status == UserCheckedOut {
		fmt.Println("checkout: already done")
		return
	}

	// 2c. シャワー利用中チェック -> shower is still in use
	if session.CurrentShowerID != 0 {
		fmt.Println("checkout: shower is still in use")
		return
	}

	// 2d. (応用) クーポンチェック
	// (この後のステップで実装します)

	// --- 1. 料金計算ロジック ---
	seatFee := calculateSeatFee(session, currentTime)
	foodFee := calculateFoodFee(session /*, couponIDs*/)
	totalFee := seatFee + foodFee + session.TotalShowerFee

	// --- 3. 正常処理 ---
	fmt.Printf("checkout: %d\n", totalFee)

	// 状態更新
	session.Status = UserCheckedOut
	seat := &seats[session.SeatID]
	seat.UserID = 0 // UserIDは先にクリア

	// ★ 修正点: [応用実装] 清掃ロジック
	// cleaningDurationSeat (分) は main.go で time.Duration として定義済み
	if cleaningDurationSeat > 0 {
		seat.Status = SeatCleaning
		// 現在時刻 + 清掃時間(分) = 利用可能になる時刻
		seat.AvailableAt = currentTime.Add(cleaningDurationSeat)
	} else {
		// 基本実装 (清掃時間 0) の場合
		seat.Status = SeatAvailable
	}
}
