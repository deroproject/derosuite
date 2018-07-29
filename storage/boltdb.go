// Copyright 2017-2018 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// 64 bit arch will use this DB

package storage

import "os"
import "fmt"
import "sync"
import "strconv" // has intsize which give whether int is 64 bits or 32 bits
import "runtime"
import "path/filepath"
import "encoding/binary"

import "github.com/romana/rlog"
import bolt "github.com/coreos/bbolt"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"

type BoltStore struct {
	DB         *bolt.DB
	tx         *bolt.Tx
	sync.Mutex // lock this struct
	rw sync.RWMutex
}

// this object is returned
type BoltTXWrapper struct {
	bdb *BoltStore
	tx  *bolt.Tx
}

var Bolt_backend *BoltStore = &BoltStore{} // global variable
var logger *log.Entry

func (b *BoltStore) Init(params map[string]interface{}) (err error) {
	logger = globals.Logger.WithFields(log.Fields{"com": "STORE"})
	current_path := filepath.Join(globals.GetDataDirectory(), "derod_database.db")

	if params["--simulator"] == true {
		current_path = filepath.Join(os.TempDir(), "derod_simulation.db") // sp
	}
	logger.Infof("Initializing boltdb store at path %s", current_path)

	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.

	options := &bolt.Options{InitialMmapSize :1 * 1024 * 1024 * 1024}
	if runtime.GOOS != "windows" && strconv.IntSize == 64 {
		options.InitialMmapSize *= 40 // default allocation 40 GB
	}else{
		options.InitialMmapSize = 0 // on windows, make it 0	
	}

	b.DB, err = bolt.Open(current_path, 0600, options)
	if err != nil {
		logger.Fatalf("Cannot open boltdb store err %s", err)
	}

	// if simulation, delete the file , so as it gets cleaned up automcatically
	if params["--simulator"] == true {
		os.Remove(current_path)
	}

	// place db in no sync mode
	//b.DB.NoSync = true

	return nil
}

func (b *BoltStore) Shutdown() (err error) {
	logger.Infof("Shutting boltdb store")
	if b.DB != nil {
		b.DB.Sync() // sync the DB before closing
		b.DB.Close()
	}

	return nil
}

// get a new writable tx,
// we will manage the writable txs manually
// since a block may cause changes to a number of fields which must be reflected atomically
// this function is always triggered while the atomic lock is taken
// this is done avoid a race condition in returning the tx and using it
func (b *BoltStore) get_new_writable_tx() (tx *bolt.Tx) {
	if b.tx != nil {
		tx = b.tx // use existing pending tx
	} else { // create new pending tx

		tx, err := b.DB.Begin(true) // begin a new writable tx
		if err != nil {
			logger.Warnf("Error while creating new writable tx, err %s", err)
		} else {
			b.tx = tx
			rlog.Tracef(1, "Beginning new writable TX")

		}
	}
	return b.tx
}

// Commit the pending transaction to  disk
func (b *BoltStore) BeginTX(writable bool) (DBTX, error) {

	//  logger.Warnf(" new writable tx, err")
	txwrapper := &BoltTXWrapper{}
	tx, err := b.DB.Begin(writable) // begin a new writable tx
	if err != nil {
		logger.Warnf("Error while creating new writable tx, err %s", err)
		return nil, fmt.Errorf("Error while creating new writable tx, err %s", err)
	}
	txwrapper.tx = tx
	txwrapper.bdb = b // parent DB reference

	//logger.Warnf(" created new writable tx, err")
	return txwrapper, nil
}

func (b *BoltTXWrapper) Commit() error {
	err := b.tx.Commit()
	if err != nil {
		logger.Warnf("Error while committing tx, err %s", err)
		return err
	}
	//logger.Warnf(" Commiting TX")

	return nil
}

// Commit the pending transaction to  disk
func (b *BoltStore) Commit() {
	b.Lock()
	if b.tx != nil {
		rlog.Tracef(1, "Committing writable TX")
		err := b.tx.Commit()
		if err != nil {
			logger.Warnf("Error while commit tx, err %s", err)
		}
		b.tx = nil
	} else {
		logger.Warnf("Trying to Commit a NULL transaction, NOT possible")
	}
	b.Unlock()
}

// Roll back existing changes to  disk
func (b *BoltTXWrapper) Rollback() {
	// logger.Warnf(" Rollbacking  TX")
	b.tx.Rollback()
}

// Roll back existing changes to  disk
// TODO implement this
func (b *BoltTXWrapper) Sync() {

}

// Roll back existing changes to  disk
func (b *BoltStore) Rollback() {
	b.Lock()
	if b.tx != nil {
		rlog.Tracef(1, "Rollbacking writable TX")
		b.tx.Rollback()
		b.tx = nil
	} else {
		//logger.Warnf("Trying to Rollback a NULL transaction, NOT possible")
	}
	b.Unlock()
}

// sync the DB to disk
func (b *BoltStore) Sync() {
	b.Lock()
	if b.DB != nil {
		b.DB.Sync() // sync the DB
	}
	b.Unlock()
}

