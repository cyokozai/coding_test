# クエリロジック設計図

1. シャワー利用のロジック (shower-start / shower-end)

シャワー利用は、利用客 (UserSession) と シャワーリソース (Shower) の2つの状態が連動して変化します。

handleShowerStart (シャワー使用開始)

利用客（座席IDで指定）がシャワーを使おうとする時の処理です。

```mermaid
flowchart TD
    A[shower-start: クエリ受信] --> B{座席IDは利用中か？}
    B -- No --> C["Error: seat not used"]
    B -- Yes --> D{"該当の利用客は<br/>既にシャワー利用中か？<br/>(session.CurrentShowerID != 0)"}
    D -- Yes --> E["Error: already started"]
    D -- No --> F{"空きシャワー(Available)は<br/>存在するか？"}
    F -- No --> G["Error: fully occupied"]
    F -- Yes --> H["空きシャワーのうち<br/>ID最小のもの (showerID) を取得"]
    H --> I[状態更新: シャワー]
    I --> J[状態更新: 利用客]
    J --> K["Output: shower-start: {showerID}"]

    subgraph 状態更新_シャワー
        direction LR
        S1("showers[showerID].Status = ShowerOccupied")
        S2("showers[showerID].UserID = session.ID")
    end
    
    subgraph 状態更新_利用客
        direction LR
        U1("session.CurrentShowerID = showerID")
        U2("session.ShowerStartTime = currentTime")
    end
```

handleShowerEnd (シャワー使用終了)

利用客がシャワー利用を終える時の処理です。シャワー料金の計算がここで行われます。

```mermaid
flowchart TD
    A[shower-end: クエリ受信] --> B{座席IDは利用中か？}
    B -- No --> C["Error: seat not used"]
    B -- Yes --> D{"該当の利用客は<br/>シャワーを利用中か？<br/>(session.CurrentShowerID == 0)"}
    D -- Yes --> E["Error: not started"]
    D -- No --> F["利用中のシャワーID取得<br/>(showerID = session.CurrentShowerID)"]
    F --> G["シャワー利用時間(分)を計算<br/>calcDurationMinutes(session.ShowerStartTime, currentTime)"]
    G --> H["シャワー料金を計算<br/>calculateShowerFee(利用秒数)"]
    H --> I[session.TotalShowerFee に料金を加算]
    I --> J[状態更新: シャワー]
    J --> K[状態更新: 利用客]
    K --> L["Output: shower-end: {利用時間(分)}"]

    subgraph 状態更新_シャワー_基本実装
        direction LR
        S1("showers[showerID].Status = ShowerAvailable")
        S2("showers[showerID].UserID = 0")
    end
    
    subgraph 状態更新_利用客
        direction LR
        U1("session.CurrentShowerID = 0")
    end
```

2. 精算処理のロジック (checkout)

最も複雑なクエリです。まず「エラーチェック」と「正常系の状態遷移」を整理し、次に「料金計算」を詳細化します。

handleCheckout (エラーチェックと状態遷移)

```mermaid
flowchart TD
    A[checkout: クエリ受信] --> B{"利用客IDは存在するか？<br/>(users[userID])"}
    B -- No --> C["Error: invalid user"]
    B -- Yes --> D{"既に精算済みか？<br/>(session.Status == UserCheckedOut)"}
    D -- Yes --> E["Error: already done"]
    D -- No --> F{"シャワー利用中か？<br/>(session.CurrentShowerID != 0)"}
    F -- Yes --> G["Error: shower is still in use"]
    F -- No --> H(応用実装: クーポンチェック)
    H -- 基本実装はパス --> I[料金計算実行]
    I --> J["座席料金計算<br/>calculateSeatFee(...)"]
    J --> K["フード料金計算<br/>calculateFoodFee(...)"]
    K --> L[合計金額 = 座席料金 + フード料金 + session.TotalShowerFee]
    L --> M[状態更新: 利用客]
    M --> N[状態更新: 座席]
    N --> O["Output: checkout: {合計金額}"]

    subgraph 状態更新_利用客
        direction LR
        U1("session.Status = UserCheckedOut")
    end

    subgraph 状態更新_座席_基本実装
        direction LR
        S1("seats[session.SeatID].Status = SeatAvailable")
        S2("seats[session.SeatID].UserID = 0")
    end
```

calculateSeatFee (座席料金計算：最小料金制)

handleCheckout の中核となるロジックです。

```mermaid
flowchart TD
    A["calculateSeatFee 呼び出し<br/>(session, checkoutTime)"] --> B["利用時間(秒)を計算<br/>durationSec = checkoutTime.Sub(session.CheckInTime).Seconds()"]
    B --> C["座席タイプ取得<br/>seatType = seatTypes[session.SeatTypeID]"]
    C --> D["基本料金を計算<br/>baseFee = calculateBasicFee(durationSec, seatType.BasicRate)"]
    D --> E[minFee = baseFee]
    E --> F{"利用可能なパックは存在するか？<br/>(len(seatType.Packs) > 0)"}
    F -- No --> L[return minFee]
    F -- Yes --> G[パックをループ処理]
    G --> H["packFee = calculatePackFee(durationSec, seatType.BasicRate, pack)"]
    H --> I{packFee < minFee ?}
    I -- Yes --> J[minFee = packFee]
    J --> K[次のパックへ]
    I -- No --> K
    K --> G
    G -- ループ終了 --> L
```

calculateBasicFee / calculateShowerFee (基本料金/シャワー料金の計算)

この2つのロジックは非常に似ていますが、加算タイミングと加算間隔が異なります。

座席基本料金: 1秒後に初回加算、以降**10分(600秒)**ごと

シャワー料金: 1秒後に初回加算、以降**15分(900秒)**ごと

```mermaid
flowchart TD
    A["計算開始 (利用秒数 durationSec, 料金 rate, 間隔 intervalSec)"] --> B{durationSec < 1 ?}
    B -- Yes --> C[return 0]
    B -- No --> D["初回加算 (count = 1)"]
    D --> E["超過秒数<br/>overSec = durationSec - 1"]
    E --> F["追加加算回数<br/>additionalCount = int(overSec / intervalSec)"]
    F --> G[totalCount = count + additionalCount]
    G --> H[totalFee = totalCount * rate]
    H --> I[return totalFee]
```

- calculateBasicFee 実行時: intervalSec = 600
- calculateShowerFee 実行時: intervalSec = 900
