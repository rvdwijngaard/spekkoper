package marktplaats

import (
	"net/url"
	"strconv"
)

type QueryRequest struct {
	Query              string `mapstructure:"q"`
	PostCode           string `mapstructure:"postcode"`
	DistanceMeters     int    `mapstructure:"distanceMeters"`
	AttributesByID     []int  `mapstructure:"f"`
	Category           int    `mapstructure:"category"`
	SubCategory        int    `mapstructure:"sub_category"`
	Limit              int
	Offset             int
	IncludeCommercials bool
}

func (qr QueryRequest) url() (string, error) {
	uri, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("searchInTitleAndDescription", "true")
	params.Add("viewOptions", "list-view")
	limit := 30
	if qr.Limit > 0 {
		limit = qr.Limit
	}
	params.Add("limit", strconv.Itoa(limit))
	offset := 0
	if qr.Offset > 0 {
		offset = qr.Offset
	}
	params.Add("offset", strconv.Itoa(offset))
	if qr.Query != "" {
		params.Add("query", qr.Query)
	}
	if qr.PostCode != "" {
		params.Add("postcode", qr.PostCode)
	}
	if qr.DistanceMeters > 0 {
		params.Add("distanceMeters", strconv.Itoa(qr.DistanceMeters))
	}
	if qr.Category > 0 {
		params.Add("l1CategoryId", strconv.Itoa(qr.Category))
	}
	if qr.SubCategory > 0 {
		params.Add("l2CategoryId", strconv.Itoa(qr.SubCategory))
	}
	for _, attr := range qr.AttributesByID {
		params.Add("attributesById[]", strconv.Itoa(attr))
	}

	uri.RawQuery = params.Encode()
	return uri.String(), nil
}

func (qr QueryRequest) Validate() error {
	return nil
}

type QueryResponse struct {
	Advertisements []Advertisement `json:"advertisements"`
}

type Location struct {
	CityName string
}

type Advertisement struct {
	ID          string
	Title       string
	Location    Location
	PriceInfo   PriceInfo
	URL         string
	ImageUrls   []string
	Description string
}

type PriceInfo struct {
	PriceCents int
}
