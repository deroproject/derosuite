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

package main

/*
import "io"
import "os"

import "fmt"
import "bytes"
import "bufio"
import "strings"
import "strconv"
import "runtime"
import "crypto/sha1"
import "encoding/hex"
import "encoding/json"
import "path/filepath"

import "github.com/romana/rlog"
import "github.com/chzyer/readline"
import "github.com/docopt/docopt-go"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/p2pv2"


import "github.com/deroproject/derosuite/config"

import "github.com/deroproject/derosuite/transaction"

//import "github.com/deroproject/derosuite/checkpoints"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/blockchain/rpcserver"
*/

//import "fmt"
import "time"
import "math/rand"

import "github.com/beevik/ntp"
import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/globals"

// these servers automatically rotate every hour as per documentation
// we also rotate them randomly
// TODO support ipv6
var timeservers = []string{
	"0.pool.ntp.org",
	"1.pool.ntp.org",
	"2.pool.ntp.org",
	"3.pool.ntp.org",
}

// continusosly checks time for deviation if possible
func time_check_routine() {

	// initial initial warning should NOT get hidden in messages

	random := rand.New(globals.NewCryptoRandSource())
	timeinsync := false
	for {

		if !timeinsync {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(2 * 60 * time.Second) // check every 2 minutes
		}

		server := timeservers[random.Int()%len(timeservers)]
		response, err := ntp.Query(server)
		if err != nil {
			rlog.Warnf("error while querying time  server %s err %s", server, err)
		} else {
			//globals.Logger.Infof("Local UTC time %+v server UTC time %+v", time.Now().UTC(), response.Time.UTC())
			if response.ClockOffset.Seconds() > -1.1 && response.ClockOffset.Seconds() < 1.1 {
				timeinsync = true
			} else {
				globals.Logger.Warnf("\nYour system time deviation is more than 1 secs (%s)."+
					"\nYou may experience chain sync issues and/or other side-effects."+
					"\nIf you are mining, your blocks may get rejected."+
					"\nPlease sync your system using NTP software (default availble in all OS)."+
					"\n eg. ntpdate pool.ntp.org  (for linux/unix)", response.ClockOffset)
			}
		}
	}
}
