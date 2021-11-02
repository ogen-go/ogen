package uri

type HeaderEncoder struct {
	*scraper
	explode bool
}

type HeaderEncoderConfig struct {
	Explode bool
}

func NewHeaderEncoder(cfg HeaderEncoderConfig) *HeaderEncoder {
	return &HeaderEncoder{
		scraper: newScraper(),
		explode: cfg.Explode,
	}
}
