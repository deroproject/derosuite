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

package globals

import "os"
import "fmt"
import "math"
import "net/url"
import "strings"
import "strconv"
import "math/big"
import "path/filepath"
import "golang.org/x/net/proxy"
import "github.com/sirupsen/logrus"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/address"

type ChainState int // block chain can only be in 2 state, either SYNCRONISED or syncing

const (
	SYNCRONISED ChainState = iota // 0
	SYNCING                       // 1
)

// all the the global variables used by the program are stored here
// since the entire logic is designed around a state machine driven by external events
// once the core starts nothing changes until there is a network state change

var Incoming_Block = make([]byte, 100) // P2P feeds it, blockchain consumes it
var Outgoing_Block = make([]byte, 100) // blockchain feeds it, P2P consumes it  only if a block has been mined

var Incoming_Tx = make([]byte, 100) // P2P feeds it, blockchain consumes it
var Outgoing_Tx = make([]byte, 100) // blockchain feeds it, P2P consumes it  only if a  user has created a Tx mined

var Subsystem_Active uint32 // atomic counter to show how many subsystems are active
var Exit_In_Progress bool

// on init this variable is updated to setup global config in 1 go
var Config config.CHAIN_CONFIG

// global logger all components will use it with context

var Logger *logrus.Logger
var Log_Level = logrus.InfoLevel         // default is info level
var ilog_formatter *logrus.TextFormatter // used while tracing code

var Dialer proxy.Dialer = proxy.Direct // for proxy and direct connections
// all outgoing connections , including DNS requests must be made using this

// all program arguments are available here
var Arguments map[string]interface{}

func Initialize() {
	var err error
	_ = err

	Config = config.Mainnet // default is mainnnet

	if Arguments["--testnet"].(bool) == true { // setup testnet if requested
		Config = config.Testnet
	}

	// formatter := &logrus.TextFormatter{DisableColors : true}

	//Logger= &logrus.Logger{Formatter:formatter}
	Logger = logrus.New()

	//Logger.Formatter = &logrus.TextFormatter{DisableColors : true}

	Logger.SetLevel(logrus.InfoLevel)
	if Arguments["--debug"].(bool) == true { // setup debug mode if requested
		Log_Level = logrus.DebugLevel
		Logger.SetLevel(logrus.DebugLevel)
	}
	Logger.AddHook(&HOOK) // add rlog hook

	// choose  socks based proxy if user requested so
	if Arguments["--socks-proxy"] != nil {
		log.Debugf("Setting up proxy using %s", Arguments["--socks-proxy"].(string))
		//uri, err := url.Parse("socks5://127.0.0.1:9000") // "socks5://demo:demo@192.168.99.100:1080"
		uri, err := url.Parse("socks5://" + Arguments["--socks-proxy"].(string)) // "socks5://demo:demo@192.168.99.100:1080"
		if err != nil {
			log.Fatalf("Error parsing socks proxy: err %s", err)
		}

		Dialer, err = proxy.FromURL(uri, proxy.Direct)
		if err != nil {
			log.Fatalf("Error creating socks proxy: err \"%s\" from data %s ", err, Arguments["--socks-proxy"].(string))
		}
	}

	// windows and logrus have issues while printing colored messages, so disable them right now
	ilog_formatter = &logrus.TextFormatter{} // this needs to be created after after top logger has been intialised
	ilog_formatter.DisableColors = true
	ilog_formatter.DisableTimestamp = true

	// lets create data directories
	err = os.MkdirAll(GetDataDirectory(), 0750)
	if err != nil {
		fmt.Printf("Error creating/accessing directory %s , err %s\n", GetDataDirectory(), err)
	}

}

// tells whether we are in mainnet mode
// if we are not mainnet, we are a testnet,
// we will only have a single mainnet ,( but we may have one or more testnets )
func IsMainnet() bool {
	if Config.Name == "mainnet" {
		return true
	}

	return false
}

