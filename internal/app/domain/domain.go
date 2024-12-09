package domain

// ShortURLDomain структура с описанием доменной сущности ShortURL
type ShortURLDomain struct {
	UUID     string
	ShortURI string
	LongURL  string
	UserID   string
	Deleted  bool
}

// CreateShortURLBatchDomain структура с описанием доменной сущности CreateShortURLBatch
type CreateShortURLBatchDomain struct {
	CorrelationUUID string
	LongURL         string
}

// CreateShortURLBatchResultDomain структура с описанием доменной сущности CreateShortURLBatchResult
type CreateShortURLBatchResultDomain struct {
	CorrelationUUID string
	ShortURI        string
}
