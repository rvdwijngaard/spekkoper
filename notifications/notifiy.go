package notifications

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

func SendWelcomeEmail(ctx context.Context, event *spekkoper.NewQueryResultEvent) error {
	ad := event.Advertisement

	body := strings.NewReader(fmt.Sprintf("%s\nprijs: %d\n%s", ad.Description, ad.PriceInfo, ad.Location))

	req, _ := http.NewRequest("POST", "https://ntfy.sh/spekkoper", body)
	req.Header.Set("Click", event.Advertisement.URL)
	req.Header.Set("X-Title", event.Advertisement.Title)
	lo.ForEach(event.Advertisement.ImageUrls, func(url string, _ int) {
		req.Header.Set("Attach", "https:"+url)
	})
	req.Header.Set("Email", "ronvanderwijngaard@kliksafe.nl")
	_, err := http.DefaultClient.Do(req)
	return err
}
