# openpoiapi-go

[OpenPOI API](https://docs.openpoiapi.com/)（日本全国のPOI検索API）を Go から呼び出すための SDK です。施設検索（`GET /v1/search`）と入力補完（`GET /v1/suggest`）を型付きで利用できます。外部依存はありません。

API の詳細は公式ドキュメント https://docs.openpoiapi.com/ を参照してください。

## 動作要件

- Go 1.26 以上

## インストール

```sh
go get github.com/yone-k/openpoiapi-go
```

## 使い方

`NewClient` でクライアントを作り、`Search` または `Suggest` を呼び出します。API に認証は不要です。

### 施設を検索する

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yone-k/openpoiapi-go"
)

func main() {
	client, err := openpoiapi.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Search(context.Background(), openpoiapi.SearchParams{
		Q:      "ラーメン",
		Center: &openpoiapi.Point{Lng: 139.67, Lat: 35.64},
		Radius: 1000,
		Limit:  10,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range resp.Results {
		fmt.Println(f.Name, f.Address)
	}
}
```

- 座標は `Point{Lng, Lat}`（経度が先）、範囲は `BBox{MinLng, MinLat, MaxLng, MaxLat}` で指定します。API と同じく経度・緯度の順です。
- `SearchParams` と `SuggestParams` のゼロ値の項目（空文字、`0`、`nil`）は送信されず、API の既定値が使われます。
- `Center` を指定すると、`Radius`（メートル）の範囲に絞り、中心に近い順に並びます。`BBox` を指定すると、`Center` と `Radius` より優先されます。

### 入力補完の候補を取得する

```go
resp, err := client.Suggest(context.Background(), openpoiapi.SuggestParams{
	Q:      "すたーば",
	BBox:   &openpoiapi.BBox{MinLng: 139.6, MinLat: 35.5, MaxLng: 139.9, MaxLat: 35.8},
	Fields: openpoiapi.SuggestFieldsMinimal,
})
if err != nil {
	log.Fatal(err)
}

for _, s := range resp.Suggestions {
	fmt.Println(s.Name, s.Lat, s.Lng)
}
```

- 候補はまず `BBox` の中から探し、0件のときだけ全国に広げます。どちらで見つかったかは `resp.Scope`（`ScopeView` / `ScopeNationwide`）で分かります。
- `Fields: openpoiapi.SuggestFieldsMinimal` を指定すると、候補ごとの `Category`・`BusinessType`・`Source`・`Licenses`・`Attributions` が省かれた軽量なレスポンスになります。ライセンスと帰属表示は `resp.Licenses` と `resp.Attributions` に1回だけ入ります。
- `resp.Vocabulary` にはカテゴリ・ブランド・地名の語彙が入ります。地名（`VocabularyTypePlace`）の `Center` と `BBox` は、そのまま `SearchParams` に渡せます。

## 設定オプション

`NewClient` に次のオプションを渡せます。

| オプション | 内容 |
|---|---|
| `WithHTTPClient(hc *http.Client)` | リクエストに使う HTTP クライアントを指定します。タイムアウトやプロキシを変えたいときに使います。`nil` はエラーになります。 |
| `WithBaseURL(rawURL string)` | 接続先を変えます。http または https の絶対 URL で、クエリとフラグメントを含まないものだけを受け付けます。 |
| `WithUserAgent(ua string)` | `User-Agent` ヘッダを置き換えます。空文字はエラーになります。 |

既定では、接続先は `https://api.openpoiapi.com`、HTTP クライアントのタイムアウトは30秒、`User-Agent` は `openpoiapi-go/<バージョン>` です。リクエストごとの待ち時間は `context` でも制御できます。

## エラー処理

`Search` と `Suggest` は次のエラーを返します。`errors.As` で種類を判定してください。

- `*openpoiapi.ValidationError`: 送信前の検証に失敗したときに返します。このときリクエストは送信されません。`Field` は `"center"`・`"bbox"`・`"radius"` のいずれかです。検証内容は次のとおりで、Search と Suggest で共通です。
  - `Center`: 経度が -180〜180、緯度が -90〜90 の有限値であること
  - `BBox`: 4つの値が上記の範囲の有限値で、`MinLng <= MaxLng` かつ `MinLat <= MaxLat` であること
  - `Radius`: 負数でないこと
- `*openpoiapi.APIError`: API が 2xx 以外を返したときに返します。`StatusCode` に HTTP ステータス、`Message` に API のエラーメッセージが入ります。
- それ以外: 通信の失敗（`context` のキャンセルを含む）や、レスポンスのデコード失敗です。

API のレート制限（定常 200 req/s、バースト 600 req/s）を超えると 429 が返ります。SDK は自動でリトライしないため、再試行するかどうかと、その間隔は呼び出し側で決めてください。

```go
resp, err := client.Search(ctx, params)
if err != nil {
	var validationErr *openpoiapi.ValidationError
	var apiErr *openpoiapi.APIError
	switch {
	case errors.As(err, &validationErr):
		log.Printf("入力が不正です（%s）: %s", validationErr.Field, validationErr.Reason)
	case errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests:
		log.Print("レート制限に達しました。時間をおいて再試行してください")
	default:
		log.Print(err)
	}
	return
}
```

## 座標と精度の値

API は施設の緯度・経度を数値または文字列で返し、座標がない施設には空文字を返します。SDK はこれを次のように変換します。

- `Facility.Lat` と `Facility.Lng`（`*float64`）: 数値と、数値として読める文字列はその値になります。空文字・`null`・数値として読めない値は `nil` になります。
- `Facility.Level`（`*int`）: ジオコーディングの精度です（1=都道府県、2=市区町村、3=町丁目、8=街区・地番）。整数と、整数として読める文字列はその値になります。それ以外は `nil` になります。
- `VocabularyItem.Center` と `VocabularyItem.BBox`: API が返す配列の要素数がそれぞれ2・4で、すべて数値の場合だけ値が入ります。それ以外は `nil` になります。

変換できない値があってもエラーにはせず、ほかの項目は通常どおり返します。

## 出典表示

API が返すデータのライセンスは、データの提供元ごとに異なります。

- 施設や候補のデータを保存するときは、`Licenses` と `Attributions` も一緒に保存してください。
- データを表示するときは「OpenPOI API」と記載し、https://openpoiapi.com/attribution.html へリンクしてください。

## ライセンス

この SDK は [MIT License](LICENSE) で提供します。API が返すデータのライセンスは提供元ごとに異なります（[出典表示](#出典表示)を参照）。
