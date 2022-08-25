package spekkoper

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"encore.app/marktplaats"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/samber/lo"
)

//
//type NewQueryResultEvent struct{ UserID string }
//
//var Results = pubsub.NewTopic[*NewQueryResultEvent]("results", pubsub.TopicConfig{
//	DeliveryGuarantee: pubsub.AtLeastOnce,
//})

//
//// Send a welcome email to everyone who signed up in the last two hours.
//var _ = cron.NewJob("welcome-email", cron.JobConfig{
//	Title:    "Send welcome emails",
//	Every:    1 * cron.Minute,
//	Endpoint: CheckAll,
//})

// CheckAll checks all registered queries for changes
//
//encore:api public
func CheckAll(ctx context.Context) error {
	queries, err := getAllRegisteredQueries(ctx)
	if err != nil {
		return errs.Wrap(err, "failed to list all registered queries")
	}
	for _, u := range queries {
		marktplaats.Query(ctx, marktplaats.QueryRequest{
			Query:              u.Query,
			PostCode:           u.PostCode,
			DistanceMeters:     u.DistanceMeters,
			Limit:              0,
			Offset:             0,
			IncludeCommercials: false,
			Category:           u.Category,
			SubCategory:        u.SubCategory,
		})
		//Results.Publish(ctx, &NewQueryResultEvent{UserID: u.ID})
	}

	return nil
}

func getAllRegisteredQueries(ctx context.Context) ([]Query, error) {
	rows, err := sqldb.Query(ctx, "SELECT id, query, category, sub_category, postcode, distance_meters FROM query")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var queries []Query

	for rows.Next() {
		u := Query{}
		err = rows.Scan(&u.ID, &u.Query, &u.Category, &u.SubCategory, &u.PostCode, &u.DistanceMeters)
		if err != nil {
			return nil, err
		}
		queries = append(queries, u)
	}
	return queries, nil
}

type Query struct {
	ID             string
	Query          string
	Category       int
	SubCategory    int
	PostCode       string
	DistanceMeters int
}

type RegisterNewQueryResponse struct {
	ID       string
	Location string `header:"Location"`
}

//encore:api public path=/register method=POST
func RegisterNewQuery(ctx context.Context, p Query) (*RegisterNewQueryResponse, error) {
	rlog.Debug(p.Query)
	rlog.Debug("fss")
	id, err := generateID()
	if err != nil {
		return nil, err
	} else if err := insert(ctx, id, p.Query, p.Category, p.SubCategory, p.PostCode, p.DistanceMeters); err != nil {
		return nil, err
	}
	return &RegisterNewQueryResponse{ID: id, Location: "/check/" + id}, nil
}

// Get retrieves the original URL for the id.
//
//encore:api public method=GET path=/query/:id
func Get(ctx context.Context, id string) (*Query, error) {
	return get(ctx, id)
}

func get(ctx context.Context, id string) (*Query, error) {
	u := &Query{}
	err := sqldb.QueryRow(ctx, `
        SELECT query, category, sub_category, postcode, distance_meters FROM query
        WHERE id = $1
    `, id).Scan(&u.Query, &u.Category, &u.SubCategory, &u.PostCode, &u.DistanceMeters)

	return u, err
}

// insert a query into the database.
func insert(ctx context.Context, id, query string, category, subCategory int, postcode string, distance int) error {
	_, err := sqldb.Exec(ctx, `
        INSERT INTO query (id, query, category, sub_category, postcode, distance_meters )
        VALUES ($1, $2, $3, $4, $5, $6)
    `, id, query, category, subCategory, postcode, distance)

	return err
}

// generateID generates a random short ID.
func generateID() (string, error) {
	var data [6]byte // 6 bytes of entropy
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data[:]), nil
}

type QueryResponse struct {
	Advertisements []Advertisement `json:"advertisements"`
}

type Location struct {
	CityName string `json:"city_name"`
}

type Advertisement struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Location  Location  `json:"location"`
	PriceInfo PriceInfo `json:"price_info"`
	URL       string    `json:"url"`
}

type PriceInfo struct {
	PriceCents int `json:"price_cents"`
}

//encore:api path=/check/:id
func Check(ctx context.Context, id string) (QueryResponse, error) {
	q, err := get(ctx, id)
	if err != nil {
		return QueryResponse{}, err
	}

	res, err := marktplaats.Query(ctx, marktplaats.QueryRequest{
		Query:              q.Query,
		PostCode:           q.PostCode,
		DistanceMeters:     q.DistanceMeters,
		Limit:              0,
		Offset:             0,
		IncludeCommercials: false,
		Category:           q.Category,
		SubCategory:        q.SubCategory,
	})
	if err != nil {
		return QueryResponse{}, err
	}

	bar := lo.Map(res.Advertisements, func(v marktplaats.Advertisement, _ int) Advertisement {
		return Advertisement{
			ID:        v.ID,
			Title:     v.Title,
			Location:  Location{},
			PriceInfo: PriceInfo{},
			URL:       v.URL,
		}
	})
	foo := QueryResponse{
		Advertisements: bar,
	}
	return foo, nil
}
