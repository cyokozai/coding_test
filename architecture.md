## 🚀 システム要件の分析と設計方針

### 1\. 概要 (Overall Goal)

  * **目的:** インターネットカフェの入退店、空席、料金を管理するシミュレーションプログラムを作成する。
  * **役割:** 時系列順に与えられるクエリ（イベント）を処理し、各クエリに対して指定された通りの標準出力を返す。
  * **実装戦略:**
    1.  **基本実装:** まず、清掃時間ゼロ、クーポンなし、空席表示なしの「基本実装」を完璧に動作させる。
    2.  **応用実装:** 基本実装が完了した後、清掃、クーポン、空席表示の機能を追加する。

-----

### 2\. 採点 (Grading Strategy)

  * **基本実装の保証:**
      * `get-vacant-seats`（空席状況表示）クエリは来ない。
      * `checkout`（精算）クエリのクーポンは `0` 個で固定。
      * `cleaning_duration_seat`（座席清掃時間）は `0`。
      * `cleaning_duration_shower`（シャワー清掃時間）は `0`。
  * **方針:** この保証を利用し、まずは清掃やクーポン関連のロジックを一切省いたシンプルなモデルで実装を進めます。

-----

### 3\. 詳細な仕様 (Core Entities & State)

システム全体で管理・保持すべき「状態（データ）」をGoの構造体として設計します。

#### `### 時刻` (Time)

  * **形式:** `YYYY/MM/DD-hh:mm:ss`
  * **Goでの扱い:** `time.Parse` を使用する。レイアウト文字列は `2006/01/02-15:04:05`。全ての時刻は `time.Time` オブジェクトとして扱う。
  * **期間計算:**
      * `time.Sub()` で `time.Duration` を取得する。
      * **分単位（切り上げ）:** `math.Ceil(duration.Seconds() / 60.0)` で計算する。

#### `### 座席` (Seats)

  * **管理対象:** `n` 個の座席 (ID: `1`～`n`)。
  * **状態:** 各座席は「空き(`Available`)」「使用中(`Occupied`)」「清掃中(`Cleaning`)」のいずれかの状態を持つ。
  * **データ構造案:**
    ```go
    type SeatStatus int
    const (
        SeatAvailable SeatStatus = iota
        SeatOccupied
        SeatCleaning // 応用実装用
    )

    type Seat struct {
        ID       int
        TypeID   int
        Status   SeatStatus
        UserID   int       // この座席を使っている利用客ID (0なら空席)
        AvailableAt time.Time // 清掃が完了し利用可能になる時刻 (応用実装用)
    }

    // システム全体で `[]Seat` (サイズ n+1) として持つ。
    // seats[1] ～ seats[n] を使用する。
    ```
  * **初期設定:** `n` と `type_i` 配列から `seats` スライスを初期化する。

#### `### 受付` (Check-in)

  * **管理対象:** 利用客 (ID: `1, 2, 3...`)
  * **状態:** 各利用客は「利用中(`InUse`)」「精算済み(`CheckedOut`)」の状態を持つ。
  * **データ構造案:**
    ```go
    type UserStatus int
    const (
        UserInUse UserStatus = iota
        UserCheckedOut
    )

    type UserSession struct {
        ID         int
        SeatID     int
        CheckInTime time.Time
        Status     UserStatus

        // 料金関連
        FoodOrders     map[int]int // フードID -> 注文回数
        TotalShowerFee int         // シャワー料金の累計

        // シャワー利用状況
        CurrentShowerID int       // 利用中のシャワーID (0なら不使用)
        ShowerStartTime time.Time // 現在のシャワー開始時刻
    }

    // システム全体で以下を持つ:
    // nextUserID int = 1
    // users map[int]*UserSession // キー: 利用客ID
    ```

#### `### 利用料金` (Fees) & `### 定額パック` (Packs) & `### 最小料金制` (Min. Fee System)

  * **管理対象:** 座席タイプごとの料金体系。
  * **データ構造案:**
    ```go
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
    ```
  * **料金計算ロジック (重要):**
    1.  **基本料金 (パックなし):**
          * 利用開始 `1` 秒後に 1回目。以降 `10` 分 (600秒) ごと。
          * `durationSec := checkoutTime.Sub(checkinTime).Seconds()`
          * `if durationSec < 1 { return 0 }`
          * `chargeCount := 1 + int((durationSec - 1) / 600)`
          * `fee := chargeCount * basicRate`
    2.  **パック適用料金:**
          * `overTimeMinutes := totalMinutes - pack.TimeMinutes`
          * `overTimeSeconds := overTimeMinutes * 60`
          * 上記「基本料金」ロジックを `overTimeSeconds` に適用して超過料金を計算。
          * `fee := pack.Price + overageFee`
    3.  **最小料金:** `checkout` 時に、(1)の料金と、全パック適用時の(2)の料金をすべて計算し、その最小値を `SeatFee` とする。

