package marktplaats

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	rawURL := "https://www.marktplaats.nl/l/huis-en-inrichting/kachels/#q:zibro|f:31,32,4205|distanceMeters:50000|postcode:3461CC"

	res, err := ParseURL(context.TODO(), rawURL)
	assert.NoError(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "zibro", res.Query)
		assert.Equal(t, 504, res.Category)
		assert.Equal(t, 513, res.SubCategory)
		assert.Equal(t, "3461CC", res.PostCode)
		assert.Equal(t, 50000, res.DistanceMeters)
		assert.Equal(t, []int{31, 32, 4205}, res.AttributesByID)
	}

	t.Run("pass in an invalid URL", func(t *testing.T) {
		_, err := ParseURL(context.TODO(), "invalid uri")
		assert.Error(t, err)
	})
}
