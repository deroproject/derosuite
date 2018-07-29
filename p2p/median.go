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

import "sort"

type uint64arr []uint64

func (a uint64arr) Len() int           { return len(a) }
func (a uint64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a uint64arr) Less(i, j int) bool { return a[i] > a[j] } // we sort descending by inversing the condition

// Median gets the median number in a slice of uint64 numbers
// this function is used in so many checks that it should have unit tests for sure
// NOTE: it will the input array

func Median(inputs []uint64) (median uint64) {

	// Start by sorting
	sort.Sort(uint64arr(inputs))

	// No math is needed if there are no numbers
	// For even numbers we add the two middle numbers
	// and divide by two using the mean function above
	// For odd numbers we just use the middle number
	l := len(inputs)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (inputs[(l/2)-1] + inputs[(l/2)]) / 2
	} else {
		median = inputs[l/2]
	}

	return median
}
