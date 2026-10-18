package openpoiapi_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/yone-k/openpoiapi-go"
)

func TestFacility_UnmarshalJSON(t *testing.T) {
	t.Run("文字列のフィールドと配列が対応するフィールドに入る", func(t *testing.T) {
		data := `{
			"name": "ラーメン一番",
			"name_kana": "ラーメンイチバン",
			"prefecture": "東京都",
			"city": "世田谷区",
			"address": "東京都世田谷区三軒茶屋1-1-1",
			"category": "ラーメン",
			"business_type": "ラーメン",
			"lat": 35.66,
			"lng": 139.75,
			"level": 8,
			"source": "jff",
			"licenses": ["CC BY 4.0"],
			"attributions": ["東京都"]
		}`

		var got openpoiapi.Facility
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		want := openpoiapi.Facility{
			Name:         "ラーメン一番",
			NameKana:     "ラーメンイチバン",
			Prefecture:   "東京都",
			City:         "世田谷区",
			Address:      "東京都世田谷区三軒茶屋1-1-1",
			Category:     "ラーメン",
			BusinessType: "ラーメン",
			Lat:          ptr(35.66),
			Lng:          ptr(139.75),
			Level:        ptr(8),
			Source:       "jff",
			Licenses:     []string{"CC BY 4.0"},
			Attributions: []string{"東京都"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Facility = %s, want %s", dump(got), dump(want))
		}
	})

	coordinateCases := []struct {
		name string
		raw  string
		want *float64
	}{
		{name: "数値はその値になる", raw: `35.66`, want: ptr(35.66)},
		{name: "数値の文字列は数値に変換される", raw: `"35.66"`, want: ptr(35.66)},
		{name: "空文字はnilになる", raw: `""`, want: nil},
		{name: "nullはnilになる", raw: `null`, want: nil},
		{name: "数値にできない文字列はnilになる", raw: `"abc"`, want: nil},
		{name: "NaNの文字列はnilになる", raw: `"NaN"`, want: nil},
		{name: "真偽値はnilになる", raw: `true`, want: nil},
	}
	for _, tc := range coordinateCases {
		t.Run("lat・lngの"+tc.name+"（エラーにならない）", func(t *testing.T) {
			data := `{"lat": ` + tc.raw + `, "lng": ` + tc.raw + `}`

			var got openpoiapi.Facility
			if err := json.Unmarshal([]byte(data), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if !reflect.DeepEqual(got.Lat, tc.want) {
				t.Errorf("Lat = %s, want %s", dump(got.Lat), dump(tc.want))
			}
			if !reflect.DeepEqual(got.Lng, tc.want) {
				t.Errorf("Lng = %s, want %s", dump(got.Lng), dump(tc.want))
			}
		})
	}

	t.Run("lat・lngのキーがないときはnilになる", func(t *testing.T) {
		var got openpoiapi.Facility
		if err := json.Unmarshal([]byte(`{"name": "x"}`), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if got.Lat != nil || got.Lng != nil {
			t.Errorf("Lat, Lng = %s, %s, want nil, nil", dump(got.Lat), dump(got.Lng))
		}
	})

	levelCases := []struct {
		name string
		raw  string
		want *int
	}{
		{name: "整数はその値になる", raw: `8`, want: ptr(8)},
		{name: "整数の文字列は整数に変換される", raw: `"3"`, want: ptr(3)},
		{name: "nullはnilになる", raw: `null`, want: nil},
		{name: "空文字はnilになる", raw: `""`, want: nil},
		{name: "整数にできない文字列はnilになる", raw: `"x"`, want: nil},
		{name: "小数はnilになる", raw: `2.5`, want: nil},
	}
	for _, tc := range levelCases {
		t.Run("levelの"+tc.name+"（エラーにならない）", func(t *testing.T) {
			var got openpoiapi.Facility
			if err := json.Unmarshal([]byte(`{"level": `+tc.raw+`}`), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if !reflect.DeepEqual(got.Level, tc.want) {
				t.Errorf("Level = %s, want %s", dump(got.Level), dump(tc.want))
			}
		})
	}

	t.Run("JSONの構文が不正ならエラーを返す", func(t *testing.T) {
		var got openpoiapi.Facility
		if err := json.Unmarshal([]byte(`{"name": `), &got); err == nil {
			t.Errorf("json.Unmarshal() error = nil, want error")
		}
	})
}

func TestSearchResponse_UnmarshalJSON(t *testing.T) {
	t.Run("countとresultsが入る", func(t *testing.T) {
		data := `{"count": 2, "results": [{"name": "A", "lat": "35.1"}, {"name": "B", "lat": ""}]}`

		var got openpoiapi.SearchResponse
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		want := openpoiapi.SearchResponse{
			Count: 2,
			Results: []openpoiapi.Facility{
				{Name: "A", Lat: ptr(35.1)},
				{Name: "B"},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SearchResponse = %s, want %s", dump(got), dump(want))
		}
	})
}