// return different directories for different networks ( mainly mainnet, testnet, simulation )
// this function is specifically for daemon
func GetDataDirectory() string {
	data_directory, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error obtaining current directory, using temp dir err %s\n", err)
		data_directory = os.TempDir()
	}

	// if user provided an option, override default
	if Arguments["--data-dir"] != nil {
		data_directory = Arguments["--data-dir"].(string)
	}

	if IsMainnet() {
		return filepath.Join(data_directory, "mainnet")
	}

	return filepath.Join(data_directory, "testnet")
}

/* this function converts a logrus entry into a txt formater based entry with no colors  for tracing*/
func CTXString(entry *logrus.Entry) string {

	entry.Level = logrus.DebugLevel
	data, _ := ilog_formatter.Format(entry)
	return string(data)
}

// never do any division operation on money due to floating point issues
// newbies, see type the next in python interpretor "3.33-3.13"
//
func FormatMoney(amount uint64) string {
	return FormatMoneyPrecision(amount, 8) // default is 8 precision after floating point
}

// 0
func FormatMoney0(amount uint64) string {
	return FormatMoneyPrecision(amount, 0)
}

//8 precision
func FormatMoney8(amount uint64) string {
	return FormatMoneyPrecision(amount, 8)
}

// 12 precision
func FormatMoney12(amount uint64) string {
	return FormatMoneyPrecision(amount, 12) // default is 8 precision after floating point
}

// format money with specific precision
func FormatMoneyPrecision(amount uint64, precision int) string {
	hard_coded_decimals := new(big.Float).SetInt64(1000000000000)
	float_amount, _, _ := big.ParseFloat(fmt.Sprintf("%d", amount), 10, 0, big.ToZero)
	result := new(big.Float)
	result.Quo(float_amount, hard_coded_decimals)
	return result.Text('f', precision) // 8 is display precision after floating point
}

// this will parse and validate an address, in reference to the current main/test mode
func ParseValidateAddress(str string) (addr *address.Address, err error) {
	addr, err = address.NewAddress(strings.TrimSpace(str))
	if err != nil {
		return
	}

	// check whether the domain is valid
	if !addr.IsDERONetwork() {
		err = fmt.Errorf("Invalid DERO address")
		return
	}

	if IsMainnet() != addr.IsMainnet() {
		if IsMainnet() {
			err = fmt.Errorf("Address belongs to DERO testnet and is invalid")
		} else {
			err = fmt.Errorf("Address belongs to DERO mainnet and is invalid")
		}
		return
	}

	return
}

// this will covert an amount in string form to atomic units
func ParseAmount(str string) (amount uint64, err error) {
	float_amount, base, err := big.ParseFloat(strings.TrimSpace(str), 10, 0, big.ToZero)

	if err != nil {
		err = fmt.Errorf("Amount could not be parsed err: %s", err)
		return
	}
	if base != 10 {
		err = fmt.Errorf("Amount should be in base 10 (0123456789)")
		return
	}
	if float_amount.Cmp(new(big.Float).Abs(float_amount)) != 0 { // number and abs(num) not equal means number is neg
		err = fmt.Errorf("Amount cannot be negative")
		return
	}

	// multiply by 12 zeroes
	hard_coded_decimals := new(big.Float).SetInt64(1000000000000)
	float_amount.Mul(float_amount, hard_coded_decimals)

	/*if !float_amount.IsInt() {
	    err =  fmt.Errorf("Amount  is invalid %s ", float_amount.Text('f',0))
	    return
	}*/

	// convert amount to uint64
	//amount, _ = float_amount.Uint64() // sanity checks again
	amount, err = strconv.ParseUint(float_amount.Text('f', 0), 10, 64)
	if err != nil {
		err = fmt.Errorf("Amount  is invalid %s ", float_amount.Text('f', 0))
		return
	}
	if amount == 0 {
		err = fmt.Errorf("0 cannot be transferred")
		return
	}

	if amount == math.MaxUint64 {
		err = fmt.Errorf("Amount  is invalid")
		return
	}

	return // return the number

}
