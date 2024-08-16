package controller

import (
	"net/http"

	"github.com/vkhrushchev/urlshortener/internal/app/db"
)

type HealthController struct {
	dbLookup *db.DBLookup
}

func NewHealthController(dbLookup *db.DBLookup) *HealthController {
	return &HealthController{
		dbLookup: dbLookup,
	}
}

func (c *HealthController) Ping(w http.ResponseWriter, r *http.Request) {
	isDBConntectionAlive := c.dbLookup.Ping(r.Context())
	if !isDBConntectionAlive {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
