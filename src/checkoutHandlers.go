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
// calculateFoodFee は、フード料金の合計を計算する (★ 応用実装)
func calculateFoodFee(session *UserSession, couponIDs []int) int {
	totalFoodFee := 0

	// 1. (応用) ルール3: 同じフードを対象とするクーポンが複数ある場合、
	//    割引額が最大のクーポン1つだけが適用される。
	//    -> フードIDごとに「適用すべき最大割引額」を計算する
	maxDiscountPerFood := make(map[int]int) // key: foodID, value: max discount
	for _, couponID := range couponIDs {
		coupon, ok := coupons[couponID]
		if !ok {
			continue // (ありえない)
		}

		targetFoodID := coupon.TargetFoodID
		discount := coupon.Discount

		// 現在の最大割引額よりも大きければ更新
		if discount > maxDiscountPerFood[targetFoodID] {
			maxDiscountPerFood[targetFoodID] = discount
		}
	}

	// 2. 割引を適用しながらフード料金を合計
	for foodID, count := range session.FoodOrders {
		price, ok := foodPrices[foodID]
		if !ok {
			continue // (ありえない)
		}

		// このフードIDに対する最大割引額を取得 (クーポンがなければ 0)
		discount := maxDiscountPerFood[foodID]

		// 割引適用後の単価
		// (仕様書より discount <= price は保証されている)
		unitPriceAfterDiscount := price - discount

		// 注文回数分を加算 (ルール2)
		totalFoodFee += unitPriceAfterDiscount * count
	}

	return totalFoodFee
}

// checkout: {時刻} {利用客 ID} {クーポンの情報}
func handleCheckout(fields []string, currentTime time.Time) {
	userID, _ := strconv.Atoi(fields[2])
	numCoupons, _ := strconv.Atoi(fields[3])
	couponIDs := make([]int, numCoupons)
	for i := 0; i < numCoupons; i++ {
		id, _ := strconv.Atoi(fields[4+i])
		couponIDs[i] = id
	}

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

	// 2d. ★ 修正点: [応用実装] クーポンチェック
	// 注文していないフードのクーポンが含まれているかチェック
	for _, couponID := range couponIDs {
		coupon, ok := coupons[couponID]
		if !ok {
			// (テストケースでは発生しないはずだが、念のため)
			fmt.Println("checkout: invalid coupon")
			return
		}
		// 注文回数を取得 (注文してなければ 0)
		orderCount := session.FoodOrders[coupon.TargetFoodID]
		if orderCount == 0 {
			// ルール4: 注文していないフードのクーポンは使えない
			fmt.Println("checkout: invalid coupon")
			return
		}
	}

	// --- 1. 料金計算ロジック ---
	seatFee := calculateSeatFee(session, currentTime)
	// ★ 修正点: couponIDs を渡す
	foodFee := calculateFoodFee(session, couponIDs)
	totalFee := seatFee + foodFee + session.TotalShowerFee

	// --- 3. 正常処理 ---
	fmt.Printf("checkout: %d\n", totalFee)

	// ... (状態更新部分は変更なし) ...
	// (State update part remains unchanged)
	session.Status = UserCheckedOut
	seat := &seats[session.SeatID]
	seat.UserID = 0

	if cleaningDurationSeat > 0 {
		seat.Status = SeatCleaning
		seat.AvailableAt = currentTime.Add(cleaningDurationSeat)
	} else {
		seat.Status = SeatAvailable
	}
}
