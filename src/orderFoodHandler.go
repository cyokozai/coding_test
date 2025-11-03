package main

import (
	"fmt"
	"strconv"
	"time"
)

// order-food: {時刻} {座席 ID} {フード ID}
func handleOrderFood(fields []string, currentTime time.Time) {
	seatID, _ := strconv.Atoi(fields[2])
	foodID, _ := strconv.Atoi(fields[3])

	// 1. seats[seatID] を確認
	// 2. StatusがOccupiedでない場合 -> seat not used
	_, session := findSessionBySeatID(seatID)
	if session == nil {
		fmt.Println("order-food: seat not used")
		return
	}

	// 3. StatusがOccupiedの場合:
	//    - session.FoodOrders[foodID]++
	session.FoodOrders[foodID]++

	//    - fmt.Println("order-food: ok")
	fmt.Println("order-food: ok")
}