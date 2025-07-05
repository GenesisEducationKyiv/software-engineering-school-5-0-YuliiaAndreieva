package ports

type TokenService interface {
	GenerateToken() (string, error)
}
