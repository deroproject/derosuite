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
import "runtime"
import "path/filepath"
import "encoding/binary"

//import "github.com/romana/rlog"
//import bolt "github.com/coreos/bbolt"
import "github.com/dgraph-io/badger"
import log "github.com/sirupsen/logrus"
import "github.com/dgraph-io/badger/options"

import "github.com/deroproject/derosuite/globals"

type BadgerDBStore struct {
	DB         *badger.DB
	sync.Mutex // lock this struct TODO we must user a reader lock
}

// this object is returned
type BadgerTXWrapper struct {
	bdb *BadgerDBStore
	tx  *badger.Txn
}

var Badger_backend *BadgerDBStore = &BadgerDBStore{} // global variable

func (b *BadgerDBStore) Init(params map[string]interface{}) (err error) {
	logger = globals.Logger.WithFields(log.Fields{"com": "STORE"})
	current_path := filepath.Join(globals.GetDataDirectory(), "derod_badger_database")

	if params["--simulator"] == true {
		current_path = filepath.Join(os.TempDir(), "derod_simulation_badger_database") // sp
	}
	logger.Infof("Initializing badger store at path %s", current_path)

	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.

	opts := badger.DefaultOptions
	opts.Dir = current_path
	opts.ValueDir = current_path

	// tune to different RAM requirements based on arch
	if runtime.GOARCH == "amd64" {
		opts.TableLoadingMode = options.MemoryMap
		opts.ValueLogLoadingMode = options.MemoryMap
	} else { // 32 bit systems/arm etc use raw file IO, no memory mapping
		opts.TableLoadingMode = options.FileIO
		opts.ValueLogLoadingMode = options.FileIO
		logger.Infof("Running in low RAM mode")
	}

	db, err := badger.Open(opts)
	if err != nil {
		logger.Fatalf("Cannot open badgerdb store err %s", err)
	}
	b.DB = db

	// if simulation, delete the file , so as it gets cleaned up automcatically
	if params["--simulator"] == true {
		os.RemoveAll(current_path)
	}

	// place db in no sync mode
	//b.DB.NoSync = true

	return nil
}

func (b *BadgerDBStore) Shutdown() (err error) {
	logger.Infof("Shutting badgerdb store")
	if b.DB != nil {
		b.DB.Close()
	}

	return nil
}

// sync the DB to disk
func (b *BadgerDBStore) Sync() {

}

// get a new writable/readable tx,
// we will manage the writable txs manually
// since a block may cause changes to a number of fields which must be reflected atomically
func (b *BadgerDBStore) BeginTX(writable bool) (DBTX, error) {

	txwrapper := &BadgerTXWrapper{}
	tx := b.DB.NewTransaction(writable) // begin a new writable tx
	txwrapper.tx = tx
	txwrapper.bdb = b // parent DB reference

	return txwrapper, nil
}

func (b *BadgerTXWrapper) Commit() error {
	err := b.tx.Commit(nil)
	if err != nil {
		logger.Warnf("Error while committing tx, err %s", err)
		return err
	}
	b.tx.Discard() // tx must always be discarded (both if commited/rollback as per documentation)

	return nil
}

// Roll back existing changes to  disk
func (b *BadgerTXWrapper) Rollback() {

	b.tx.Discard()
}

// TODO implement this
func (b *BadgerTXWrapper) Sync() {

}

// duplicates a byte array as badger db needs objects which are unmodifieable during the transaction
func Duplicate(input []byte) []byte {
	dup := make([]byte, len(input), len(input))
	copy(dup, input)
	return dup
}

func (b *BadgerTXWrapper) StoreObject(universe_name []byte, galaxy_name []byte, solar_name []byte, key []byte, data []byte) (err error) {
	fullkey := make([]byte, 0, len(universe_name)+len(galaxy_name)+len(solar_name)+len(key))

	fullkey = append(fullkey, universe_name...)
	fullkey = append(fullkey, galaxy_name...)
	fullkey = append(fullkey, solar_name...)
	fullkey = append(fullkey, key...)
	return b.tx.Set(fullkey, Duplicate(data))

}

func (b *BadgerTXWrapper) LoadObject(universe_name []byte, galaxy_name []byte, solar_name []byte, key []byte) (data []byte, err error) {

	fullkey := make([]byte, 0, len(universe_name)+len(galaxy_name)+len(solar_name)+len(key))

	fullkey = append(fullkey, universe_name...)
	fullkey = append(fullkey, galaxy_name...)
	fullkey = append(fullkey, solar_name...)
	fullkey = append(fullkey, key...)

	item, err := b.tx.Get(fullkey)
	if err == badger.ErrKeyNotFound {
		return data, badger.ErrKeyNotFound
	}
	data, err = item.ValueCopy(nil)
	if err != nil {
		return data, badger.ErrKeyNotFound
	}
	return data, nil
}

// this function stores a uint64
// this will automcatically use the transaction
func (b *BadgerTXWrapper) StoreUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data uint64) error {
	return b.StoreObject(universe_bucket, galaxy_bucket, solar_bucket, key, itob(data))

}

//  this function loads the data as 64 byte integer
func (b *BadgerTXWrapper) LoadUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) (uint64, error) {
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
