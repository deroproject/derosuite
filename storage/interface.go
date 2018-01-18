package storage

type Store interface {
	Init(param map[string]interface{}) error                                                                      // init the backend and connect to it
	Shutdown() error                                                                                              // shutdown the backend and connect to it
	StoreObject(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data []byte) error // store object to a specific universe
	LoadObject(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) ([]byte, error)     // load object from a specific universe

	StoreUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data uint64) error // store
	LoadUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) (uint64, error)     // load object
}

//var  Store Backend_Store// the system shouls chooose a backend at start Time
