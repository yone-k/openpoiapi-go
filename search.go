package openpoiapi

import (
	"context"
	"net/url"
)

// SearchParams are the conditions for Client.Search. Zero values are not
// sent, so the API defaults apply.
type SearchParams struct {
	// Q is the search keyword. Space-separated words are combined with OR.
	Q string
	// Center limits results to Radius around it and sorts them by distance.
	Center *Point
	// Radius is the distance from Center in meters. It must not be negative.
	// The API default is 50000.
	Radius int
	// BBox limits results to the area. It takes precedence over Center and
	// Radius.
	BBox *BBox
	// Limit is the maximum number of results. The API clamps it to 1-200 and
	// defaults to 50.
	Limit int
}

// Search searches facilities by keyword, location and area. It returns a
// *ValidationError without sending a request when Center, BBox or Radius is
// invalid, and an *APIError when the API responds with a non-2xx status.
func (c *Client) Search(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	if err := validateLocation(params.Center, params.BBox, params.Radius); err != nil {
		return nil, err
	}
	query := url.Values{}
	setString(query, "q", params.Q)
	setPoint(query, "center", params.Center)
	setInt(query, "radius", params.Radius)
	setBBox(query, "bbox", params.BBox)
	setInt(query, "limit", params.Limit)

	var resp SearchResponse
	if err := c.get(ctx, "/v1/search", query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
