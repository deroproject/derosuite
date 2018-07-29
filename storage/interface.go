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

package storage

/*type DB struct {
    TX interface{} // actual TX object
}*/

// complete transactional support for improved reliability
// TODO do we need to support isolation level ?????
type DBTX interface {

	// Begin(bool)  // whether to create a writable tx or readable tx
	Commit() error // commit all writes persistantly
	Rollback()     // Rollback all changes since last commit
	Sync()         // sync the data

	StoreObject(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data []byte) error // store object to a specific universe
	LoadObject(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) ([]byte, error)     // load object from a specific universe

	//  LoadObjects(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte) ([][]byte, [][]byte, error)     // load all key values for specific bucket

	StoreUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data uint64) error // store
	LoadUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) (uint64, error)     // load object
	// CreateBucket(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte) error // creates an object bucket

}

type Store interface {
	Init(param map[string]interface{}) error // init the backend and connect to it
	Shutdown() error                         // shutdown the backend

	BeginTX(bool) (DBTX, error) // actual TX object to interact with DB
	/*
	   	Commit()   // commit all writes persistantly
	   	Rollback() // Rollback all changes since last commit
	   	Sync()     // sync the data

	   	StoreObject(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data []byte) error // store object to a specific universe
	   	LoadObject(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) ([]byte, error)     // load object from a specific universe

	           LoadObjects(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte) ([][]byte, [][]byte, error)     // load all key values for specific bucket

	   	StoreUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte, data uint64) error // store
	   	LoadUint64(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte, key []byte) (uint64, error)     // load object
	           CreateBucket(universe_bucket []byte, galaxy_bucket []byte, solar_bucket []byte) error // creates an object bucket
	*/
}

//var  Store Backend_Store// the system shouls chooose a backend at start Time
