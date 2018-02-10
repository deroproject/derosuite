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

// 32 bit and other systems will use this

// +build !amd64 goleveldb

package storage

import "os"
import "fmt"
import "sync"
import "path/filepath"
import "encoding/binary"

import "github.com/romana/rlog"
import "github.com/syndtr/goleveldb/leveldb"
import "github.com/syndtr/goleveldb/leveldb/opt"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"

type BoltStore struct {
	DB         *leveldb.DB
	tx         *leveldb.Transaction
	sync.Mutex // lock this struct
}

var Bolt_backend *BoltStore = &BoltStore{} // global variable
var logger *log.Entry

func (b *BoltStore) Init(params map[string]interface{}) (err error) {
	logger = globals.Logger.WithFields(log.Fields{"com": "STORE"})
	current_path := filepath.Join(os.TempDir(), "derod_leveldb_database.db")
	logger.Infof("Initializing leveldb store at path %s", current_path)

	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	options := opt.Options{
		CompactionL0Trigger: 32,               // default value is 4
		CompactionTableSize: 64 * 1024 * 1024, // default is 2 Mib
		WriteBuffer:         16 * 1024 * 1024, // default is 4 Mib

	}
	b.DB, err = leveldb.OpenFile(current_path, &options)
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

// get a new writable tx,
// we will manage the writable txs manually
// since a block may cause changes to a number of fields which must be reflected atomically
// this function is always triggered while the atomic lock is taken
// this is done avoid a race condition in returning the tx and using it
func (b *BoltStore) get_new_writable_tx() (tx *leveldb.Transaction) {
	if b.tx != nil {
		tx = b.tx // use existing pending tx
	} else { // create new pending tx

		tx, err := b.DB.OpenTransaction() // begin a new writable tx
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
func (b *BoltStore) Rollback() {
	b.Lock()
	if b.tx != nil {
		rlog.Tracef(1, "Rollbacking writable TX")
		b.tx.Discard()
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
		//b.DB.Sync() // sync the DB
	}
	b.Unlock()
}

func (b *BoltStore) StoreObject(universe_name []byte, galaxy_name []byte, solar_name []byte, key []byte, data []byte) (err error) {

	b.Lock()
	defer b.Unlock()

	rlog.Tracef(10, "Storing object %s %s %x  data len %d\n", string(universe_name), string(galaxy_name), key, len(data))
	// open universe bucket
	tx := b.get_new_writable_tx()

	keyname := append(universe_name, append(galaxy_name, append(solar_name, key...)...)...)
	// now lets update the object attribute
	err = tx.Put(keyname, data, nil)

	return err

}

func (b *BoltStore) LoadObject(universe []byte, bucket_name []byte, solar_bucket []byte, key []byte) (data []byte, err error) {
	rlog.Tracef(10, "Loading object %s %s %x\n", string(universe), string(bucket_name), key)

	b.Lock()
	defer b.Unlock()

	var tx *leveldb.Transaction
	if b.tx != nil {
		tx = b.tx
	} else {
		tx, err = b.DB.OpenTransaction() // create writable tx and discard it
		defer tx.Discard()               // tx must be rolled back always, never commit
	}
	// open universe bucket
	{
		keyname := append(universe, append(bucket_name, append(solar_bucket, key...)...)...)

		// now lets find the object
		data, err = tx.Get(keyname, nil)
	}

	return

}

// this function stores a uint64
// this will automcatically use the lock
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