func (b *BoltStore) StoreObject(tx *bolt.Tx, universe_name []byte, galaxy_name []byte, solar_name []byte, key []byte, data []byte) (err error) {

	rlog.Tracef(10, "Storing object %s %s %x  data len %d", string(universe_name), string(galaxy_name), key, len(data))
	// open universe bucket

	universe, err := tx.CreateBucketIfNotExists(universe_name)
	if err != nil {
		logger.Errorf("Error while creating universe bucket %s", err)
		return err
	}
	galaxy, err := universe.CreateBucketIfNotExists(galaxy_name)
	if err != nil {
		logger.Errorf("Error while creating galaxy bucket %s err %s", string(galaxy_name), err)
		return err
	}

	solar, err := galaxy.CreateBucketIfNotExists(solar_name)
	if err != nil {
		logger.Errorf("Error while creating solar bucket %s err %s", string(solar_name), err)
		return err
	}
	// now lets update the object attribute
	err = solar.Put(key, data)

	return err

}

func (b *BoltTXWrapper) StoreObject(universe_name []byte, galaxy_name []byte, solar_name []byte, key []byte, data []byte) (err error) {
	return b.bdb.StoreObject(b.tx, universe_name, galaxy_name, solar_name, key, data)

}

// creates an empty bucket
func (b *BoltStore) CreateBucket(tx *bolt.Tx, universe_name []byte, galaxy_name []byte, solar_name []byte) (err error) {

	//rlog.Tracef(10, "Storing object %s %s %x  data len %d", string(universe_name), string(galaxy_name), key, len(data))

	universe, err := tx.CreateBucketIfNotExists(universe_name)
	if err != nil {
		logger.Errorf("Error while creating universe bucket %s", err)
		return err
	}
	galaxy, err := universe.CreateBucketIfNotExists(galaxy_name)
	if err != nil {
		logger.Errorf("Error while creating galaxy bucket %s err %s", string(galaxy_name), err)
		return err
	}

	solar, err := galaxy.CreateBucketIfNotExists(solar_name)
	if err != nil {
		logger.Errorf("Error while creating solar bucket %s err %s", string(solar_name), err)
		return err
	}
	_ = solar

	return err

}

func (b *BoltTXWrapper) CreateBucket(universe_name []byte, galaxy_name []byte, solar_name []byte) (err error) {
	return b.bdb.CreateBucket(b.tx, universe_name, galaxy_name, solar_name)
}

// a tx is shared by multiple goroutines, so they are protected by a mutex
func (b *BoltStore) LoadObject(tx *bolt.Tx, universe_name []byte, bucket_name []byte, solar_bucket []byte, key []byte) (data []byte, err error) {
	//rlog.Tracef(10, "Loading object %s %s %x", string(universe_name), string(bucket_name), key)

	b.Lock()
	defer b.Unlock()
        //b.rw.RLock()
        //defer b.rw.RUnlock()

	// open universe bucket
	{
		universe := tx.Bucket(universe_name)
		if universe == nil {
			return data, fmt.Errorf("No Such Universe %x", universe_name)
		}
		bucket := universe.Bucket(bucket_name)
		if bucket == nil {
			return data, fmt.Errorf("No Such Bucket %x", bucket_name)
		}

		solar := bucket.Bucket(solar_bucket)
		if solar == nil {
			return data, fmt.Errorf("No Such Bucket %x", solar_bucket)
		}

		// now lets find the object
		value := solar.Get(key)

		data = make([]byte, len(value), len(value))

		copy(data, value) // job done
	}

	return

}

func (b *BoltTXWrapper) LoadObject(universe_name []byte, bucket_name []byte, solar_bucket []byte, key []byte) (data []byte, err error) {
	return b.bdb.LoadObject(b.tx, universe_name, bucket_name, solar_bucket, key)
}

func (b *BoltStore) LoadObjects(tx *bolt.Tx, universe_name []byte, bucket_name []byte, solar_bucket []byte) (keys [][]byte, values [][]byte, err error) {
	//rlog.Tracef(10, "Loading object %s %s %x", string(universe_name), string(bucket_name), key)

	b.Lock()
	defer b.Unlock()

	// open universe bucket
	{
		universe := tx.Bucket(universe_name)
		if universe == nil {
			return keys, values, fmt.Errorf("No Such Universe %x", universe_name)
		}
		bucket := universe.Bucket(bucket_name)
		if bucket == nil {
			return keys, values, fmt.Errorf("No Such Bucket %x", bucket_name)
		}

		solar := bucket.Bucket(solar_bucket)
		if solar == nil {
			return keys, values, fmt.Errorf("No Such Bucket %x", solar_bucket)
		}

		// Create a cursor for iteration.
		cursor := solar.Cursor()

		// Iterate over items in sorted key order. This starts from the
		// first key/value pair and updates the k/v variables to the
		// next key/value on each iteration.
		//
		// The loop finishes at the end of the cursor when a nil key is returned.
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {

			key := make([]byte, len(k), len(k))
			value := make([]byte, len(v), len(v))
			copy(key, k)   // job done
			copy(value, v) // job done
			keys = append(keys, key)
			values = append(values, value)
			// fmt.Printf("A %s is %s.", k, v)
		}

	}

	return

}

func (b *BoltTXWrapper) LoadObjects(universe_name []byte, galaxy_name []byte, solar_name []byte) (keys [][]byte, values [][]byte, err error) {
	return b.bdb.LoadObjects(b.tx, universe_name, galaxy_name, solar_name)

}

// this function stores a uint64
// this will automcatically use the lock
func (b *BoltTXWrapper) StoreUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data uint64) error {
	return b.bdb.StoreObject(b.tx, universe_bucket, galaxy_bucket, solar_bucket, key, itob(data))

}

//  this function loads the data as 64 byte integer
func (b *BoltTXWrapper) LoadUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) (uint64, error) {
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
