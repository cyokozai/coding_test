package main

import (
	"time"
)

// get-vacant-seats: {時刻}
func handleGetVacantSeats(fields []string, currentTime time.Time) {
	// TODO: ロジック実装 (応用実装)
	// 1. 座席タイプごとに空席をカウント (m種類)
	// 2. (応用) `status == Cleaning` かつ `currentTime >= availableAt` も空席とみなす
	// 3. 結果をスライスに格納
	// 4. 指定フォーマットで出力
}
