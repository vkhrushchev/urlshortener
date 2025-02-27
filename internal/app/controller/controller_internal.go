package controller

import (
	"context"
	"encoding/json"
	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"net/http"
)

type statsProvider interface {
	GetStats(ctx context.Context) (urlCount int, userCount int, err error)
}

// InternalController используется для обработки запросов приложения из доверенной сети
type InternalController struct {
	statsProvider statsProvider
}

// NewInternalController создает новый экземпляр структуры InternalController
func NewInternalController(statsProvider statsProvider) *InternalController {
	return &InternalController{statsProvider: statsProvider}
}

// GetStats возвращает статистику по сервису
//
//	@Summary	статистика по сервису
//	@Accepts	plain
//	@Produce	json
//	@Success	200
//	@Failure	500	{string}	string	"внутренняя ошибка сервиса"
//	@Router		/api/internal/stats [get]
func (c *InternalController) GetStats(w http.ResponseWriter, r *http.Request) {
	urlCount, userCount, err := c.statsProvider.GetStats(r.Context())
	if err != nil {
		log.Errorw("controller: failed to get stats", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiResponse := dto.APIInternalGetStatsResponse{
		URLCount:  urlCount,
		UserCount: userCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiResponse); err != nil {
		log.Errorw("controller: failed to encode response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
