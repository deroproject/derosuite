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

package rpcserver

// get block template handler not implemented

//import "fmt"
import "context"

//import	"log"
//import 	"net/http"

import "encoding/hex"
import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/structures"

type SubmitBlock_Handler struct{}

func (h SubmitBlock_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	// parameter is an array of blockdata
	var block_data [2]string

	//logger.Infof("Submitting block results")

	if err := jsonrpc.Unmarshal(params, &block_data); err != nil {
                logger.Warnf("Submitted block could be json parsed")
		return nil, err
	}

	block_data_bytes, err := hex.DecodeString(block_data[0])
	if err != nil {
		logger.Infof("Submitting block could not be decoded")
		return structures.SubmitBlock_Result{
			Status: "Could NOT decode block",
		}, nil
	}

	hashing_blob, err := hex.DecodeString(block_data[1])
	if err != nil || len(block_data[1]) == 0 {
		logger.Infof("Submitting block hashing_blob could not be decoded")
		return structures.SubmitBlock_Result{
			Status: "Could NOT decode block",
		}, nil
	}

	blid, result, err := chain.Accept_new_block(block_data_bytes, hashing_blob)

	if result {
		logger.Infof("Submitted block %s accepted",blid)
		return structures.SubmitBlock_Result{
                        BLID: blid.String(),
			Status: "OK",
		}, nil
	}

	if err != nil {
		logger.Infof("Submitting block %s err %s",blid, err)
		return structures.SubmitBlock_Result{
			Status: err.Error(),
		}, nil
	}

	logger.Infof("Submitting block rejected err %s", err)
	return structures.SubmitBlock_Result{
		Status: "REJECTED",
	}, nil

}
