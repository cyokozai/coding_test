package main

import "time"

type ShowerStatus int

const (
    ShowerAvailable ShowerStatus = iota
    ShowerOccupied
    ShowerCleaning // 応用実装用
)

type Shower struct {
    ID        int
    Status    ShowerStatus
    UserID    int       // このシャワーを使っている利用客ID (0なら空き)
    AvailableAt time.Time // 清掃が完了し利用可能になる時刻 (応用実装用)
}

// システム全体で以下を持つ:
// showers []Shower // サイズ s+1
// showerCharge int // シャワーの基本料金 (15分ごと)