#### `### フードの注文` (Food Orders)

  * **管理対象:** フード `f` 種類 (ID: `1`～`f`)
  * **データ構造案:**
    ```go
    // システム全体で `map[int]int` (キー: フードID, 値: 価格) として持つ。
    // foodPrices[1] ～ foodPrices[f]
    ```
  * **ロジック:** `order-food` クエリ時、該当 `UserSession` の `FoodOrders[foodID]` をインクリメントする。

#### `### シャワールームの使用` (Showers)

  * **管理対象:** `s` 個のシャワールーム (ID: `1`～`s`)
  * **状態:** 各シャワーは「空き(`Available`)」「使用中(`Occupied`)」「清掃中(`Cleaning`)」の状態を持つ。
  * **データ構造案:**
    ```go
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
    ```
  * **料金計算ロジック:**
      * 座席の基本料金計算ロジックとほぼ同じ。
      * 利用開始 `1` 秒後に 1回目。以降 `15` 分 (900秒) ごと。
      * `durationSec := endTime.Sub(startTime).Seconds()`
      * `if durationSec < 1 { return 0 }`
      * `chargeCount := 1 + int((durationSec - 1) / 900)`
      * `fee := chargeCount * showerCharge`
      * この `fee` を `UserSession.TotalShowerFee` に加算する。

#### `### 料金の精算` (Checkout)

  * **ロジック:**
    1.  `userID` で `UserSession` を引く。
    2.  エラーチェック（存在、精算済み、シャワー使用中）。
    3.  `SeatFee` を「最小料金制」ロジックで計算。
    4.  `FoodFee` を `UserSession.FoodOrders` と `foodPrices` から計算。
    5.  `ShowerFee` は `UserSession.TotalShowerFee` を使う。
    6.  `Total = SeatFee + FoodFee + ShowerFee`
    7.  `UserSession.Status` を `CheckedOut` に変更。
    8.  `Seat.Status` を `Available` (基本実装) / `Cleaning` (応用実装) に変更。

-----

### 4\. 応用実装の考慮 (Future Work)

以下の項目は、まず基本実装（`main_basic.go`）を作成した後に、それをコピーして（`main_advanced.go`）機能追加します。

  * **`### フードクーポン対応` (Coupons):**
      * **データ構造:** `map[int]Coupon` (キー: クーポンID)
      * `checkout` 処理を拡張。
          * エラーチェック（注文してないフードのクーポン）
          * 割引ロジック（フード毎に最大割引を適用）
  * **`### 空席状況の閲覧` (Vacant Seats):**
      * `get-vacant-seats` クエリハンドラを追加。
      * 全座席をループし、`seat.Status == SeatAvailable` OR (`seat.Status == SeatCleaning` AND `queryTime >= seat.AvailableAt`) の座席をカウントする。
  * **清掃 (Cleaning):**
      * `checkout` 時: `seat.Status` を `Cleaning` に設定し、`seat.AvailableAt = queryTime.Add(cleaning_duration_seat * time.Minute)` を設定する。
      * `shower-end` 時: 同様に `shower.Status` を `Cleaning` に設定し、`shower.AvailableAt` を設定する。
      * `checkin` 時: 「空き」の定義を `status == Available` OR (`status == Cleaning` AND `currentTime >= availableAt`) に変更する。
      * `shower-start` 時: 同上。

-----

### 5\. 実装ステップ (Implementation Plan)

1.  **ステップ 0: 構造体の定義と初期化**
      * `main` 関数を作成。
      * `bufio.Scanner` で標準入力を読み込む準備。
      * 上記の `Seat`, `SeatType`, `Pack`, `Shower`, `UserSession` などの構造体を定義する。
      * 入力の `n`, `m`, `f`, `s`, `q` などのヘッダー部分をすべて読み込み、`seats []Seat`, `seatTypes map[int]SeatType`, `foodPrices map[int]int`, `showers []Shower` などのグローバルな状態管理変数を初期化する。
2.  **ステップ 1: クエリの解析**
      * `q` 回のクエリをループで回す。
      * 各行を `strings.Split` で分割し、`query[0]` (`checkin:`, `order-food:` など) で処理を分岐する `switch` 文を作成する。
      * 時刻パーサー `parseTime(string) time.Time` ヘルパー関数を作成する。
3.  **ステップ 2: 基本実装クエリの逐次実装**
      * まずは `checkin:` を実装。
      * 次に `get-duration:` を実装。
      * `order-food:` を実装。
      * `shower-start:` と `shower-end:` (料金計算含む) を実装。
      * 最後に `checkout:` (料金計算含む) を実装。
4.  **ステップ 3: 基本実装のテスト**
      * `00_sample_00.in` の入出力が一致することを確認する。
5.  **ステップ 4: 応用実装**
      * （ファイルコピー後）清掃ロジック、クーポンロジック、`get-vacant-seats` を追加実装する。
