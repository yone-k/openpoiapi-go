package openpoiapi_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/yone-k/openpoiapi-go"
)

func ExampleClient_Search() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"count": 1, "results": [{"name": "ラーメン一番", "address": "東京都世田谷区三軒茶屋1-1-1", "lat": 35.643, "lng": 139.670, "licenses": ["CC BY 4.0"], "attributions": ["東京都"]}]}`)
	}))
	defer srv.Close()

	client, err := openpoiapi.NewClient(openpoiapi.WithBaseURL(srv.URL))
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

	fmt.Println(resp.Count)
	for _, f := range resp.Results {
		fmt.Println(f.Name, f.Address)
		if f.Lat != nil && f.Lng != nil {
			fmt.Println(*f.Lat, *f.Lng)
		}
	}
	// Output:
	// 1
	// ラーメン一番 東京都世田谷区三軒茶屋1-1-1
	// 35.643 139.67
}

func ExampleClient_Suggest() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"count": 1, "scope": "view", "suggestions": [{"name": "スターバックス 渋谷店", "address": "東京都渋谷区道玄坂1-1-1", "lat": 35.658, "lng": 139.701, "level": 8}], "vocabulary": [{"type": "brand", "label": "スターバックス", "query": "スターバックス"}], "completion": "スターバックス 渋谷店", "licenses": ["CDLA-Permissive-2.0"], "attributions": ["Overture Maps"]}`)
	}))
	defer srv.Close()

	client, err := openpoiapi.NewClient(openpoiapi.WithBaseURL(srv.URL))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Suggest(context.Background(), openpoiapi.SuggestParams{
		Q:      "すたーば",
		BBox:   &openpoiapi.BBox{MinLng: 139.6, MinLat: 35.5, MaxLng: 139.9, MaxLat: 35.8},
		Fields: openpoiapi.SuggestFieldsMinimal,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Scope)
	for _, s := range resp.Suggestions {
		fmt.Println(s.Name, s.Lat, s.Lng)
	}
	for _, v := range resp.Vocabulary {
		fmt.Println(v.Type, v.Label)
	}
	if resp.Completion != nil {
		fmt.Println(*resp.Completion)
	}
	fmt.Println(resp.Licenses)
	// Output:
	// view
	// スターバックス 渋谷店 35.658 139.701
	// brand スターバックス
	// スターバックス 渋谷店
	// [CDLA-Permissive-2.0]
}
