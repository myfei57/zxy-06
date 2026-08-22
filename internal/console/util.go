package console

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func chiURL(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func storeNow() time.Time {
	return time.Now().UTC()
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}
