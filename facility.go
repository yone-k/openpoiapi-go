package openpoiapi

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
)

// SearchResponse is the response of Client.Search.
type SearchResponse struct {
	// Count is the number of facilities returned.
	Count int `json:"count"`
	// Results are the matched facilities.
	Results []Facility `json:"results"`
}

// Facility is a single facility returned by Client.Search. One record
// represents one business license or notification, so the same facility may
// appear more than once with different business types.
type Facility struct {
	// Name is the facility name.
	Name string `json:"name"`
	// NameKana is the facility name in katakana.
	NameKana string `json:"name_kana"`
	// Prefecture is the prefecture name.
	Prefecture string `json:"prefecture"`
	// City is the normalized municipality name.
	City string `json:"city"`
	// Address is the street address.
	Address string `json:"address"`
	// Category is the normalized category shared across data sources. It is
	// "unknown" when the data source has little category information.
	Category string `json:"category"`
	// BusinessType has the same value as Category and is kept for backward
	// compatibility.
	BusinessType string `json:"business_type"`
	// Lat is the WGS84 latitude, or nil when the facility has no coordinates
	// or the API returns a value that is not a finite number.
	Lat *float64 `json:"lat"`
	// Lng is the WGS84 longitude, or nil when the facility has no coordinates
	// or the API returns a value that is not a finite number.
	Lng *float64 `json:"lng"`
	// Level is the geocoding precision (1 = prefecture, 2 = municipality,
	// 3 = town block, 8 = lot), or nil when it is unknown or the API returns
	// a value that is not an integer.
	Level *int `json:"level"`
	// Source identifies the data source of the representative record, such
	// as "jff" or "overture".
	Source string `json:"source"`
	// Licenses lists the licenses of the source data.
	Licenses []string `json:"licenses"`
	// Attributions lists the attributions of the source data. Store them
	// together with the facility.
	Attributions []string `json:"attributions"`
}

// UnmarshalJSON decodes a facility, normalizing lat and lng (number or
// string) to *float64 and level (integer, string or null) to *int. Values
// that cannot be converted become nil instead of causing an error.
func (f *Facility) UnmarshalJSON(data []byte) error {
	type facility Facility
	aux := struct {
		*facility
		Lat   json.RawMessage `json:"lat"`
		Lng   json.RawMessage `json:"lng"`
		Level json.RawMessage `json:"level"`
	}{facility: (*facility)(f)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	f.Lat = parseFloat(aux.Lat)
	f.Lng = parseFloat(aux.Lng)
	f.Level = parseInt(aux.Level)
	return nil
}

func parseFloat(raw json.RawMessage) *float64 {
	text, ok := scalarText(raw)
	if !ok {
		return nil
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

func parseInt(raw json.RawMessage) *int {
	text, ok := scalarText(raw)
	if !ok {
		return nil
	}
	v, err := strconv.Atoi(text)
	if err != nil {
		return nil
	}
	return &v
}

func scalarText(raw json.RawMessage) (string, bool) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	// Decoding numbers as float64 would turn 8.0 into 8 and let parseInt
	// accept it; json.Number keeps the literal so non-integers are rejected.
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return "", false
	}
	switch x := v.(type) {
	case json.Number:
		return x.String(), true
	case string:
		return x, true
	default:
		return "", false
	}
}
