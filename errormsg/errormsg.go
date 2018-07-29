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

package errormsg

import "fmt"

var ErrPanic = fmt.Errorf("Panic Occurred, keep cool")

// storage errors
var ErrInvalidStorageTX = fmt.Errorf("Invalid Storage Transaction")

// consensus errors
var ErrInvalidBlock = fmt.Errorf("Invalid Block")
var ErrInvalidPoW = fmt.Errorf("Invalid PoW")
var ErrAlreadyExists = fmt.Errorf("Already Exists")
var ErrPastMissing = fmt.Errorf("Past blocks are missing")
var ErrInvalidTimestamp = fmt.Errorf("Invalid Timestamp in past")
var ErrFutureTimestamp = fmt.Errorf("Invalid Timestamp in future")
var ErrTXDoubleSpend = fmt.Errorf("TX double spend attempt")
var ErrTXDead = fmt.Errorf("DEAD TX included in block")
var ErrInvalidSize = fmt.Errorf("Invalid Size")
var ErrInvalidTX = fmt.Errorf("Invalid TX")
