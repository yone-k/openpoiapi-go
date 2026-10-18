package openpoiapi_test

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/yone-k/openpoiapi-go"
)

const emptySearchBody = `{"count": 0, "results": []}`

func TestSearch(t *testing.T) {
	t.Run("GETで/v1/searchにリクエストする", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL)

		if _, err := client.Search(context.Background(), openpoiapi.SearchParams{}); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		got := srv.LastRequest(t)
		if got.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", got.Method, http.MethodGet)
		}
		if got.Path != "/v1/search" {
			t.Errorf("path = %q, want %q", got.Path, "/v1/search")
		}
	})

	t.Run("条件を指定しないとクエリパラメータを送らない", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL)

		if _, err := client.Search(context.Background(), openpoiapi.SearchParams{}); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if got := srv.LastRequest(t).Query; len(got) != 0 {
			t.Errorf("query = %v, want empty", got)
		}
	})

	t.Run("全項目を指定するとクエリパラメータとして送る", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL)

		params := openpoiapi.SearchParams{
			Q:      "世田谷区 カフェ",
			Center: &openpoiapi.Point{Lng: 139.75, Lat: 35.66},
			Radius: 1000,
			BBox:   &openpoiapi.BBox{MinLng: 139.6, MinLat: 35.5, MaxLng: 139.9, MaxLat: 35.8},
			Limit:  10,
		}
		if _, err := client.Search(context.Background(), params); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		want := url.Values{
			"q":      {"世田谷区 カフェ"},
			"center": {"139.75,35.66"},
			"radius": {"1000"},
			"bbox":   {"139.6,35.5,139.9,35.8"},
			"limit":  {"10"},
		}
		if got := srv.LastRequest(t).Query; !reflect.DeepEqual(got, want) {
			t.Errorf("query = %v, want %v", got, want)
		}
	})

	t.Run("座標を指数表記にせずに送る", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL)

		params := openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: 0.0000001, Lat: -35.5}}
		if _, err := client.Search(context.Background(), params); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if got := srv.LastRequest(t).Query.Get("center"); got != "0.0000001,-35.5" {
			t.Errorf("center = %q, want %q", got, "0.0000001,-35.5")
		}
	})

	limitCases := []struct {
		limit int
		want  string
	}{
		{limit: -1, want: "-1"},
		{limit: 500, want: "500"},
	}
	for _, tc := range limitCases {
		t.Run("範囲外のlimit「"+tc.want+"」もそのまま送る", func(t *testing.T) {
			srv := newStubServer(t, http.StatusOK, emptySearchBody)
			client := newTestClient(t, srv.URL)

			if _, err := client.Search(context.Background(), openpoiapi.SearchParams{Limit: tc.limit}); err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if got := srv.LastRequest(t).Query.Get("limit"); got != tc.want {
				t.Errorf("limit = %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("centerなしのradiusもそのまま送る", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL)

		if _, err := client.Search(context.Background(), openpoiapi.SearchParams{Radius: 1000}); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if got := srv.LastRequest(t).Query.Get("radius"); got != "1000" {
			t.Errorf("radius = %q, want %q", got, "1000")
		}
	})

	t.Run("既定のUser-AgentとAcceptヘッダを送る", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL)

		if _, err := client.Search(context.Background(), openpoiapi.SearchParams{}); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		header := srv.LastRequest(t).Header
		if got := header.Get("User-Agent"); got != "openpoiapi-go/0.1.0" {
			t.Errorf("User-Agent = %q, want %q", got, "openpoiapi-go/0.1.0")
		}
		if got := header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want %q", got, "application/json")
		}
	})

	t.Run("WithUserAgentで指定したUser-Agentを送る", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL, openpoiapi.WithUserAgent("myapp/1.0"))

		if _, err := client.Search(context.Background(), openpoiapi.SearchParams{}); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if got := srv.LastRequest(t).Header.Get("User-Agent"); got != "myapp/1.0" {
			t.Errorf("User-Agent = %q, want %q", got, "myapp/1.0")
		}
	})

	t.Run("WithBaseURLのパスの後ろにエンドポイントのパスを連結する", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL+"/prefix/")

		if _, err := client.Search(context.Background(), openpoiapi.SearchParams{}); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if got := srv.LastRequest(t).Path; got != "/prefix/v1/search" {
			t.Errorf("path = %q, want %q", got, "/prefix/v1/search")
		}
	})

	t.Run("WithHTTPClientで渡したHTTPクライアントを使う", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		var calls atomic.Int32
		hc := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			calls.Add(1)
			return http.DefaultTransport.RoundTrip(r)
		})}
		client := newTestClient(t, srv.URL, openpoiapi.WithHTTPClient(hc))

		if _, err := client.Search(context.Background(), openpoiapi.SearchParams{}); err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if got := calls.Load(); got != 1 {
			t.Errorf("transport calls = %d, want 1", got)
		}
	})

	t.Run("200応答をSearchResponseにデコードする", func(t *testing.T) {
		body := `{"count": 1, "results": [{"name": "A", "lat": "35.1", "lng": 139.7, "level": 8, "licenses": ["CC BY 4.0"]}]}`
		srv := newStubServer(t, http.StatusOK, body)
		client := newTestClient(t, srv.URL)

		got, err := client.Search(context.Background(), openpoiapi.SearchParams{})
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		want := &openpoiapi.SearchResponse{
			Count: 1,
			Results: []openpoiapi.Facility{
				{Name: "A", Lat: ptr(35.1), Lng: ptr(139.7), Level: ptr(8), Licenses: []string{"CC BY 4.0"}},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Search() = %s, want %s", dump(got), dump(want))
		}
	})

	validationCases := []struct {
		name      string
		params    openpoiapi.SearchParams
		wantField string
	}{
		{name: "centerの経度が180を超える", params: openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: 180.1, Lat: 0}}, wantField: "center"},
		{name: "centerの経度が-180未満", params: openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: -180.1, Lat: 0}}, wantField: "center"},
		{name: "centerの緯度が90を超える", params: openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: 0, Lat: 90.1}}, wantField: "center"},
		{name: "centerの緯度が-90未満", params: openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: 0, Lat: -90.1}}, wantField: "center"},
		{name: "centerの経度がNaN", params: openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: math.NaN(), Lat: 0}}, wantField: "center"},
		{name: "centerの緯度が+Inf", params: openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: 0, Lat: math.Inf(1)}}, wantField: "center"},
		{name: "bboxの西端が-180未満", params: openpoiapi.SearchParams{BBox: &openpoiapi.BBox{MinLng: -180.1, MinLat: 0, MaxLng: 1, MaxLat: 1}}, wantField: "bbox"},
		{name: "bboxの北端が90を超える", params: openpoiapi.SearchParams{BBox: &openpoiapi.BBox{MinLng: 0, MinLat: 0, MaxLng: 1, MaxLat: 90.1}}, wantField: "bbox"},
		{name: "bboxの西端が東端より大きい", params: openpoiapi.SearchParams{BBox: &openpoiapi.BBox{MinLng: 2, MinLat: 0, MaxLng: 1, MaxLat: 1}}, wantField: "bbox"},
		{name: "bboxの南端が北端より大きい", params: openpoiapi.SearchParams{BBox: &openpoiapi.BBox{MinLng: 0, MinLat: 2, MaxLng: 1, MaxLat: 1}}, wantField: "bbox"},
		{name: "bboxの東端がNaN", params: openpoiapi.SearchParams{BBox: &openpoiapi.BBox{MinLng: 0, MinLat: 0, MaxLng: math.NaN(), MaxLat: 1}}, wantField: "bbox"},
		{name: "radiusが負数", params: openpoiapi.SearchParams{Radius: -1}, wantField: "radius"},
		{
			name: "centerとbboxの両方が不正ならcenterを返す",
			params: openpoiapi.SearchParams{
				Center: &openpoiapi.Point{Lng: 200, Lat: 0},
				BBox:   &openpoiapi.BBox{MinLng: 2, MinLat: 0, MaxLng: 1, MaxLat: 1},
			},
			wantField: "center",
		},
	}
	for _, tc := range validationCases {
		t.Run(tc.name+"とValidationErrorを返してリクエストを送らない", func(t *testing.T) {
			srv := newStubServer(t, http.StatusOK, emptySearchBody)
			client := newTestClient(t, srv.URL)

			_, err := client.Search(context.Background(), tc.params)

			var validationErr *openpoiapi.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("Search() error = %v, want *ValidationError", err)
			}
			if validationErr.Field != tc.wantField {
				t.Errorf("Field = %q, want %q", validationErr.Field, tc.wantField)
			}
			if got := len(srv.Requests()); got != 0 {
				t.Errorf("requests = %d, want 0", got)
			}
		})
	}

	boundaryCases := []struct {
		name   string
		params openpoiapi.SearchParams
	}{
		{name: "centerが北東の端", params: openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: 180, Lat: 90}}},
		{name: "centerが南西の端", params: openpoiapi.SearchParams{Center: &openpoiapi.Point{Lng: -180, Lat: -90}}},
		{name: "bboxの西端と東端・南端と北端が等しい", params: openpoiapi.SearchParams{BBox: &openpoiapi.BBox{MinLng: 139.7, MinLat: 35.6, MaxLng: 139.7, MaxLat: 35.6}}},
	}
	for _, tc := range boundaryCases {
		t.Run(tc.name+"ならエラーにせずリクエストを送る", func(t *testing.T) {
			srv := newStubServer(t, http.StatusOK, emptySearchBody)
			client := newTestClient(t, srv.URL)

			if _, err := client.Search(context.Background(), tc.params); err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if got := len(srv.Requests()); got != 1 {
				t.Errorf("requests = %d, want 1", got)
			}
		})
	}

	apiErrorCases := []struct {
		name        string
		status      int
		body        string
		wantMessage string
	}{
		{name: "400のerrorをメッセージにする", status: http.StatusBadRequest, body: `{"error":"bad center"}`, wantMessage: "bad center"},
		{name: "429のmessageをメッセージにする", status: http.StatusTooManyRequests, body: `{"message":"Too Many Requests"}`, wantMessage: "Too Many Requests"},
		{name: "本文がJSONでなければステータスの説明をメッセージにする", status: http.StatusInternalServerError, body: `oops`, wantMessage: "Internal Server Error"},
		{name: "本文にerrorもmessageもなければステータスの説明をメッセージにする", status: http.StatusBadGateway, body: `{}`, wantMessage: "Bad Gateway"},
	}
	for _, tc := range apiErrorCases {
		t.Run("2xx以外の応答では"+tc.name+"APIErrorを返す", func(t *testing.T) {
			srv := newStubServer(t, tc.status, tc.body)
			client := newTestClient(t, srv.URL)

			_, err := client.Search(context.Background(), openpoiapi.SearchParams{})

			var apiErr *openpoiapi.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("Search() error = %v, want *APIError", err)
			}
			want := openpoiapi.APIError{StatusCode: tc.status, Message: tc.wantMessage}
			if *apiErr != want {
				t.Errorf("APIError = %+v, want %+v", *apiErr, want)
			}
		})
	}

	t.Run("200応答のJSONが不正ならAPIErrorでもValidationErrorでもないエラーを返す", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, `{"count":`)
		client := newTestClient(t, srv.URL)

		_, err := client.Search(context.Background(), openpoiapi.SearchParams{})
		if err == nil {
			t.Fatalf("Search() error = nil, want error")
		}

		var apiErr *openpoiapi.APIError
		if errors.As(err, &apiErr) {
			t.Errorf("Search() error = %v, want not *APIError", err)
		}
		var validationErr *openpoiapi.ValidationError
		if errors.As(err, &validationErr) {
			t.Errorf("Search() error = %v, want not *ValidationError", err)
		}
	})

	t.Run("キャンセル済みのcontextを渡すとcontext.Canceledを含むエラーを返す", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySearchBody)
		client := newTestClient(t, srv.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.Search(ctx, openpoiapi.SearchParams{})

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Search() error = %v, want context.Canceled", err)
		}
	})
}
