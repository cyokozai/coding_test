package main

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// get-duration: {時刻} {座席 ID}
func handleGetDuration(fields []string, currentTime time.Time) {
	seatID, _ := strconv.Atoi(fields[2]) // 1. seats[seatID] を確認
	_, session := findSessionBySeatID(seatID) // 2. StatusがOccupiedでない場合 -> seat not used
	if session == nil {
		fmt.Println("get-duration: seat not used") // findSessionBySeatID は seat not used (Status != Occupied) の場合 session = nil を返す

		return
	}

	// 3. StatusがOccupiedの場合:
	duration := currentTime.Sub(session.CheckInTime)
	minutes := int(math.Ceil(duration.Seconds() / 60.0))
	
	fmt.Printf("get-duration: %d\n", minutes)
}