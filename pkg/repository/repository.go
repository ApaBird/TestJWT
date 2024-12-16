package repository

type Storage interface {
	SaveToken(token string, guid string) error
	GetToken(guid string) (string, error)
	Close() error
}
