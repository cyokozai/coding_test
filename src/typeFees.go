package main

type Pack struct {
    TimeMinutes int // パック時間 (分)
    Price       int // パック料金
}

type SeatType struct {
    ID        int
    BasicRate int    // 基本料金 (10分ごと)
    Packs     []Pack // 定額パックのリスト
}

// システム全体で `map[int]SeatType` (キー: 座席タイプID) として持つ。
// seatTypes[1] ～ seatTypes[m]