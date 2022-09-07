package spekkoper

import (
	"context"
	"testing"

	"encore.app/marktplaats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterNewQuery(t *testing.T) {
	q := Query{
		Query:          "bikes",
		Category:       10,
		SubCategory:    20,
		PostCode:       "0000XX",
		DistanceMeters: 99,
	}
	ctx := context.TODO()
	res, err := Post(ctx, q)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.ID)
	t.Run("get the query by its id", func(t *testing.T) {
		res2, err := Get(ctx, res.ID)
		assert.NoError(t, err)
		assert.Equal(t, res, res2)
	})
	t.Run("get a non existent query", func(t *testing.T) {
		_, err := Get(ctx, "foo")
		assert.Error(t, err)
	})
}

type mpMock struct {
	mock.Mock
}

func (m *mpMock) Query(ctx context.Context, request marktplaats.QueryRequest) (*marktplaats.QueryResponse, error) {
	args := m.Called(ctx, request)
	if v, ok := args[0].(*marktplaats.QueryResponse); ok {
		return v, args.Error(1)
	}

	return nil, args.Error(1)
}

func TestRun(t *testing.T) {
	q := &Query{
		Query:          "bikes",
		Category:       10,
		SubCategory:    20,
		PostCode:       "0000XX",
		DistanceMeters: 99,
	}
	ctx := context.TODO()
	var err error
	q, err = Post(ctx, *q)
	if err != nil {
		t.Fatal(err)
	}
	if q == nil || q.ID == "" {
		t.Fatal()
	}
	id := q.ID
	m := &mpMock{}
	m.On("Query", mock.Anything, marktplaats.QueryRequest{
		Query:              q.Query,
		PostCode:           q.PostCode,
		DistanceMeters:     q.DistanceMeters,
		Limit:              0,
		Offset:             0,
		IncludeCommercials: false,
		Category:           q.Category,
		SubCategory:        q.SubCategory,
	}).Return(&marktplaats.QueryResponse{
		Advertisements: []marktplaats.Advertisement{
			{
				ID:    "foo",
				Title: "bar",
				Location: marktplaats.Location{
					CityName: "Ede",
				},
				PriceInfo: marktplaats.PriceInfo{
					PriceCents: 10,
				},
				URL: "https://foo.bar.com",
			},
		},
	}, nil).Once()

	s := &Service{
		marktplaats: m,
	}

	res, err := s.Run(ctx, id)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Advertisements, 1)

	t.Run("get new items", func(t *testing.T) {
		m.On("Query", mock.Anything, marktplaats.QueryRequest{
			Query:              q.Query,
			PostCode:           q.PostCode,
			DistanceMeters:     q.DistanceMeters,
			Limit:              0,
			Offset:             0,
			IncludeCommercials: false,
			Category:           q.Category,
			SubCategory:        q.SubCategory,
		}).Return(&marktplaats.QueryResponse{
			Advertisements: []marktplaats.Advertisement{
				{
					ID:    "bar",
					Title: "baz",
					Location: marktplaats.Location{
						CityName: "Rotterdam",
					},
					PriceInfo: marktplaats.PriceInfo{
						PriceCents: 100,
					},
					URL: "https://foo.bar.com",
				},
			},
		}, nil).Once()
		res, err := s.Run(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Advertisements, 1)
		assert.Equal(t, "bar", res.Advertisements[0].ID)
	})
}
