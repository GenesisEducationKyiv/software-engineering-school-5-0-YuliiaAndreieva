package weather

type Response struct {
	TempC     float64 `json:"temp_c"`
	Humidity  int     `json:"humidity"`
	Condition struct {
		Text string `json:"text"`
	} `json:"condition"`
}

type currentEnvelope struct {
	Current Response  `json:"current"`
	Error   *APIError `json:"error,omitempty"`
}

type SearchItem struct {
	Name string `json:"name"`
}
