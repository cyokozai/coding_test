package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)


var (
	// getStdin() 入力を保持
	lines            []string // 標準入力の全行
	currentLineIndex int      = 0      // 現在読み込んでいる行番号

	// 座席・清掃
	seats                []Seat // 1-indexed (size n+1)
	cleaningDurationSeat time.Duration

	// 座席タイプ・料金
	seatTypes map[int]SeatType // 1-indexed (key: SeatTypeID)

	// フード・クーポン
	foodPrices map[int]int  // 1-indexed (key: FoodID)
	coupons    map[int]Coupon // 1-indexed (key: CouponID)

	// シャワー
	showers                []Shower // 1-indexed (size s+1)
	showerCharge           int
	cleaningDurationShower time.Duration

	// 利用客管理
	nextUserID int = 1                  // 次に発行する利用客ID
	users      map[int]*UserSession // key: UserID
)

// ---------- 時刻・入力パーサー ----------
const timeLayout = "2006/01/02-15:04:05"

// parseTime は文字列をtime.Timeに変換する
func parseTime(s string) time.Time {
	t, _ := time.Parse(timeLayout, s)

	return t
}

// nextLine は標準入力から1行読み込む
func nextLine() string {
	if currentLineIndex >= len(lines) {
		fmt.Fprintf(os.Stderr, "Error: Reached end of input unexpectedly\n")
		
		return ""
	}
	line := lines[currentLineIndex]

	currentLineIndex++

	return line
}

// nextInt は標準入力から1行読み込み、整数に変換する
func nextInt() int {
	n, _ := strconv.Atoi(nextLine())

	return n
}

// nextInts はスペース区切りの1行を整数スライスに変換する
func nextInts(line string) []int {
	parts := strings.Split(line, " ")
	ints := make([]int, len(parts))
	for i, s := range parts {
		ints[i], _ = strconv.Atoi(s)
	}

	return ints
}

// ---------- 初期入力パーサー ----------
func parseInitialInput() {
	// N, M, cleaning_duration_seat
	line1 := nextLine()
	parts1 := nextInts(line1)
	n, m := parts1[0], parts1[1]
	cleaningDurationSeat = time.Duration(parts1[2]) * time.Minute

	// 座席配列の初期化 (1-indexed)
	seats = make([]Seat, n+1)

	// 座席タイプ (type_1 ... type_n)
	line2 := nextLine()
	typeIDs := nextInts(line2)
	for i := 1; i <= n; i++ {
		seats[i] = Seat{
			ID:     i,
			TypeID: typeIDs[i-1], // 入力は0-indexed相当だが座席IDは1-indexed
			Status: SeatAvailable,
		}
	}

	// 座席タイプごとの料金情報 (pack_1 ... pack_m)
	seatTypes = make(map[int]SeatType) // 修正: 初期化
	for i := 1; i <= m; i++ {
		// basic, p
		lineBasic := nextLine()
		partsBasic := nextInts(lineBasic)
		basicRate, p := partsBasic[0], partsBasic[1]

		st := SeatType{
			ID:        i,
			BasicRate: basicRate,
			Packs:     make([]Pack, p),
		}

		for j := 0; j < p; j++ {
			linePack := nextLine()
			partsPack := nextInts(linePack)
			st.Packs[j] = Pack{
				TimeMinutes: partsPack[0],
				Price:       partsPack[1],
			}
		}
		seatTypes[i] = st
	}

	f := nextInt()
	foodPrices = make(map[int]int) // 修正: 初期化
	lineFoodPrices := nextLine()
	foodPriceList := nextInts(lineFoodPrices)
	for i := 1; i <= f; i++ {
		foodPrices[i] = foodPriceList[i-1] // 0-indexed相当 -> 1-indexed
	}

	c := nextInt()
	coupons = make(map[int]Coupon) // 修正: 初期化
	for i := 1; i <= c; i++ {
		lineCoupon := nextLine()
		partsCoupon := nextInts(lineCoupon)
		coupons[i] = Coupon{
			ID:           i,
			TargetFoodID: partsCoupon[0],
			Discount:     partsCoupon[1],
		}
	}

	// シャワー (s, shower_charge, cleaning_duration_shower)
	lineShower := nextLine()
	partsShower := nextInts(lineShower)
	s := partsShower[0]
	showerCharge = partsShower[1]
	cleaningDurationShower = time.Duration(partsShower[2]) * time.Minute

	// シャワー配列の初期化 (1-indexed)
	showers = make([]Shower, s+1)
	for i := 1; i <= s; i++ {
		showers[i] = Shower{
			ID:     i,
			Status: ShowerAvailable,
		}
	}
}

