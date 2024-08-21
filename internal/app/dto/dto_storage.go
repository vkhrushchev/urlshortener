package dto

type StorageShortURLEntry struct {
	UUID     string `json:"uuid"`
	ShortURI string `json:"short_url"`
	LongURL  string `json:"original_url"`
	UserID   string `json:"user_id"`
	Deleted  bool   `json:"is_deleted"`
}
