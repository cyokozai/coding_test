package main

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// checkout: {時刻} {利用客 ID} {クーポンの情報}
func handleCheckout(fields []string, currentTime time.Time) {
	userID, _ := strconv.Atoi(fields[2])
	// numCoupons, _ := strconv.Atoi(fields[3])
	// couponIDs := []int{}
	// for i := 0; i < numCoupons; i++ {
	// 	id, _ := strconv.Atoi(fields[4+i])
	// 	couponIDs = append(couponIDs, id)
	// }

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
	// 基本実装では numCoupons == 0 が保証されているため、チェックをスキップ

	// --- 1. 料金計算ロジック ---
	seatFee := calculateSeatFee(session, currentTime)
	foodFee := calculateFoodFee(session /*, couponIDs*/) // 基本実装
	totalFee := seatFee + foodFee + session.TotalShowerFee

	// --- 3. 正常処理 ---
	fmt.Printf("checkout: %d\n", totalFee)

	// 状態更新
	session.Status = UserCheckedOut
	seat := &seats[session.SeatID]
	seat.Status = SeatAvailable // 基本実装
	seat.UserID = 0
	// (応用実装: seat.Status = SeatCleaning, seat.AvailableAt = ...)
}

// calculateSeatFee は、最小料金制に基づき座席料金を計算する
func calculateSeatFee(session *UserSession, checkoutTime time.Time) int {
	duration := checkoutTime.Sub(session.CheckInTime)
	durationSec := duration.Seconds()
	// durationMin := int(math.Ceil(durationSec / 60.0)) // (参考: 最小料金制には使わない)

	// 座席からタイプIDを取得
	seat := seats[session.SeatID]
	seatType, ok := seatTypes[seat.TypeID]
	if !ok {
		return 0
	}
	basicRate := seatType.BasicRate

	// 1. 定額パックを適用しなかった場合の座席料金 (10分=600秒)
	baseFee := calculateFee(durationSec, basicRate, 10*60)
	minFee := baseFee

	// 2. 適用可能な各定額パックを適用した場合の座席料金
	for _, pack := range seatType.Packs {
		packFee := pack.Price
		packDurationSec := float64(pack.TimeMinutes * 60)

		// 利用時間がパック時間を超過した場合
		if durationSec > packDurationSec {
			// 超過時間 (秒)
			overTimeSec := durationSec - packDurationSec
			// 超過料金
			// ★注意: 仕様書「超過時間分の利用に対して基本料金を適用」
			// この「基本料金」の計算は、利用開始1秒後からカウントする calculateFee ではなく、
			// 純粋な超過時間 (10分ごと) で計算する必要がある。
			//
			// 例: 超過 600秒 (10分) -> 1回
			// 例: 超過 601秒 -> 2回
			//
			// 修正: パック超過の基本料金計算は、
			// 「1秒後」ルールではなく「10分ごと（切り上げ）」ルールである。
			//
			// 0秒超過 -> 0
			// 1-600秒超過 -> 1 * basicRate
			// 601-1200秒超過 -> 2 * basicRate
			//
			// overTimeSecを分に切り上げ、10分単位で切り上げる
			overTimeMinutes := int(math.Ceil(overTimeSec / 60.0))
			// 10分単位の課金回数
			// chargeCount := (overTimeMinutes + 9) / 10 // 整数演算での切り上げ (例: 1分->1, 10分->1, 11分->2)

			// 修正 (2025/11/04): サンプル1の仕様を再確認
			// サンプル1: 86分30秒 (1時間26分30秒)
			// 60分パック: 600 + 200 * 3 = 1200
			//   超過: 26分30秒 -> 27分 (切り上げ)
			//   27分は 10分*3回 (1-10, 11-20, 21-30)
			//   chargeCount = (27 + 9) / 10 = 3
			//
			// サンプル1: 10時間17分30秒 (617.5秒) -> 618分
			// 120分パック: 900 + ...
			//   超過: 618 - 120 = 498分
			//   chargeCount = (498 + 9) / 10 = 50
			//   900 + 200 * 50 = 10900
			if overTimeMinutes > 0 {
				chargeCount := (overTimeMinutes + 9) / 10
				overageFee := chargeCount * basicRate
				packFee += overageFee
			}
		}

		if packFee < minFee {
			minFee = packFee
		}
	}

	return minFee
}

// calculateFoodFee は、フード料金の合計を計算する
func calculateFoodFee(session *UserSession /*, couponIDs []int*/) int {
	totalFoodFee := 0

	// 1. クーポン割引なしの合計金額
	for foodID, count := range session.FoodOrders {
		price, ok := foodPrices[foodID]
		if ok {
			totalFoodFee += price * count
		}
	}

	// 2. クーポン割引
	// ...

	return totalFoodFee
}
