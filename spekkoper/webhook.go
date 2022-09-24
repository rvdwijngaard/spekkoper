package spekkoper

import (
	"encoding/json"
	"io"
	"net/http"

	"encore.dev/rlog"
)

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
	rlog.Info("got payload", "payload", string(b))

	var dto struct {
		Data struct {
			UserName string `json:"username"`
			Email    string `json:"email"`
		}
	}
	err = json.Unmarshal(b, &dto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var u = User{
		UserName: dto.Data.UserName,
		Email:    dto.Data.Email,
	}

	rlog.Info("new user registered", "user", u)

	w.WriteHeader(http.StatusOK)
}
