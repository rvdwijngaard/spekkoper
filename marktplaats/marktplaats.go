package marktplaats

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"encore.dev/beta/errs"
	"github.com/samber/lo"
)

type QueryRequest struct {
	// Query defines the foo
	Query              string
	PostCode           string
	DistanceMeters     int
	Limit              int
	Offset             int
	IncludeCommercials bool
	Category           int
	SubCategory        int
}

const baseURL = "https://marktplaats.nl/lrp/api/search"

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
	ID        string
	Title     string
	Location  Location
	PriceInfo PriceInfo
	URL       string
}

type PriceInfo struct {
	PriceCents int
}

func Query(ctx context.Context, q QueryRequest) (*QueryResponse, error) {
	url, err := q.url()
	if err != nil {
		return nil, errs.Wrap(err, "could not get marktplaats query url")
	}
	res, err := fetch(url)
	if err != nil {
		return nil, errs.Wrap(err, "could not fetch marktplaats results")
	}
	listings := res.Listings

	if !q.IncludeCommercials {
		listings = lo.Filter(listings, func(l Listing, _ int) bool {
			return l.SellerInformation.SellerWebsiteUrl == "" || !l.SellerInformation.ShowWebsiteUrl
		})
	}

	ads := make([]Advertisement, len(listings))
	for i, listing := range listings {
		ads[i] = Advertisement{
			ID:    listing.ItemId,
			Title: listing.Title,
			Location: Location{
				CityName: listing.Location.CityName,
			},
			PriceInfo: PriceInfo{
				PriceCents: listing.PriceInfo.PriceCents,
			},
			URL: "https://marktplaats.nl" + listing.VipUrl,
		}
	}

	return &QueryResponse{Advertisements: ads}, nil
}

func fetch(url string) (*resultDto, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, errs.Wrap(err, "marktplaats query failure")
	}

	defer res.Body.Close()

	var payload resultDto

	b, err := io.ReadAll(res.Body)
	if err != nil {
		if err != nil {
			return nil, errs.Wrap(err, "marktplaats query failure")
		}
	}
	err = json.Unmarshal(b, &payload)
	if err != nil {
		return nil, errs.Wrap(err, "could not unmarshal query data from marktplaats")
	}
	return &payload, nil
}
