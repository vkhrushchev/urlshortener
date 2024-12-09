package domain

type ShortURLDomain struct {
	UUID     string
	ShortURI string
	LongURL  string
	UserID   string
	Deleted  bool
}

type CreateShortURLBatchDomain struct {
	CorrelationUUID string
	LongURL         string
}

type CreateShortURLBatchResultDomain struct {
	CorrelationUUID string
	ShortURI        string
}
