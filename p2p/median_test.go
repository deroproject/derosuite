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

// these tests have been taken from the test_math.h from epee tests
func Test_Median(t *testing.T) {

	{ // test with 0 elements
		var array []uint64
		if Median(array) != 0 {
			t.Errorf("Testing failed\n")
		}
	}

	{ // test with 1 element
		array := []uint64{1}
		if Median(array) != 1 {
			t.Errorf("Testing failed\n")
		}
	}

	{ // test with 2 element
		array := []uint64{1, 10}
		if Median(array) != 5 {
			t.Errorf("Testing failed\n")
		}
	}

	{ // test with 3 element
		array := []uint64{0, 9, 3}
		if Median(array) != 3 {
			t.Errorf("Testing failed\n")
		}
	}

	{ // test with 4 element
		array := []uint64{77, 9, 22, 60}
		if Median(array) != 41 {
			t.Errorf("Testing failed\n")
		}
	}

	{ // test with 5 element
		array := []uint64{77, 9, 22, 60, 11}
		if Median(array) != 22 {
			t.Errorf("Testing failed\n")
		}
	}

}
