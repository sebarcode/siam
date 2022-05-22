package siam

type Storage interface {
	Load(pool *SessionPool) error
	Store(pool *SessionPool) error
	Get(id string) (*Session, error)
	Remove(id string)
	Write(sess *Session) error
	Close()
}
