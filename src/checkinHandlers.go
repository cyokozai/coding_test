package main

import (
	"fmt"
	"strconv"
	"time"
)


// checkin: {時刻} {座席タイプ}
func handleCheckIn(fields []string, currentTime time.Time) {
	seatTypeID, _ := strconv.Atoi(fields[2])

	// 1. 指定seatTypeIDの空席を探す (座席ID最小)
	foundSeatID := 0
	for i := 1; i < len(seats); i++ {
		// updateAllStates 実行済みなので、単純にAvailableかチェックするだけで良い
		if seats[i].TypeID == seatTypeID && seats[i].Status == SeatAvailable {
			foundSeatID = seats[i].ID
			break // 座席IDが最小のもの
		}
	}

	// 3. 見つからない場合
	if foundSeatID == 0 {
		fmt.Println("checkin: fully occupied")
		return
	}

	// 4. 見つかった場合:
	//    - 新しいUserSessionを作成
	userID := nextUserID
	nextUserID++

	session := &UserSession{
		ID:              userID,
		SeatID:          foundSeatID,
		CheckInTime:     currentTime,
		Status:          UserInUse,
		FoodOrders:      make(map[int]int),
		TotalShowerFee:  0,
		CurrentShowerID: 0,
	}

	//    - usersマップに追加 (nextUserID)
	users[userID] = session

	//    - Seatの状態を更新 (Status, UserID)
	seats[foundSeatID].Status = SeatOccupied
	seats[foundSeatID].UserID = userID

	//    - 出力
	fmt.Printf("checkin: userid = %d, seatid = %d\n", userID, foundSeatID)
}
