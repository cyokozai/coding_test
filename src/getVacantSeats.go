package main

import (
	"fmt"
	"strings"
	"time"
)

// get-vacant-seats: {時刻}
func handleGetVacantSeats(fields []string, currentTime time.Time) {
	// 1. 座席タイプごとに空席をカウント (m種類)
	//    キー: SeatTypeID, 値: 空席数
	vacanciesPerType := make(map[int]int)

	// 全座席をスキャン
	for i := 1; i < len(seats); i++ {
		seat := &seats[i]
		isAvailable := false

		// 2. [応用実装] の「空席」定義
		// (updateAllStates は既に実行済みだが、
		//  currentTime ちょうどに利用可能になるケースを拾うため、
		//  ここで再度チェックするのが最も安全)
		if seat.Status == SeatAvailable {
			isAvailable = true
		} else if seat.Status == SeatCleaning && !currentTime.Before(seat.AvailableAt) {
			// 清掃中で、現在時刻が利用可能時刻を過ぎている (以上である)
			isAvailable = true
		}

		if isAvailable {
			vacanciesPerType[seat.TypeID]++
		}
	}

	// 3. 結果を指定フォーマットで出力
	//    仕様: 空席が存在する座席タイプについて、座席タイプ ID の昇順に出力
	
	outputLines := []string{}
	v := 0 // 空席が存在する座席タイプの種類数

	// main.go で seatTypes (map) は 1 から m まで初期化されている
	// 1 から m (seatTypesの数) までループすれば、自動的にID昇順になる
	m := len(seatTypes) 
	for seatTypeID := 1; seatTypeID <= m; seatTypeID++ {
		
		count := vacanciesPerType[seatTypeID] // マップになければ 0 が返る

		// 仕様: 空席が存在する座席タイプのみ出力
		if count > 0 {
			v++
			outputLines = append(outputLines, fmt.Sprintf("%d %d", seatTypeID, count))
		}
	}

	// 4. 最終出力
	fmt.Printf("get-vacant-seats: %d\n", v)
	if v > 0 {
		fmt.Println(strings.Join(outputLines, "\n"))
	}
}
