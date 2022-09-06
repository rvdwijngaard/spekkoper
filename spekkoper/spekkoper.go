package spekkoper

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"encore.app/marktplaats"
	"encore.dev/beta/errs"
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

//encore:api public path=/register method=POST
func RegisterNewQuery(ctx context.Context, p Query) (*Query, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	} else if err := insert(ctx, id, p.Query, p.Category, p.SubCategory, p.PostCode, p.DistanceMeters); err != nil {
		return nil, err
	}
	p.ID = id
	return &p, nil
}

// Get retrieves the query configuration for the id.
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
func Check(ctx context.Context, id string) (*QueryResponse, error) {
	q, err := get(ctx, id)
	if err != nil {
		return nil, err
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
		return nil, err
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

	stored, err := getResultIDsFromDB(ctx, id)
	if err != nil {
		return nil, err
	}

	newAds := lo.Filter(bar, func(advertisement Advertisement, _ int) bool {
		return !lo.Contains(stored, advertisement.ID)
	})

	tx, err := sqldb.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// get results of the current stuff
	lo.ForEach(newAds, func(advertisement Advertisement, _ int) {
		err = storeResult(ctx, id, advertisement)
	})
	if err != nil {
		if err := tx.Rollback(); err != nil {
			//rlog.Error(err)
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, errs.Wrap(err, "could not write query results to db")
	}

	foo := &QueryResponse{
		Advertisements: newAds,
	}
	return foo, nil
}

func storeResult(ctx context.Context, queryID string, advertisement Advertisement) error {
	_, err := sqldb.Exec(ctx, `
        INSERT INTO query_result (query_id, result_id)
        VALUES ($1, $2)
    `, queryID, advertisement.ID)

	return err
}

func getResultIDsFromDB(ctx context.Context, queryID string) ([]string, error) {
	query := `
		SELECT result_id
        FROM query_result 
        WHERE id = $1
	`
	rows, err := sqldb.Query(ctx, query, queryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string

		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
