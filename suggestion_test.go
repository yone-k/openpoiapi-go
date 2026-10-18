package openpoiapi_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/yone-k/openpoiapi-go"
)

func TestSuggestResponse_UnmarshalJSON(t *testing.T) {
	t.Run("full応答の全フィールドが入る", func(t *testing.T) {
		data := `{
			"count": 2,
			"scope": "view",
			"suggestions": [
				{
					"name": "スターバックス 渋谷店",
					"address": "東京都渋谷区道玄坂1-1-1",
					"prefecture": "東京都",
					"city": "渋谷区",
					"category": "カフェ",
					"business_type": "カフェ",
					"lat": 35.658,
					"lng": 139.701,
					"level": 8,
					"source": "overture",
					"licenses": ["CDLA-Permissive-2.0"],
					"attributions": ["Overture Maps"],
					"matched": ["スターバ"]
				},
				{
					"name": "スターバックス 恵比寿店",
					"address": "東京都渋谷区恵比寿1-1-1",
					"lat": 35.646,
					"lng": 139.710,
					"level": null
				}
			],
			"vocabulary": [
				{
					"type": "place",
					"label": "渋谷区",
					"query": "渋谷区",
					"prefecture": "東京都",
					"city": "渋谷区",
					"center": [139.70, 35.66],
					"bbox": [139.66, 35.64, 139.72, 35.69],
					"count": 1200
				}
			],
			"completion": "スターバックス 渋谷店"
		}`

		var got openpoiapi.SuggestResponse
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		want := openpoiapi.SuggestResponse{
			Count: 2,
			Scope: openpoiapi.ScopeView,
			Suggestions: []openpoiapi.Suggestion{
				{
					Name:         "スターバックス 渋谷店",
					Address:      "東京都渋谷区道玄坂1-1-1",
					Prefecture:   "東京都",
					City:         "渋谷区",
					Category:     "カフェ",
					BusinessType: "カフェ",
					Lat:          35.658,
					Lng:          139.701,
					Level:        ptr(8),
					Source:       "overture",
					Licenses:     []string{"CDLA-Permissive-2.0"},
					Attributions: []string{"Overture Maps"},
					Matched:      []string{"スターバ"},
				},
				{
					Name:    "スターバックス 恵比寿店",
					Address: "東京都渋谷区恵比寿1-1-1",
					Lat:     35.646,
					Lng:     139.710,
				},
			},
			Vocabulary: []openpoiapi.VocabularyItem{
				{
					Type:       openpoiapi.VocabularyTypePlace,
					Label:      "渋谷区",
					Query:      "渋谷区",
					Prefecture: "東京都",
					City:       "渋谷区",
					Center:     &openpoiapi.Point{Lng: 139.70, Lat: 35.66},
					BBox:       &openpoiapi.BBox{MinLng: 139.66, MinLat: 35.64, MaxLng: 139.72, MaxLat: 35.69},
					Count:      1200,
				},
			},
			Completion: ptr("スターバックス 渋谷店"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SuggestResponse = %s, want %s", dump(got), dump(want))
		}
	})

	t.Run("completionがnullならnilになる", func(t *testing.T) {
		var got openpoiapi.SuggestResponse
		if err := json.Unmarshal([]byte(`{"count": 0, "scope": "view", "suggestions": [], "completion": null}`), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if got.Completion != nil {
			t.Errorf("Completion = %q, want nil", *got.Completion)
		}
	})

	t.Run("minimal応答では候補の省かれた項目がゼロ値でレスポンス直下にライセンスが入る", func(t *testing.T) {
		data := `{
			"count": 1,
			"scope": "nationwide",
			"suggestions": [
				{"name": "A", "address": "東京都千代田区", "prefecture": "東京都", "city": "千代田区", "lat": 35.68, "lng": 139.76, "level": 3, "matched": []}
			],
			"vocabulary": [],
			"completion": null,
			"licenses": ["CC BY 4.0"],
			"attributions": ["東京都"]
		}`

		var got openpoiapi.SuggestResponse
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		want := openpoiapi.SuggestResponse{
			Count: 1,
			Scope: openpoiapi.ScopeNationwide,
			Suggestions: []openpoiapi.Suggestion{
				{Name: "A", Address: "東京都千代田区", Prefecture: "東京都", City: "千代田区", Lat: 35.68, Lng: 139.76, Level: ptr(3), Matched: []string{}},
			},
			Vocabulary:   []openpoiapi.VocabularyItem{},
			Licenses:     []string{"CC BY 4.0"},
			Attributions: []string{"東京都"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SuggestResponse = %s, want %s", dump(got), dump(want))
		}
	})

	t.Run("仕様にないscopeとtypeの値をそのまま保持する", func(t *testing.T) {
		data := `{"count": 0, "scope": "world", "suggestions": [], "vocabulary": [{"type": "station", "label": "x", "query": "x"}]}`

		var got openpoiapi.SuggestResponse
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if got.Scope != "world" {
			t.Errorf("Scope = %q, want %q", got.Scope, "world")
		}
		if len(got.Vocabulary) != 1 || got.Vocabulary[0].Type != "station" {
			t.Errorf("Vocabulary = %s, want one item with type station", dump(got.Vocabulary))
		}
	})
}

func TestVocabularyItem_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name       string
		data       string
		wantCenter *openpoiapi.Point
		wantBBox   *openpoiapi.BBox
	}{
		{
			name:       "centerが3要素ならnilになる",
			data:       `{"type": "place", "label": "x", "query": "x", "center": [1, 2, 3], "bbox": [1, 2, 3, 4]}`,
			wantCenter: nil,
			wantBBox:   &openpoiapi.BBox{MinLng: 1, MinLat: 2, MaxLng: 3, MaxLat: 4},
		},
		{
			name:       "bboxが2要素ならnilになる",
			data:       `{"type": "place", "label": "x", "query": "x", "center": [1, 2], "bbox": [1, 2]}`,
			wantCenter: &openpoiapi.Point{Lng: 1, Lat: 2},
			wantBBox:   nil,
		},
		{
			name:       "centerに文字列の要素があればnilになる",
			data:       `{"type": "place", "label": "x", "query": "x", "center": ["1", 2]}`,
			wantCenter: nil,
			wantBBox:   nil,
		},
		{
			name:       "category語彙でcenterとbboxがなければnilになる",
			data:       `{"type": "category", "label": "ラーメン", "query": "ラーメン"}`,
			wantCenter: nil,
			wantBBox:   nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name+"（エラーにならない）", func(t *testing.T) {
			var got openpoiapi.VocabularyItem
			if err := json.Unmarshal([]byte(tc.data), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if !reflect.DeepEqual(got.Center, tc.wantCenter) {
				t.Errorf("Center = %s, want %s", dump(got.Center), dump(tc.wantCenter))
			}
			if !reflect.DeepEqual(got.BBox, tc.wantBBox) {
				t.Errorf("BBox = %s, want %s", dump(got.BBox), dump(tc.wantBBox))
			}
		})
	}

	t.Run("JSONの構文が不正ならエラーを返す", func(t *testing.T) {
		var got openpoiapi.VocabularyItem
		if err := json.Unmarshal([]byte(`{"type": `), &got); err == nil {
			t.Errorf("json.Unmarshal() error = nil, want error")
		}
	})
}
