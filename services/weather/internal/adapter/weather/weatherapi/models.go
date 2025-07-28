package weatherapi

type response struct {
	TempC     float64 `json:"temp_c"`
	Humidity  int     `json:"humidity"`
	Condition struct {
		Text string `json:"text"`
	} `json:"condition"`
	WindKph float64 `json:"wind_kph"`
}

type currentEnvelope struct {
	Current response `json:"current"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type searchItem struct {
	Name string `json:"name"`
}

type searchEnvelope struct {
	Data []searchItem `json:"data"`
}
