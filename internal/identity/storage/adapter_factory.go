package storage

type Adapter interface{}

type AdapterFactory interface {
	CreateStorageAdapter(name string) Adapter
}
