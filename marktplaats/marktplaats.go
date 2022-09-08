package marktplaats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"encore.dev/beta/errs"
	"github.com/PuerkitoBio/goquery"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

const baseURL = "https://marktplaats.nl/lrp/api/search"

func extractCategoriesFromHtml(rawURL string) (map[string]int, error) {
	res, err := http.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errs.Wrap(err, "could not parse html to extract categories")
	}

	opts := doc.Find("select[name='categoryId']").Find("option")
	categories := map[string]int{}
	opts.Each(func(i int, selection *goquery.Selection) {
		attr, _ := selection.Attr("value")
		id, _ := strconv.Atoi(attr)
		key := strings.Replace(selection.Text(), " ", "-", -1)
		key = strings.ToLower(key)
		categories[key] = id
	})
	return categories, nil
}

func ParseURL(ctx context.Context, rawURL string) (*QueryRequest, error) {
	//  "https://www.marktplaats.nl/l/huis-en-inrichting/kachels/#q:zibro|f:31,32,4205|distanceMeters:50000|postcode:3461CC"
	uri, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	categories, err := extractCategoriesFromHtml(rawURL)
	if err != nil {
		return nil, err
	}
	attrs := map[string]interface{}{}

	// parse categories from path
	path := strings.Split(uri.Path, "/")
	if len(path) > 4 {
		attrs["category"] = categories[path[2]]
		attrs["sub_category"] = categories[path[3]]
	}

	// parse attributes from uri fragment
	for _, s := range strings.Split(uri.Fragment, "|") {
		x := strings.Split(s, ":")
		if len(x) == 2 {
			attrs[x[0]] = x[1]
		}
	}
	q := QueryRequest{}
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &q,
		DecodeHook:       mapstructure.StringToSliceHookFunc(","),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(attrs); err != nil {
		return nil, err
	}
	return &q, nil
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
	listings = lo.Filter(listings, func(l Listing, _ int) bool {
		return l.PriceInfo.PriceType != "RESERVED"
	})
	ads := lo.Map(listings, func(listing Listing, _ int) Advertisement {
		return Advertisement{
			ID:    listing.ItemId,
			Title: listing.Title,
			Location: Location{
				CityName: listing.Location.CityName,
			},
			PriceInfo: PriceInfo{
				PriceCents: listing.PriceInfo.PriceCents,
			},
			URL:         "https://marktplaats.nl" + listing.VipUrl,
			ImageUrls:   listing.ImageUrls,
			Description: listing.Description,
			Date:        listing.Date,
		}
	})

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
