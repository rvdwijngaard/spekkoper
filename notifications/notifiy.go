package notifications

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encore.app/spekkoper"
	"encore.dev/pubsub"
	"github.com/samber/lo"
)

var _ = pubsub.NewSubscription(
	spekkoper.NewAds, "send-new-ad-notification",
	pubsub.SubscriptionConfig[*spekkoper.NewQueryResultEvent]{
		Handler: SendWelcomeEmail,
	},
)

var secrets struct {
	ForwardEmailAddress string // ed25519 private key for SSH server
}

func SendWelcomeEmail(ctx context.Context, event *spekkoper.NewQueryResultEvent) error {
	ad := event.Advertisement

	body := strings.NewReader(fmt.Sprintf("%s\ndatum:%s\nprijs: €%d\n%s", ad.Date.Format(time.RFC3339), ad.Description, ad.PriceInfo.PriceCents/100, ad.Location.CityName))

	req, _ := http.NewRequest("POST", "https://ntfy.sh/spekkoper", body)
	req.Header.Set("Click", event.Advertisement.URL)
	req.Header.Set("X-Title", event.Advertisement.Title)
	lo.ForEach(event.Advertisement.ImageUrls, func(url string, _ int) {
		req.Header.Set("Attach", "https:"+url)
	})
	req.Header.Set("Email", secrets.ForwardEmailAddress)
	_, err := http.DefaultClient.Do(req)
	return err
}