// ---------- 状態更新 ----------
func updateAllStates(currentTime time.Time) {
	// 基本実装では cleaningDuration が 0 なので、Cleaning状態にはならず、
	// この関数は実質何もしないが、応用実装のためにロジックを入れておく。

	// 座席の清掃チェック
	for i := 1; i < len(seats); i++ {
		seat := &seats[i] // ポインタで操作
		// 清掃中で、現在時刻が利用可能時刻を過ぎていれば
		if seat.Status == SeatCleaning && !currentTime.Before(seat.AvailableAt) {
			seat.Status = SeatAvailable
			seat.UserID = 0
		}
	}

	// シャワーの清掃チェック
	for i := 1; i < len(showers); i++ {
		shower := &showers[i] // ポインタで操作
		if shower.Status == ShowerCleaning && !currentTime.Before(shower.AvailableAt) {
			shower.Status = ShowerAvailable
			shower.UserID = 0
		}
	}
}

// findSessionBySeatID 座席IDから利用客セッションを検索する
func findSessionBySeatID(seatID int) (int, *UserSession) {
    if seatID < 1 || seatID >= len(seats) {
        return 0, nil
    }
    
    seat := &seats[seatID]
    if seat.Status != SeatOccupied {
        return 0, nil
    }
    
    userID := seat.UserID
    session, exists := users[userID]
    if !exists {
        return 0, nil
    }
    
    return userID, session
}

// calculateFee は、基本料金（座席・シャワー）の計算ロジック
// 1秒後に初回加算、その後は intervalSec ごとに加算
// durationSec: 利用秒数
// rate: 1回あたりの料金
// intervalSec: 2回目以降の加算間隔（秒）
// 2回目以降の加算
// (durationSec - 1) が intervalSec をいくつ含むか
// 例: duration=1.0, interval=600 -> overSec=0.0, add=0
// 例: duration=600.9, interval=600 -> overSec=599.9, add=0
// 例: duration=601.0, interval=600 -> overSec=600.0, add=1
func calculateFee(durationSec float64, rate int, intervalSec int) int {
	if durationSec < 1.0 {
		return 0
	}

	// 1. 初回加算 (1秒経過時点で発生)
	count := 1
	// 2. 2回目以降の加算
	if intervalSec > 0 {
		overSec := durationSec - 1.0
		additionalCount := int(overSec / float64(intervalSec))
		count += additionalCount
	}

	return count * rate
}

// ---------- メインロジック ----------
func main() {
	// 標準入力スキャナのセットアップ
	// Note: バッファサイズを大きくすると、大量の入力にも対応できる
	// --- ステップ0: 全入力を読み込む ---
	lines = getStdin()

	// グローバル状態変数の初期化
	users = make(map[int]*UserSession)

	// --- ステップ1: 初期入力の解析 ---
	// parseInitialInput は内部で nextLine() を呼び出すため、
	// lines が設定された後に呼び出す
	parseInitialInput()

	// --- ステップ2: クエリの解析と実行 ---
	q := nextInt() // q (クエリ数)
	for i := 0; i < q; i++ {
		line := nextLine()
		if line == "" {
			// 入力が尽きた場合
			break
		}
		fields := strings.Split(line, " ")
		command := fields[0]

		currentTime := parseTime(fields[1])
		updateAllStates(currentTime)

		switch command {
		case "checkin:":
			handleCheckIn(fields, currentTime)
		case "get-duration:":
			handleGetDuration(fields, currentTime)
		case "checkout:":
			handleCheckout(fields, currentTime)
		case "order-food:":
			handleOrderFood(fields, currentTime)
		case "shower-start:":
			handleShowerStart(fields, currentTime)
		case "shower-end:":
			handleShowerEnd(fields, currentTime)
		case "get-vacant-seats:":
			handleGetVacantSeats(fields, currentTime)
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		}
	}
}

// getStdin は標準入力から全行を読み込む
func getStdin() []string {
	stdin, _ := ioutil.ReadAll(os.Stdin)

	return strings.Split(strings.TrimRight(string(stdin), "\n"), "\n")
}
