package main

import "time"

type hotspot struct {
	Lng            float64   `json:"lng"`
	Lat            float64   `json:"lat"`
	TimestampAdded time.Time `json:"timestamp_added"`
	Status         struct {
		Online      string   `json:"online"`
		ListenAddrs []string `json:"listen_addrs"`
		Height      int      `json:"height"`
	} `json:"status"`
	RewardScale      float64 `json:"reward_scale"`
	Owner            string  `json:"owner"`
	Nonce            int     `json:"nonce"`
	Name             string  `json:"name"`
	Location         string  `json:"location"`
	LastPocChallenge int     `json:"last_poc_challenge"`
	LastChangeBlock  int     `json:"last_change_block"`
	Geocode          struct {
		ShortStreet  string `json:"short_street"`
		ShortState   string `json:"short_state"`
		ShortCountry string `json:"short_country"`
		ShortCity    string `json:"short_city"`
		LongStreet   string `json:"long_street"`
		LongState    string `json:"long_state"`
		LongCountry  string `json:"long_country"`
		LongCity     string `json:"long_city"`
		CityID       string `json:"city_id"`
	} `json:"geocode"`
	BlockAdded int    `json:"block_added"`
	Block      int    `json:"block"`
	Address    string `json:"address"`
}

type reward struct {
	Total     float64   `json:"total"`
	Timestamp time.Time `json:"timestamp"`
	Sum       int       `json:"sum"`
	Stddev    float64   `json:"stddev"`
	Min       float64   `json:"min"`
	Median    float64   `json:"median"`
	Max       float64   `json:"max"`
	Avg       float64   `json:"avg"`
}

type price struct {
	Timestamp time.Time `json:"timestamp"`
	Price     int       `json:"price"`
	Block     int       `json:"block"`
}

type hotspotsResponse struct {
	Data []hotspot `json:"data"`
}

type rewardsResponse struct {
	Meta struct {
		MinTime time.Time `json:"min_time"`
		MaxTime time.Time `json:"max_time"`
		Bucket  string    `json:"bucket"`
	} `json:"meta"`
	Data []reward `json:"data"`
}

type priceResponse struct {
	Data price `json:"data"`
}
