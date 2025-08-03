package domain

type Frequency string

const (
	Daily   Frequency = "daily"
	Weekly  Frequency = "weekly"
	Monthly Frequency = "monthly"
)

type Weather struct {
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
	WindSpeed   float64 `json:"wind_speed"`
}

type Subscription struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	City      string    `json:"city"`
	Frequency Frequency `json:"frequency"`
	Confirmed bool      `json:"confirmed"`
	Token     string    `json:"token"`
}

type ListSubscriptionsQuery struct {
	Frequency Frequency `json:"frequency"`
	LastID    int       `json:"last_id"`
	PageSize  int       `json:"page_size"`
}

type SubscriptionList struct {
	Subscriptions []Subscription `json:"subscriptions"`
	LastIndex     int            `json:"last_index"`
}

type WeatherMailSuccessInfo struct {
	Email   string  `json:"email"`
	City    string  `json:"city"`
	Weather Weather `json:"weather"`
	Token   string  `json:"token"`
}

type WeatherMailErrorInfo struct {
	Email string `json:"email"`
	City  string `json:"city"`
}

type BroadcastRequest struct {
	Frequency Frequency `json:"frequency" binding:"required"`
}

type WeatherRequest struct {
	City string `json:"city"`
}

type WeatherResponse struct {
	Weather Weather `json:"weather"`
}

type WeatherErrorEmailRequest struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type WeatherUpdateEmailRequest struct {
	To               string `json:"to"`
	Subject          string `json:"subject"`
	Name             string `json:"name"`
	City             string `json:"city"`
	Description      string `json:"description"`
	Temperature      int    `json:"temperature"`
	Humidity         int    `json:"humidity"`
	WindSpeed        int    `json:"windSpeed"`
	UnsubscribeToken string `json:"unsubscribeToken"`
}
