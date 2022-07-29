package storagedriver

type StorageDriver interface {
	GetLocalStorageLocation() string
	UploadTar(string) error
}

type localStorageDriver struct {
	localStorageLocation string
}

func (l *localStorageDriver) GetLocalStorageLocation() string {
	return l.localStorageLocation
}

func (l *localStorageDriver) UploadTar(storageLocation string) error {
	// noop for local
	return nil
}

type StorageDriverParams struct {
	LocalStorageLocation string
}

func NewStorageDriver(params StorageDriverParams) (StorageDriver, error) {
	return &localStorageDriver{params.LocalStorageLocation}, nil
}
