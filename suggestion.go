package openpoiapi

import "encoding/json"

// Scope tells where Client.Suggest found its suggestions.
type Scope string

const (
	// ScopeView means the suggestions were found within the given area.
	ScopeView Scope = "view"
	// ScopeNationwide means nothing was found within the given area, so the
	// search was widened to the whole country.
	ScopeNationwide Scope = "nationwide"
)

// VocabularyType is the kind of a VocabularyItem.
type VocabularyType string

const (
	// VocabularyTypeCategory is a business category, such as ramen.
	VocabularyTypeCategory VocabularyType = "category"
	// VocabularyTypeBrand is a chain store brand, such as Starbucks.
	VocabularyTypeBrand VocabularyType = "brand"
	// VocabularyTypePlace is a prefecture or municipality.
	VocabularyTypePlace VocabularyType = "place"
)

// SuggestResponse is the response of Client.Suggest.
type SuggestResponse struct {
	// Count is the number of suggestions returned.
	Count int `json:"count"`
	// Scope tells whether the suggestions were found within the given area or
	// nationwide.
	Scope Scope `json:"scope"`
	// Suggestions are ordered best first and can be rendered as they are.
	Suggestions []Suggestion `json:"suggestions"`
	// Vocabulary lists up to five categories, brands and places that start
	// with the input, best first.
	Vocabulary []VocabularyItem `json:"vocabulary"`
	// Completion is ghost text that continues the input, or nil when there is
	// none.
	Completion *string `json:"completion"`
	// Licenses lists the licenses of all suggestions. It is set only when
	// Fields is SuggestFieldsMinimal.
	Licenses []string `json:"licenses"`
	// Attributions lists the attributions of all suggestions. It is set only
	// when Fields is SuggestFieldsMinimal.
	Attributions []string `json:"attributions"`
}

// Suggestion is a single suggestion returned by Client.Suggest. Facilities
// without coordinates are never suggested. When Fields is
// SuggestFieldsMinimal, Category, BusinessType, Source, Licenses and
// Attributions are left empty; use the fields of SuggestResponse instead.
type Suggestion struct {
	// Name is the facility name.
	Name string `json:"name"`
	// Address starts with the prefecture and can be displayed on one line.
	Address string `json:"address"`
	// Prefecture is the prefecture name.
	Prefecture string `json:"prefecture"`
	// City is the normalized municipality name.
	City string `json:"city"`
	// Category is the normalized category shared across data sources. It is
	// "unknown" when the data source has little category information.
	Category string `json:"category"`
	// BusinessType has the same value as Category and is kept for backward
	// compatibility.
	BusinessType string `json:"business_type"`
	// Lat is the WGS84 latitude.
	Lat float64 `json:"lat"`
	// Lng is the WGS84 longitude.
	Lng float64 `json:"lng"`
	// Level is the geocoding precision (1 = prefecture, 2 = municipality,
	// 3 = town block, 8 = lot), or nil when it is unknown.
	Level *int `json:"level"`
	// Source identifies the data source of the representative record, such
	// as "jff" or "overture".
	Source string `json:"source"`
	// Licenses lists the licenses of the source data.
	Licenses []string `json:"licenses"`
	// Attributions lists the attributions of the source data. Store them
	// together with the suggestion.
	Attributions []string `json:"attributions"`
	// Matched lists the input words contained in Name, as typed. Use it to
	// highlight the matched parts.
	Matched []string `json:"matched"`
}

// VocabularyItem is a category, brand or place that matches the input.
// Prefecture, City, Center, BBox and Count are set only for
// VocabularyTypePlace.
type VocabularyItem struct {
	// Type is the kind of the item.
	Type VocabularyType `json:"type"`
	// Label is the display name.
	Label string `json:"label"`
	// Query is the string to pass as Q when the item is selected.
	Query string `json:"query"`
	// Prefecture is the prefecture name of a place.
	Prefecture string `json:"prefecture"`
	// City is the municipality name of a place. It is empty for a prefecture.
	City string `json:"city"`
	// Center is the average location of the facilities in a place, or nil
	// when the API does not return a pair of numbers.
	Center *Point `json:"center"`
	// BBox covers the facilities in a place, or nil when the API does not
	// return four numbers.
	BBox *BBox `json:"bbox"`
	// Count is the number of facilities with coordinates in a place.
	Count int `json:"count"`
}

// UnmarshalJSON decodes a vocabulary item, converting center ([lng, lat])
// to *Point and bbox ([minLng, minLat, maxLng, maxLat]) to *BBox. Arrays
// with a different length or non-number elements become nil instead of
// causing an error.
func (v *VocabularyItem) UnmarshalJSON(data []byte) error {
	type vocabularyItem VocabularyItem
	aux := struct {
		*vocabularyItem
		Center json.RawMessage `json:"center"`
		BBox   json.RawMessage `json:"bbox"`
	}{vocabularyItem: (*vocabularyItem)(v)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	v.Center = nil
	if c, ok := parseNumbers(aux.Center, 2); ok {
		v.Center = &Point{Lng: c[0], Lat: c[1]}
	}
	v.BBox = nil
	if b, ok := parseNumbers(aux.BBox, 4); ok {
		v.BBox = &BBox{MinLng: b[0], MinLat: b[1], MaxLng: b[2], MaxLat: b[3]}
	}
	return nil
}

func parseNumbers(raw json.RawMessage, n int) ([]float64, bool) {
	// []float64 would silently decode null elements as 0, so elements are
	// decoded as any and checked one by one.
	var items []any
	if err := json.Unmarshal(raw, &items); err != nil || len(items) != n {
		return nil, false
	}
	numbers := make([]float64, n)
	for i, item := range items {
		number, ok := item.(float64)
		if !ok {
			return nil, false
		}
		numbers[i] = number
	}
	return numbers, true
}
