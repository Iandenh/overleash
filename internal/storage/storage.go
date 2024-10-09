package storage

type Store interface {
	Read(key string) ([]byte, error)
	Write(key string, data []byte) (writeErr error)
}
