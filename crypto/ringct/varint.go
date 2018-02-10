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

package ringct

import "io"
import "fmt"

// these functions basically what golang varint does, (  however theere is a minor spec diff, so these are here for exact compatibility)

func ReadVarInt(buf io.Reader) (result uint64, err error) {
	b := make([]byte, 1)
	var r uint64
	var n int
	for i := 0; ; i++ {
		n, err = buf.Read(b)
		if err != nil {
			return
		}
		if n != 1 {
			err = fmt.Errorf("Buffer ended prematurely for varint")
			return
		}
		r += (uint64(b[0]) & 0x7f) << uint(i*7)
		if uint64(b[0])&0x80 == 0 {
			break
		}
	}
	result = r
	return
}

func Uint64ToBytes(num uint64) (result []byte) {
	for ; num >= 0x80; num >>= 7 {
		result = append(result, byte((num&0x7f)|0x80))
	}
	result = append(result, byte(num))
	return
}
