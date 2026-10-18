package openpoiapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/yone-k/openpoiapi-go"
)

const emptySuggestBody = `{"count": 0, "scope": "view", "suggestions": [], "vocabulary": [], "completion": null}`

func TestSuggest(t *testing.T) {
	t.Run("GETで/v1/suggestにリクエストする", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySuggestBody)
		client := newTestClient(t, srv.URL)

		if _, err := client.Suggest(context.Background(), openpoiapi.SuggestParams{}); err != nil {
			t.Fatalf("Suggest() error = %v", err)
		}

		got := srv.LastRequest(t)
		if got.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", got.Method, http.MethodGet)
		}
		if got.Path != "/v1/suggest" {
			t.Errorf("path = %q, want %q", got.Path, "/v1/suggest")
		}
	})

	t.Run("条件を指定しないとクエリパラメータを送らない", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySuggestBody)
		client := newTestClient(t, srv.URL)

		if _, err := client.Suggest(context.Background(), openpoiapi.SuggestParams{}); err != nil {
			t.Fatalf("Suggest() error = %v", err)
		}

		if got := srv.LastRequest(t).Query; len(got) != 0 {
			t.Errorf("query = %v, want empty", got)
		}
	})

	t.Run("全項目を指定するとクエリパラメータとして送る", func(t *testing.T) {
		srv := newStubServer(t, http.StatusOK, emptySuggestBody)
		client := newTestClient(t, srv.URL)

		params := openpoiapi.SuggestParams{
			Q:      "すたーば",
			BBox:   &openpoiapi.BBox{MinLng: 139.6, MinLat: 35.5, MaxLng: 139.9, MaxLat: 35.8},
			Center: &openpoiapi.Point{Lng: 139.75, Lat: 35.66},
			Radius: 500,
			Limit:  5,
			Fields: openpoiapi.SuggestFieldsMinimal,
		}
		if _, err := client.Suggest(context.Background(), params); err != nil {
			t.Fatalf("Suggest() error = %v", err)
		}

		want := url.Values{
			"q":      {"すたーば"},
			"bbox":   {"139.6,35.5,139.9,35.8"},
			"center": {"139.75,35.66"},
			"radius": {"500"},
			"limit":  {"5"},
			"fields": {"minimal"},
		}
		if got := srv.LastRequest(t).Query; !reflect.DeepEqual(got, want) {
			t.Errorf("query = %v, want %v", got, want)
		}
	})

	validationCases := []struct {
		name      string
		params    openpoiapi.SuggestParams
		wantField string
	}{
		{name: "centerの緯度が90を超える", params: openpoiapi.SuggestParams{Center: &openpoiapi.Point{Lng: 0, Lat: 90.1}}, wantField: "center"},
		{name: "bboxの南端が北端より大きい", params: openpoiapi.SuggestParams{BBox: &openpoiapi.BBox{MinLng: 0, MinLat: 2, MaxLng: 1, MaxLat: 1}}, wantField: "bbox"},
		{name: "radiusが負数", params: openpoiapi.SuggestParams{Radius: -1}, wantField: "radius"},
	}
	for _, tc := range validationCases {
		t.Run(tc.name+"とSearchと同じくValidationErrorを返してリクエストを送らない", func(t *testing.T) {
			srv := newStubServer(t, http.StatusOK, emptySuggestBody)
			client := newTestClient(t, srv.URL)

			_, err := client.Suggest(context.Background(), tc.params)

			var validationErr *openpoiapi.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("Suggest() error = %v, want *ValidationError", err)
			}
			if validationErr.Field != tc.wantField {
				t.Errorf("Field = %q, want %q", validationErr.Field, tc.wantField)
			}
			if got := len(srv.Requests()); got != 0 {
				t.Errorf("requests = %d, want 0", got)
			}
		})
	}

	t.Run("200応答をSuggestResponseにデコードする", func(t *testing.T) {
		body := `{"count": 1, "scope": "nationwide", "suggestions": [{"name": "A", "address": "東京都", "lat": 35.6, "lng": 139.7, "level": 3}], "vocabulary": [], "completion": "A"}`
		srv := newStubServer(t, http.StatusOK, body)
		client := newTestClient(t, srv.URL)

		got, err := client.Suggest(context.Background(), openpoiapi.SuggestParams{Q: "A"})
		if err != nil {
			t.Fatalf("Suggest() error = %v", err)
		}

		want := &openpoiapi.SuggestResponse{
			Count:       1,
			Scope:       openpoiapi.ScopeNationwide,
			Suggestions: []openpoiapi.Suggestion{{Name: "A", Address: "東京都", Lat: 35.6, Lng: 139.7, Level: ptr(3)}},
			Vocabulary:  []openpoiapi.VocabularyItem{},
			Completion:  ptr("A"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Suggest() = %s, want %s", dump(got), dump(want))
		}
	})

	t.Run("400応答ではerrorをメッセージにしたAPIErrorを返す", func(t *testing.T) {
		srv := newStubServer(t, http.StatusBadRequest, `{"error":"bad"}`)
		client := newTestClient(t, srv.URL)

		_, err := client.Suggest(context.Background(), openpoiapi.SuggestParams{})

		var apiErr *openpoiapi.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("Suggest() error = %v, want *APIError", err)
		}
		want := openpoiapi.APIError{StatusCode: http.StatusBadRequest, Message: "bad"}
		if *apiErr != want {
			t.Errorf("APIError = %+v, want %+v", *apiErr, want)
		}
	})
}
