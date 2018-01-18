package storage


import "os"
import "fmt"
import "path/filepath"
import "encoding/binary"

import "github.com/romana/rlog"
import bolt "github.com/coreos/bbolt"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"

type BoltStore struct {
	DB *bolt.DB
}

var Bolt_backend *BoltStore = &BoltStore{} // global variable
var logger *log.Entry

func (b *BoltStore) Init(params map[string]interface{}) (err error) {
	logger = globals.Logger.WithFields(log.Fields{"com": "STORE"})
	current_path := filepath.Join(os.TempDir(), "derod_database.db")
	logger.Infof("Initializing boltdb store at path %s", current_path)

	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	b.DB, err = bolt.Open(current_path, 0600, nil)
	if err != nil {
		logger.Fatalf("Cannot open boltdb store err %s\n", err)
	}

	return nil
}

func (b *BoltStore) Shutdown() (err error) {
	logger.Infof("Shutting boltdb store")
	if b.DB != nil {
		b.DB.Close()
	}

	return nil
}

func (b *BoltStore) StoreObject(universe_name []byte, galaxy_name []byte, solar_name []byte, key []byte, data []byte) (err error) {
	rlog.Tracef(10, "Storing object %s %s %x  data len %d\n", string(universe_name), string(galaxy_name), key, len(data))
	// open universe bucket
	err = b.DB.Update(func(tx *bolt.Tx) error {
		universe, err := tx.CreateBucketIfNotExists(universe_name)
		if err != nil {
			logger.Errorf("Error while creating universe bucket %s\n", err)
			return err
		}
		galaxy, err := universe.CreateBucketIfNotExists(galaxy_name)
		if err != nil {
			logger.Errorf("Error while creating galaxy bucket %s err\n", string(galaxy_name), err)
			return err
		}

		solar, err := galaxy.CreateBucketIfNotExists(solar_name)
		if err != nil {
			logger.Errorf("Error while creating solar bucket %s err\n", string(solar_name), err)
			return err
		}
		// now lets update the object attribute
		err = solar.Put(key, data)

		return err
	})
	return

}

func (b *BoltStore) LoadObject(universe []byte, bucket_name []byte, solar_bucket []byte, key []byte) (data []byte, err error) {
	rlog.Tracef(10, "Loading object %s %s %x\n", string(universe), string(bucket_name), key)
	// open universe bucket
	err = b.DB.View(func(tx *bolt.Tx) error {
		universe := tx.Bucket(universe)
		if universe == nil {
			return fmt.Errorf("No Such Universe %x\n", universe)
		}
		bucket := universe.Bucket(bucket_name)
		if bucket == nil {
			return fmt.Errorf("No Such Bucket %x\n", bucket_name)
		}

		solar := bucket.Bucket(solar_bucket)
		if solar == nil {
			return fmt.Errorf("No Such Bucket %x\n", solar)
		}

		// now lets find the object
		value := solar.Get(key)

		data = make([]byte, len(value), len(value))

		copy(data, value) // job done

		return nil
	})
	return
}

// this function stores a uint64
func (b *BoltStore) StoreUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data uint64) error {
	return b.StoreObject(universe_bucket, galaxy_bucket, solar_bucket, key, itob(data))

}

//  this function loads the data as 64 byte integer
func (b *BoltStore) LoadUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) (uint64, error) {
	object_data, err := b.LoadObject(universe_bucket, galaxy_bucket, solar_bucket, key)
	if err != nil {
		return 0, err
	}

	if len(object_data) == 0 {
		return 0, fmt.Errorf("No value stored here, we should look more")
	}

	if len(object_data) != 8 {
		panic("Database corruption, invalid data ")
	}

	value := binary.BigEndian.Uint64(object_data)
	return value, nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
