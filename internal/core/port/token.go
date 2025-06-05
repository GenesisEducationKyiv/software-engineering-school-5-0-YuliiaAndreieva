package port

type TokenService interface {
	GenerateToken() (string, error)
}
