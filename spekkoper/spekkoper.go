package spekkoper

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"encore.app/marktplaats"
	"encore.dev/beta/errs"
	"encore.dev/cron"
	"encore.dev/pubsub"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/samber/lo"
)

type marktplaatsClient interface {
	Query(ctx context.Context, request marktplaats.QueryRequest) (*marktplaats.QueryResponse, error)
}

type mclient struct{}

func (m *mclient) Query(ctx context.Context, request marktplaats.QueryRequest) (*marktplaats.QueryResponse, error) {
	return marktplaats.Query(ctx, request)
}

// encore:service
type Service struct {
	marktplaats marktplaatsClient
}

func initService() (*Service, error) {
	return &Service{
		marktplaats: &mclient{},
	}, nil
}

type NewQueryResultEvent struct{ Advertisement marktplaats.Advertisement }

var NewAds = pubsub.NewTopic[*NewQueryResultEvent]("new-advertisements", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// Periodically check all registered queries for new advertisements
var _ = cron.NewJob("run-all-registered-queries", cron.JobConfig{
	Title:    "Run all registered queries",
	Every:    1 * cron.Minute,
	Endpoint: CheckAll,
})

type User struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
}

// Webhook receives incoming webhooks from aut0
//
//encore:api public raw
func Webhook(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var u User
	err = json.Unmarshal(b, &u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rlog.Info("new user registered", "user", u)

	w.WriteHeader(http.StatusOK)
}

//encore:api private
func CheckAll(ctx context.Context) error {
	srv :=
		&Service{
			marktplaats: &mclient{},
		}
	queries, err := getAllRegisteredQueries(ctx)
	if err != nil {
		return errs.Wrap(err, "failed to list all registered queries")
	}
	for _, u := range queries {
		if _, err := srv.Run(ctx, u.ID); err != nil {
			rlog.Error("could not run query", "err", err)
		}
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
	AttributesByID []int
}

type PostQueryRequest struct {
	// QueryURL is the markplaats query URL copied from the browser,
	// for example: https://www.marktplaats.nl/l/huis-en-inrichting/kachels/#q:zibro|f:31,32,4205|distanceMeters:50000|postcode:3901EF
	QueryURL string
	Query    Query
}

func parseQueryFromURL(ctx context.Context, queryURL string) (Query, error) {
	res, err := marktplaats.ParseURL(ctx, queryURL)
	if err != nil {
		return Query{}, err
	}
	return Query{
		Query:          res.Query,
		Category:       res.Category,
		SubCategory:    res.SubCategory,
		PostCode:       res.PostCode,
		DistanceMeters: res.DistanceMeters,
		AttributesByID: res.AttributesByID,
	}, nil
}

// Post creates a new query
//
//encore:api public path=/query method=POST
func Post(ctx context.Context, r PostQueryRequest) (*Query, error) {
	q := r.Query
	if r.QueryURL != "" {
		var err error
		q, err = parseQueryFromURL(ctx, r.QueryURL)
		if err != nil {
			return nil, err
		}
	}
	id, err := generateID()
	if err != nil {
		return nil, err
	}
	q.ID = id
	if err != nil {
		return nil, err
	} else if err := insert(ctx, q.ID, q.Query, q.Category, q.SubCategory, q.PostCode, q.DistanceMeters, q.AttributesByID); err != nil {
		return nil, err
	}
	return &q, nil
}

// Get retrieves the query configuration for the id.
//
//encore:api public method=GET path=/query/:id
func Get(ctx context.Context, id string) (*Query, error) {
	return get(ctx, id)
}

// Delete deletes the query configuration for the id.
//
//encore:api public method=DELETE path=/query/:id
func Delete(ctx context.Context, id string) error {
	_, err := sqldb.Exec(ctx, "DELETE FROM query_result WHERE query_id=$1", id)
	if err != nil {
		return err
	}
	_, err = sqldb.Exec(ctx, "DELETE FROM query WHERE id=$1", id)
	if err != nil {
		return err
	}
	return nil
}

type ListResult struct {
	Queries []Query
}

// Lists all registered queris
//
//encore:api public method=GET path=/query
func List(ctx context.Context) (*ListResult, error) {
	query := `
		SELECT id, query, category,sub_category,postcode, 	distance_meters,attributes_by_id
        FROM query         
	`
	rows, err := sqldb.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []Query
	for rows.Next() {
		var q Query
		err := rows.Scan(&q.ID, &q.Query, &q.Category, &q.SubCategory, &q.PostCode, &q.DistanceMeters, &q.AttributesByID)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}
	return &ListResult{queries}, nil
}

func get(ctx context.Context, id string) (*Query, error) {
	u := &Query{
		ID: id,
	}
	err := sqldb.QueryRow(ctx, `
        SELECT query, category, sub_category, postcode, distance_meters, attributes_by_id FROM query
        WHERE id = $1
    `, id).Scan(&u.Query, &u.Category, &u.SubCategory, &u.PostCode, &u.DistanceMeters, &u.AttributesByID)

	return u, err
}

// insert a query into the database.
func insert(ctx context.Context, id, query string, category, subCategory int, postcode string, distance int, attrs []int) error {
	_, err := sqldb.Exec(ctx, `
        INSERT INTO query (id, query, category, sub_category, postcode, distance_meters, attributes_by_id )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, id, query, category, subCategory, postcode, distance, attrs)

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
	Advertisements []marktplaats.Advertisement `json:"advertisements"`
}

type RunParams struct {
	Limit              int
	Offset             int
	IncludeCommercials bool
}

// Run executes a stored query
//
//encore:api public path=/query/:id/run
func (srv *Service) Run(ctx context.Context, id string) (*QueryResponse, error) {
	q, err := get(ctx, id)
	if err != nil {
		return nil, err
	}

	res, err := srv.marktplaats.Query(ctx, marktplaats.QueryRequest{
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

	stored, err := getResultIDsFromDB(ctx, id)
	if err != nil {
		return nil, err
	}

	newAds := lo.Filter(res.Advertisements, func(advertisement marktplaats.Advertisement, _ int) bool {
		return !lo.Contains(stored, advertisement.ID)
	})

	tx, err := sqldb.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// get results of the current stuff
	lo.ForEach(newAds, func(advertisement marktplaats.Advertisement, _ int) {
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

	lo.ForEach(newAds, func(ad marktplaats.Advertisement, _ int) {
		if _, err := NewAds.Publish(ctx, &NewQueryResultEvent{Advertisement: ad}); err != nil {
			rlog.Error("could not publish new ad", "err", err)
		}
	})
	foo := &QueryResponse{
		Advertisements: newAds,
	}
	return foo, nil
}

func storeResult(ctx context.Context, queryID string, ad marktplaats.Advertisement) error {
	_, err := sqldb.Exec(ctx, `
        INSERT INTO query_result (query_id, result_id, title, city, url, price_in_cents, image_urls)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, queryID, ad.ID, ad.Title, ad.Location.CityName, ad.URL, ad.PriceInfo.PriceCents, ad.ImageUrls)

	return err
}

func getResultIDsFromDB(ctx context.Context, queryID string) ([]string, error) {
	query := `
		SELECT result_id
        FROM query_result 
        WHERE query_id = $1
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
