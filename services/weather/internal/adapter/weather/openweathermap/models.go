package openweathermap

type Response struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Cod     interface{} `json:"cod"`
	Message string      `json:"message"`
}
