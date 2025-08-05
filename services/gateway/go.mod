module gateway

go 1.24

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/go-playground/validator/v10 v10.20.0
	github.com/joho/godotenv v1.5.1
	github.com/prometheus/client_golang v1.19.0
	github.com/sirupsen/logrus v1.9.3
	shared v0.0.0-00010101000000-000000000000
)

replace shared => ../../shared
