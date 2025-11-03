package main

import (
	"fmt"
	"io/ioutil"
	// "math"
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

// ---------- クエリハンドラ (スタブ) ----------
// これからこれらの中身を実装していきます。

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

// get-duration: {時刻} {座席 ID}
func handleGetDuration(fields []string, currentTime time.Time) {
	// seatID, _ := strconv.Atoi(fields[2])

	// TODO: ロジック実装
	// 1. seats[seatID] を確認
	// 2. StatusがOccupiedでない場合 -> seat not used
	// 3. StatusがOccupiedの場合:
	//    - seat.UserIDを使って users マップから UserSession を取得
	//    - duration := currentTime.Sub(session.CheckInTime)
	//    - minutes := int(math.Ceil(duration.Seconds() / 60.0))
	//    - fmt.Printf("get-duration: %d\n", minutes)
}

// checkout: {時刻} {利用客 ID} {クーポンの情報}
func handleCheckout(fields []string, currentTime time.Time) {
	// userID, _ := strconv.Atoi(fields[2])
	// numCoupons, _ := strconv.Atoi(fields[3])
	// couponIDs := []int{}
	// for i := 0; i < numCoupons; i++ {
	// 	id, _ := strconv.Atoi(fields[4+i])
	// 	couponIDs = append(couponIDs, id)
	// }

	// TODO: ロジック実装 (基本実装では numCoupons == 0)
	// 1. 料金計算ロジック（別関数に切り出す）
	//    - calculateSeatFee(session, currentTime)
	//    - calculateFoodFee(session)
	//    - totalFee := seatFee + foodFee + session.TotalShowerFee
	// 2. エラーチェック
	//    - user存在チェック -> invalid user
	//    - 精算済みチェック -> already done
	//    - シャワー利用中チェック -> shower is still in use
	//    - (応用) クーポンチェック -> invalid coupon
	// 3. 正常処理
	//    - fmt.Printf("checkout: %d\n", totalFee)
	//    - session.Status = UserCheckedOut
	//    - seat.Status = SeatAvailable (基本実装)
	//    - seat.UserID = 0
}

// order-food: {時刻} {座席 ID} {フード ID}
func handleOrderFood(fields []string, currentTime time.Time) {
	// seatID, _ := strconv.Atoi(fields[2])
	// foodID, _ := strconv.Atoi(fields[3])

	// TODO: ロジック実装
	// 1. seats[seatID] を確認
	// 2. StatusがOccupiedでない場合 -> seat not used
	// 3. StatusがOccupiedの場合:
	//    - seat.UserIDを使って session を取得
	//    - session.FoodOrders[foodID]++
	//    - fmt.Println("order-food: ok")
}

// shower-start: {時刻} {座席 ID}
func handleShowerStart(fields []string, currentTime time.Time) {
	// seatID, _ := strconv.Atoi(fields[2])

	// TODO: ロジック実装
	// 1. seats[seatID] を確認 -> seat not used
	// 2. seat.UserIDで session を取得
	// 3. session.CurrentShowerID != 0 の場合 -> already started
	// 4. 空きシャワーを探す (基本実装: Status == ShowerAvailable)
	// 5. 見つからない場合 -> fully occupied
	// 6. 見つかった場合 (showerID):
	//    - showers[showerID].Status = ShowerOccupied
	//    - showers[showerID].UserID = session.ID
	//    - session.CurrentShowerID = showerID
	//    - session.ShowerStartTime = currentTime
	//    - fmt.Printf("shower-start: %d\n", showerID)
}

// shower-end: {時刻} {座席 ID}
func handleShowerEnd(fields []string, currentTime time.Time) {
	// seatID, _ := strconv.Atoi(fields[2])

	// TODO: ロジック実装
	// 1. seats[seatID] を確認 -> seat not used
	// 2. seat.UserIDで session を取得
	// 3. session.CurrentShowerID == 0 の場合 -> not started
	// 4. 正常処理:
	//    - showerID := session.CurrentShowerID
	//    - 料金計算 (calculateShowerFee(session.ShowerStartTime, currentTime))
	//    - session.TotalShowerFee に加算
	//    - 利用時間計算 (get-duration と同様のロジック)
	//    - fmt.Printf("shower-end: %d\n", minutes)
	//    - session.CurrentShowerID = 0
	//    - showers[showerID].Status = ShowerAvailable (基本実装)
	//    - showers[showerID].UserID = 0
}

// get-vacant-seats: {時刻}
func handleGetVacantSeats(fields []string, currentTime time.Time) {
	// TODO: ロジック実装 (応用実装)
	// 1. 座席タイプごとに空席をカウント (m種類)
	// 2. (応用) `status == Cleaning` かつ `currentTime >= availableAt` も空席とみなす
	// 3. 結果をスライスに格納
	// 4. 指定フォーマットで出力
}
