package marktplaats

import "time"

type Listing struct {
	ItemId      string `json:"itemId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PriceInfo   struct {
		PriceCents int    `json:"priceCents"`
		PriceType  string `json:"priceType"`
	} `json:"priceInfo"`
	Location struct {
		CityName            string  `json:"cityName"`
		CountryName         string  `json:"countryName"`
		CountryAbbreviation string  `json:"countryAbbreviation"`
		DistanceMeters      int     `json:"distanceMeters"`
		IsBuyerLocation     bool    `json:"isBuyerLocation"`
		OnCountryLevel      bool    `json:"onCountryLevel"`
		Abroad              bool    `json:"abroad"`
		Latitude            float64 `json:"latitude"`
		Longitude           float64 `json:"longitude"`
	} `json:"location"`
	Date              time.Time `json:"date"`
	ImageUrls         []string  `json:"imageUrls"`
	SellerInformation struct {
		SellerId         int    `json:"sellerId"`
		SellerName       string `json:"sellerName"`
		ShowSoiUrl       bool   `json:"showSoiUrl"`
		ShowWebsiteUrl   bool   `json:"showWebsiteUrl"`
		IsVerified       bool   `json:"isVerified"`
		SellerWebsiteUrl string `json:"sellerWebsiteUrl"`
	} `json:"sellerInformation"`
	CategoryId           int    `json:"categoryId"`
	PriorityProduct      string `json:"priorityProduct"`
	VideoOnVip           bool   `json:"videoOnVip"`
	UrgencyFeatureActive bool   `json:"urgencyFeatureActive"`
	NapAvailable         bool   `json:"napAvailable"`
	Attributes           []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes"`
	Traits    []string `json:"traits"`
	Verticals []string `json:"verticals"`
	Pictures  []struct {
		Id                 int64  `json:"id"`
		ExtraSmallUrl      string `json:"extraSmallUrl"`
		MediumUrl          string `json:"mediumUrl"`
		LargeUrl           string `json:"largeUrl"`
		ExtraExtraLargeUrl string `json:"extraExtraLargeUrl"`
		AspectRatio        struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"aspectRatio"`
	} `json:"pictures"`
	VipUrl string `json:"vipUrl"`
}

type resultDto struct {
	Listings []Listing     `json:"listings"`
	TopBlock []interface{} `json:"topBlock"`
	Facets   []struct {
		Key        string `json:"key"`
		Type       string `json:"type"`
		Categories []struct {
			Id               int         `json:"id"`
			Selected         bool        `json:"selected"`
			IsValuableForSeo bool        `json:"isValuableForSeo"`
			Dominant         bool        `json:"dominant"`
			Label            string      `json:"label"`
			Key              string      `json:"key"`
			ParentId         *int        `json:"parentId"`
			ParentKey        interface{} `json:"parentKey"`
			HistogramCount   int         `json:"histogramCount,omitempty"`
		} `json:"categories,omitempty"`
		Id             int    `json:"id,omitempty"`
		Label          string `json:"label,omitempty"`
		AttributeGroup []struct {
			AttributeValueKey   string `json:"attributeValueKey"`
			AttributeValueId    int    `json:"attributeValueId,omitempty"`
			AttributeValueLabel string `json:"attributeValueLabel,omitempty"`
			Selected            bool   `json:"selected"`
			IsValuableForSeo    bool   `json:"isValuableForSeo"`
			HistogramCount      int    `json:"histogramCount,omitempty"`
			Default             bool   `json:"default,omitempty"`
		} `json:"attributeGroup,omitempty"`
		SingleSelect bool `json:"singleSelect,omitempty"`
		CategoryId   int  `json:"categoryId,omitempty"`
	} `json:"facets"`
	TotalResultCount  int    `json:"totalResultCount"`
	CorrelationId     string `json:"correlationId"`
	SuggestedQuery    string `json:"suggestedQuery"`
	OriginalQuery     string `json:"originalQuery"`
	SuggestedSearches []struct {
		Filters struct {
			Query struct {
				Text        string `json:"text"`
				DisplayText string `json:"displayText"`
			} `json:"query"`
			Categories []struct {
				Id           int    `json:"id"`
				CategoryName string `json:"categoryName"`
				ParentId     int    `json:"parentId"`
				ParentName   string `json:"parentName"`
			} `json:"categories"`
		} `json:"filters"`
	} `json:"suggestedSearches"`
	SortOptions []struct {
		SortBy    string `json:"sortBy"`
		SortOrder string `json:"sortOrder"`
	} `json:"sortOptions"`
	IsSearchSaved      bool          `json:"isSearchSaved"`
	HasErrors          bool          `json:"hasErrors"`
	AlternativeLocales []interface{} `json:"alternativeLocales"`
	SearchRequest      struct {
		OriginalRequest struct {
			Categories struct {
				L1Category struct {
					Id       int    `json:"id"`
					Key      string `json:"key"`
					FullName string `json:"fullName"`
				} `json:"l1Category"`
				L2Category struct {
					Id       int    `json:"id"`
					Key      string `json:"key"`
					FullName string `json:"fullName"`
				} `json:"l2Category"`
			} `json:"categories"`
			SearchQuery string `json:"searchQuery"`
			Attributes  struct {
			} `json:"attributes"`
			AttributesById  []interface{} `json:"attributesById"`
			AttributesByKey []interface{} `json:"attributesByKey"`
			AttributeRanges []interface{} `json:"attributeRanges"`
			AttributeLabels []interface{} `json:"attributeLabels"`
			SortOptions     struct {
				SortBy        string `json:"sortBy"`
				SortOrder     string `json:"sortOrder"`
				SortAttribute string `json:"sortAttribute"`
			} `json:"sortOptions"`
			Pagination struct {
				Offset int `json:"offset"`
				Limit  int `json:"limit"`
			} `json:"pagination"`
			Distance struct {
				Postcode       string `json:"postcode"`
				DistanceMeters int    `json:"distanceMeters"`
			} `json:"distance"`
			ViewOptions struct {
				Kind string `json:"kind"`
			} `json:"viewOptions"`
			SearchInTitleAndDescription bool `json:"searchInTitleAndDescription"`
			BypassSpellingSuggestion    bool `json:"bypassSpellingSuggestion"`
		} `json:"originalRequest"`
		Categories struct {
			L1Category struct {
				Id       int    `json:"id"`
				Key      string `json:"key"`
				FullName string `json:"fullName"`
			} `json:"l1Category"`
			L2Category struct {
				Id       int    `json:"id"`
				Key      string `json:"key"`
				FullName string `json:"fullName"`
			} `json:"l2Category"`
		} `json:"categories"`
		SearchQuery string `json:"searchQuery"`
		Attributes  struct {
		} `json:"attributes"`
		AttributesById  []interface{} `json:"attributesById"`
		AttributesByKey []interface{} `json:"attributesByKey"`
		AttributeRanges []interface{} `json:"attributeRanges"`
		AttributeLabels []interface{} `json:"attributeLabels"`
		SortOptions     struct {
			SortBy        string `json:"sortBy"`
			SortOrder     string `json:"sortOrder"`
			SortAttribute string `json:"sortAttribute"`
		} `json:"sortOptions"`
		Pagination struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		} `json:"pagination"`
		Distance struct {
			Postcode       string `json:"postcode"`
			DistanceMeters int    `json:"distanceMeters"`
		} `json:"distance"`
		ViewOptions struct {
			Kind string `json:"kind"`
		} `json:"viewOptions"`
		SearchInTitleAndDescription bool `json:"searchInTitleAndDescription"`
		BypassSpellingSuggestion    bool `json:"bypassSpellingSuggestion"`
	} `json:"searchRequest"`
	SearchCategory        int `json:"searchCategory"`
	SearchCategoryOptions []struct {
		FullName  string `json:"fullName"`
		Id        int    `json:"id"`
		Key       string `json:"key"`
		Name      string `json:"name"`
		ParentId  int    `json:"parentId,omitempty"`
		ParentKey string `json:"parentKey,omitempty"`
	} `json:"searchCategoryOptions"`
	SeoFriendlyAttributes []string `json:"seoFriendlyAttributes"`
	AttributeHierarchy    struct {
		OfferedSince []struct {
			AttributeValueId    interface{} `json:"attributeValueId"`
			AttributeValueLabel interface{} `json:"attributeValueLabel"`
			AttributeValueKey   string      `json:"attributeValueKey"`
			AttributeLabel      string      `json:"attributeLabel"`
			IsDefault           bool        `json:"isDefault"`
		} `json:"offeredSince"`
	} `json:"attributeHierarchy"`
	CategoriesById struct {
		Field1 struct {
			FullName string `json:"fullName"`
			Id       int    `json:"id"`
			Key      string `json:"key"`
			Name     string `json:"name"`
			ParentId int    `json:"parentId"`
		} `json:"280"`
		Field2 struct {
			FullName string `json:"fullName"`
			Id       int    `json:"id"`
			Key      string `json:"key"`
			Name     string `json:"name"`
		} `json:"1847"`
	} `json:"categoriesById"`
	MetaTags struct {
		MetaTitle       string `json:"metaTitle"`
		MetaDescription string `json:"metaDescription"`
		PageTitleH1     string `json:"pageTitleH1"`
	} `json:"metaTags"`
}
