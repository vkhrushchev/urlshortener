package entity

// ShortURLEntity структура с описанием сущности ShortURL для хранения в репозитории
type ShortURLEntity struct {
	UUID     string `json:"uuid"`
	ShortURI string `json:"short_url"`
	LongURL  string `json:"original_url"`
	UserID   string `json:"user_id"`
	Deleted  bool   `json:"is_deleted"`
}
