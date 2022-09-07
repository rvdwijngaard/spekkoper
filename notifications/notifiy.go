package notifications

import (
	"context"

	"encore.app/spekkoper"
	"encore.dev/pubsub"
	"encore.dev/rlog"
)

var _ = pubsub.NewSubscription(
	spekkoper.NewAds, "send-new-ad-notification",
	pubsub.SubscriptionConfig[*spekkoper.NewQueryResultEvent]{
		Handler: SendWelcomeEmail,
	},
)

func SendWelcomeEmail(ctx context.Context, event *spekkoper.NewQueryResultEvent) error {
	rlog.Info("got a new ad", "ad", event.Advertisement)
	return nil
}
