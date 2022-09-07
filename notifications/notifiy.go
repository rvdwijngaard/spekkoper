package notifications

import (
	"context"
	"net/http"
	"strings"

	"encore.app/spekkoper"
	"encore.dev/pubsub"
)

var _ = pubsub.NewSubscription(
	spekkoper.NewAds, "send-new-ad-notification",
	pubsub.SubscriptionConfig[*spekkoper.NewQueryResultEvent]{
		Handler: SendWelcomeEmail,
	},
)

func SendWelcomeEmail(ctx context.Context, event *spekkoper.NewQueryResultEvent) error {
	req, _ := http.NewRequest("POST", "https://ntfy.sh/spekkoper",
		strings.NewReader(`New advertisement. 🐶`))
	req.Header.Set("Click", event.Advertisement.URL)
	//req.Header.Set("Attach", "https://nest.com/view/yAxkasd.jpg")
	//req.Header.Set("Actions", "http, Open door, https://api.nest.com/open/yAxkasd, clear=true")
	req.Header.Set("Email", "ronvanderwijngaard@kliksafe.nl")
	_, err := http.DefaultClient.Do(req)
	return err
}
