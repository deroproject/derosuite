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

package p2p

import "testing"
import "encoding/hex"

func Test_Levin_Header_Deserialisation(t *testing.T) {

	// this structure has been manually pulled from wireshark stream
	raw_data := "0121010101010101e20000000000000001e9030000000000000100000001000000"
	raw_data_blob, _ := hex.DecodeString(raw_data)

	_ = raw_data_blob

	var lheader Levin_Header

	err := lheader.DeSerialize(raw_data_blob)

	if err != nil {
		t.Error("DeSerialize Levin header Failed\n")
		return
	}

	// now serialize once again

	serialised, _ := lheader.Serialize()
	if raw_data != hex.EncodeToString(serialised) {
		t.Error("Serialize Levin_Header Failed")
	}

}

func Test_Levin_Data_Header(t *testing.T) {

	// this structure has been manually pulled from wireshark stream
	raw_data := "01110101010102010100"
	raw_data_blob, _ := hex.DecodeString(raw_data)

	_ = raw_data_blob

	var lheader Levin_Data_Header

	err := lheader.DeSerialize(raw_data_blob)

	if err != nil {
		t.Error("DeSerialize Levin Data header Failed\n")
		return
	}

	// now serialize once again

	serialised, _ := lheader.Serialize()
	if raw_data != hex.EncodeToString(serialised) {
		t.Errorf("Serialize Levin_Data_Header Failed \n%s correct value \n%s Our value", raw_data, hex.EncodeToString(serialised))
	}
}
