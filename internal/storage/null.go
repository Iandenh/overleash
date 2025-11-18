package storage

type NullStore struct{}

func (f *NullStore) Read(filename string) ([]byte, error) {
	return []byte("{}"), nil
}

func (f *NullStore) Write(filename string, data []byte) (writeErr error) {
	return nil
}
