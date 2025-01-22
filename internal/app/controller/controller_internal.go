package controller

import (
	"encoding/json"
	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/app/usecase"
	"net/http"
)

// InternalController используется для обработки запросов приложения из доверенной сети
type InternalController struct {
	statsUseCase usecase.IStatsUseCase
}

// NewInternalController создает новый экземпляр структуры InternalController
func NewInternalController(statsUseCase usecase.IStatsUseCase) *InternalController {
	return &InternalController{statsUseCase: statsUseCase}
}

func (c *InternalController) GetStats(w http.ResponseWriter, r *http.Request) {
	urlCount, userCount, err := c.statsUseCase.GetStats(r.Context())
	if err != nil {
		log.Errorw("controller: failed to get stats", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiResponse := dto.APIInternalGetStatsResponse{
		UrlCount:  urlCount,
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
