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

import "net/url"
import "strconv"
import "golang.org/x/net/proxy"
import "github.com/sirupsen/logrus"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/config"

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

/* this function converts a logrus entry into a txt formater based entry with no colors  for tracing*/
func CTXString(entry *logrus.Entry) string {

	entry.Level = logrus.DebugLevel
	data, err := ilog_formatter.Format(entry)
	_ = err
	return string(data)
}

// never do any division operation on money
// ISSUE: dividing 10 into 3 parts will give 3.33.. etc., adding result 3 times will 9.9999, so some money is lost
//
func FormatMoney(amount uint64) string {
	amountf := float64(amount) / 1000000000000.0 // float64 gives 14 char precision, we need only 12
	return strconv.FormatFloat(amountf, 'f', 12, 64)
}
