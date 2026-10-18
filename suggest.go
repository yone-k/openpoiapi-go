package openpoiapi

import (
	"context"
	"net/url"
)

// SuggestFields selects the shape of each suggestion returned by
// Client.Suggest.
type SuggestFields string

const (
	// SuggestFieldsFull returns every field of each suggestion. It is the API
	// default.
	SuggestFieldsFull SuggestFields = "full"
	// SuggestFieldsMinimal omits Category, BusinessType, Source, Licenses and
	// Attributions from each suggestion and returns Licenses and Attributions
	// once in SuggestResponse instead.
	SuggestFieldsMinimal SuggestFields = "minimal"
)

// SuggestParams are the conditions for Client.Suggest. Zero values are not
// sent, so the API defaults apply.
type SuggestParams struct {
	// Q is the partially typed keyword. Space-separated words are combined
	// with AND. When it is empty, the API returns no suggestions.
	Q string
	// BBox is the visible map area. Suggestions are searched within it first
	// and nationwide only when nothing is found.
	BBox *BBox
	// Center is the map center used to sort suggestions by distance.
	Center *Point
	// Radius is the distance from Center in meters, used only when BBox is
	// nil. It must not be negative. The API default is 50000.
	Radius int
	// Limit is the maximum number of suggestions. The API clamps it to 1-20
	// and defaults to 8.
	Limit int
	// Fields selects the shape of each suggestion. It is sent as is.
	Fields SuggestFields
}

// Suggest returns completion candidates for a partially typed keyword. It
// returns a *ValidationError without sending a request when Center, BBox or
// Radius is invalid, and an *APIError when the API responds with a non-2xx
// status.
func (c *Client) Suggest(ctx context.Context, params SuggestParams) (*SuggestResponse, error) {
	if err := validateLocation(params.Center, params.BBox, params.Radius); err != nil {
		return nil, err
	}
	query := url.Values{}
	setString(query, "q", params.Q)
	setBBox(query, "bbox", params.BBox)
	setPoint(query, "center", params.Center)
	setInt(query, "radius", params.Radius)
	setInt(query, "limit", params.Limit)
	setString(query, "fields", string(params.Fields))

	var resp SuggestResponse
	if err := c.get(ctx, "/v1/suggest", query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
