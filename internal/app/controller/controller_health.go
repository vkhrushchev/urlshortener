package controller

import (
	"net/http"

	"github.com/vkhrushchev/urlshortener/internal/app/db"
)

// HealthController обрабатывает запросы helhtcheck-а
type HealthController struct {
	dbLookup *db.DBLookup // dbLookup - обертка для доступа к БД
}

// NewHealthController создает новый экземпляр структуры HealthController
//
//	dbLookup - обертка для доступа к БД
func NewHealthController(dbLookup *db.DBLookup) *HealthController {
	return &HealthController{
		dbLookup: dbLookup,
	}
}

// Ping проверяет что сервер "жив"
//
//	@Summary	проверка работоспособности сервера
//	@Accepts	plain
//	@Produce	plain
//	@Success	200
//	@Failure	500	{string}	string	"внутренняя ошибка сервиса"
//	@Router		/ping [get]
func (c *HealthController) Ping(w http.ResponseWriter, r *http.Request) {
	isDBConnectionAlive := c.dbLookup.Ping(r.Context())
	if !isDBConnectionAlive {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
