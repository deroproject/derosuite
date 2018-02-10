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

// this files defines all the templates

var header_template string = `
{{define "header"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <META HTTP-EQUIV="CACHE-CONTROL" CONTENT="NO-CACHE">
    <title>{{ .title }}</title>
    <!--<link rel="stylesheet" type="text/css" href="/css/style.css">-->
    <style type="text/css">
body {
    margin: 0;
    padding: 0;
    color: green;
    background-color: white;
}

h1, h2, h3, h4, h5, h6 {
    text-align: center;
}

.center {
    margin: auto;
    width: 96%;
    /*border: 1px solid #73AD21;
    padding: 10px;*/
}

tr, li, #pages, .info {
    font-family: "Lucida Console", Monaco, monospace;
    font-size  : 12px;
    height: 22px;
}

#pages
{
    margin-top: 15px;
}

td {
    text-align: center;
}

a:link {
    text-decoration: none;
    color: blue;
}

a:visited {
    text-decoration: none;
    color: blue;
}

a:hover {
    text-decoration: underline;
    color: blue;
}

a:active {
    text-decoration: none;
    color: blue;
}

form {
    display: inline-block;
    text-align: center;
}

.style-1 input[type="text"] {
    padding: 2px;
    border: solid 1px #dcdcdc;
    transition: box-shadow 0.3s, border 0.3s;
}
.style-1 input[type="text"]:focus,
.style-1 input[type="text"].focus {
    border: solid 1px #707070;
    box-shadow: 0 0 5px 1px #969696;
}


.tabs {
    position: relative;
    min-height: 220px; /* This part sucks */
    clear: both;
    margin: 25px 0;
}

.tab {
    float: left;
}

.tab label {
    background: white;
    padding: 10px;
    border: 1px solid #ccc;
    margin-left: -1px;
    position: relative;
    left: 1px;
}

.tab [type=radio] {
    display: none;
}

.content {
    position: absolute;
    top: 28px;
    left: 0;
    background: white;
    right: 0;
    bottom: 0;
    padding: 20px;
    border: 1px solid #ccc;
}

[type=radio]:checked ~ label {
    background: #505050 ;
    border-bottom: 1px solid green;
    z-index: 2;
}

[type=radio]:checked ~ label ~ .content {
    z-index: 1;
}

input#toggle-1[type=checkbox] {
    position: absolute;
    /*top: -9999px;*/
    left: -9999px;

}
label#show-decoded-inputs {
    /*-webkit-appearance: push-button;*/
    /*-moz-appearance: button;*/
    display: inline-block;
    /*margin: 60px 0 10px 0;*/
    cursor: pointer;
    background-color: white;;
    color: green;
    width: 100%;
    text-align: center;
}

div#decoded-inputs{
    display: none;
}

/* Toggled State */
input#toggle-1[type=checkbox]:checked ~ div#decoded-inputs {
    display: block;
}
    </style>

</head>
<body>
<div>

    <div class="center">
        <h1 class="center"><a href="/">{{ .title }} {{if .testnet}} TestNet  {{end}}</a></h1>
<!--        <h4 style="font-size: 15px; margin: 0px">(no javascript - no cookies - no web analytics trackers - no images - open sourced)</h4> -->
    </div>


    <div class="center">
        <form action="/search" method="get" style="width:100%; margin-top:15px" class="style-1">
            <input type="text" name="value" size="120"
                   placeholder="block height, block hash, transaction hash">
            <input type="submit" value="Search">
        </form>
    </div>

</div>

{{if .Network_Difficulty}}
<div class="center">
     <h3 style="font-size: 12px; margin-top: 20px">
        Server time: {{ .servertime }}  | <a href="/txpool">Transaction pool</a>
        </h3>


        <h3 style="font-size: 12px; margin-top: 5px; margin-bottom: 3">
            Network difficulty: {{ .Network_Difficulty }}
            | Hash rate: {{ .hash_rate }} MH&#x2F;s
            | Mempool size : {{ .txpool_size }}
<!--            | Fee per kb: 0.001198930000
            | Median block size limit: 292.97 kB
-->
        </h3>

</div>
{{end}}
{{end}}
`

var block_template string = `{{define "block"}}
{{ template "header" . }}
<div>
    <H4>Block hash (height): {{.block.Hash}} ({{.block.Height}})</H4>
    <H5>Previous block: <a href="/block/{{.block.Prev_Hash}}">{{.block.Prev_Hash}}</a></H5>
<!--    
    <H5>Next block: <a href="/block/a8ade20d5cad5e23105cfc25687beb2498844a984b1450330c67705b6c720596">a8ade20d5cad5e23105cfc25687beb2498844a984b1450330c67705b6c720596</a></H5>
    -->
    <table class="center">
        <tr>
            <td>Timestamp [UCT] (epoch):</td><td>{{.block.Block_time}} ({{.block.Epoch}})</td>
            <td>Age [h:m:s]:</td><td>{{.block.Age}}</td>
            <td>Δ [h:m:s]:</td><td></td>
        </tr>
        <tr>
            <td>Major.minor version:</td><td>{{.block.Major_Version}}.{{.block.Minor_Version}}</td>
            <td>Block reward:</td><td>{{.block.Reward}}</td>
            <td>Block size [kB]:</td><td>{{.block.Size}}</td>
        </tr>
        <tr>
            <td>nonce:</td><td>{{.block.Nonce}}</td>
            <td>Total fees:</td><td>{{.block.Fees}}</td>
            <td>No of txs:</td><td>{{.block.Tx_Count}}</td>
        </tr>
    </table>

    <h3>Miner reward transaction</h3>
    <table class="center">
        <tr>
            <td>hash</td>
            <td>outputs</td>
            <td>size [kB]</td>
            <td>version</td>
        </tr>
            <tr>
                <td><a href="/tx/{{.block.Mtx.Hash}}">{{.block.Mtx.Hash}}</a>
                <td>{{.block.Mtx.Amount}}</td>
                <td>{{.block.Mtx.Size}}</td>
                <td>{{.block.Mtx.Version}}</td>
            </tr>

    </table>

    <h3>Transactions ({{.block.Tx_Count}})</h3>
        <table class="center" style="width:80%">
            <tr>
                <td>hash</td>
                <td>outputs</td>
                <td>fee</td>
                <td>ring size</td>
                <td>in/out</td>
                
                <td>version</td>
                <td>size [kB]</td>
            </tr>
            {{range .block.Txs}}
                <tr>
                    <td><a href="/tx/{{.Hash}}">{{.Hash}}</a></td>
                    <td>?</td>
                    <td>{{.Fee}}</td>
                    <td>{{.Ring_size}}</td>
                    <td>{{.In}}/{{.Out}}</td>                    
                    <td>{{.Version}}</td>
                    <td>{{.Size}}</td>
                </tr>
                {{end}}
        </table>

</div>

{{ template "footer" . }} 
{{end}}
`

var tx_template string = `{{define "tx"}}
{{ template "header" . }}
            <div>
                <H4 style="margin:5px">Tx hash: {{.info.Hash}}</H4>
                <H5 style="margin:5px">Tx prefix hash: {{.info.PrefixHash}}</H5>
                <H5 style="margin:5px">Tx public key: TODO</H5>
  
                <table class="center" style="width: 80%; margin-top:10px">
                    <tr>
                        <td>Timestamp: {{.info.Timestamp}} </td>
                        <td>Timestamp [UCT]: {{.info.Block_time}}</td>
                        <td>Age [y:d:h:m:s]: {{.info.Age}} </td>
                    </tr>
                    <tr>
                        <td>Block: <a href="/block/{{.info.Height}}">{{.info.Height}}</a></td>
                        <td>Fee: {{.info.Fee}}</td>
                        <td>Tx size: {{.info.Size}} kB</td>
                    </tr>
                    <tr>
                        <td>Tx version: {{.info.Version}}</td>
                        <td>No of confirmations: {{.info.Depth}}</td>
                        <td>Signature type:  {{.info.Type}}</td>
                    </tr>
                    <tr>
                        <td colspan="3">Extra: {{.info.Extra}}</td>
                    </tr>
                </table>
              <h3>{{.info.Out}} output(s) for total of {{.info.Amount}} dero</h3>
              <div class="center">
                  <table class="center">
                      <tr>
                          <td>stealth address</td>
                          <td>amount</td>
                          <td>amount idx</td>
                      </tr>
                      
                      {{range $i, $e := .info.OutAddress}} 
                      <tr>
                          <td>{{ $e }}</td>
                          <td>{{$.info.Amount}}</td>
                          <td>{{index $.info.OutOffset $i}}</td>
                      </tr>
                      {{end}}
                  </table>
              </div>
              
<!-- TODO currently we donot enable user to prove or decode something -->

{{if eq .info.CoinBase false}}

                    <h3>{{.info.In}} input(s) for total of ? dero</h3>
                <div class="center">
                   <table class="center">
                   {{range .info.Keyimages}} 
                    <tr>
                      <td style="text-align: center;">
                          key image  {{ . }}
                      </td>
                      <td>amount: ?</td>
                   </tr>
                   {{end}}
                  </table>
              </div>
            {{end}}
           </div>
{{ template "footer" . }} 
              
{{end}}`

var txpool_template string = `{{define "txpool"}}
<h2 style="margin-bottom: 0px">
   Transaction pool
</h2>
<h4 style="font-size: 12px; margin-top: 0px">(no of txs: {{ .txpool_size }}, size: 0.00 kB, updated every 5 seconds)</h4>
<div class="center">
    
      <table class="center" style="width:80%">
            <tr>
                <td>age [h:m:s]</td>
                <td>transaction hash</td>
                <td>fee</td>
                <td>outputs</td>
                <td>in(nonrct)/out</td>
                <td>ring size</td>
                <td>tx size [kB]</td>
            </tr>
            
           
            {{range .mempool}}
            <tr>
                    <td></td>
                    <td><a href="/tx/{{.Hash}}">{{.Hash}}</a></td>
                    <td>{{.Fee}}</td>
                    <td>N/A</td>
                    <td>{{.In}}/{{.Out}}</td>
                    <td>{{.Ring_size}}</td>
                    <td>{{.Size}}</td>

            </tr>
            
            {{end}}
        </table>



</div>
{{end}}`

// full page txpool_template
var txpool_page_template string = `{{define "txpool_page"}}
{{ template "header" . }}
{{ template "txpool" . }}
{{ template "footer" . }}  
{{end}}`

var main_template string = `
{{define "main"}}
{{ template "header" . }}
{{ template "txpool" . }}

      <h2 style="margin-bottom: 0px">Transactions in the last 11 blocks</h2>

    <h4 style="font-size: 14px; margin-top: 0px">(Median size of these blocks: 0.09 kB)</h4>

    <div class="center">

            <table class="center">
                <tr>
                    <td>height</td>
                    <td>age [h:m:s]<!--(Δm)--></td>
                    <td>size [kB]<!--(Δm)--></td>
                    <td>tx hash</td>
                    <td>fees</td>
                    <td>outputs</td>
                    <td>in(nonrct)/out</td>
                    <td>ring size</td>
                    <td>tx size [kB]</td>
                </tr>
                
                
                {{range .block_array}}
                <tr>
                    <td><a href="/block/{{.Height}}">{{.Height}}</a></td>
                       <td>{{.Age}}</td>
                    <td>{{.Size}}</td>
                    <td><a href="/tx/{{.Mtx.Hash}}">{{.Mtx.Hash}}</a></td>
                    <td>N/A</td>
                    <td>{{.Mtx.Amount}}</td>
                    <td>{{.Mtx.In}}/{{.Mtx.Out}}</td>
                    <td>0</td>
                    <td>{{.Mtx.Size}}</td>

                </tr>
                
                {{range .Txs}}
                <tr>
                    <td></td>
                       <td></td>
                    <td></td>
                    <td><a href="/tx/{{.Hash}}">{{.Hash}}</a></td>
                    <td>{{.Fee}}</td>
                    <td>N/A</td>
                    <td>{{.In}}/{{.Out}}</td>
                    <td>{{.Ring_size}}</td>
                    <td>{{.Size}}</td>

                </tr>
                {{end}}
                
                {{end}}
              </table>  
        {{ template "paging" . }}
        
        </div>
       {{ template "footer" . }}         
{{end}}`

var paging_template string = `{{ define "paging"}}

<div id="pages" class="center" style="text-align: center;">
               <a href="/page/{{.previous_page}}">previous page</a> |
               <a href="/">first page</a> |
                current page: {{.current_page}}/<a href="/page/{{.total_page}}">{{.total_page}}</a>
                | <a href="/page/{{.next_page}}">next page</a>
            </div>

{{end}}`

var footer_template string = ` {{define "footer"}}
<div class="center">
    <h6 style="margin-top:10px">
        <a href="https://github.com/deroproject/">DERO explorer source code</a>
        | explorer version (api): under development (1.0)
        | dero version: golang pre-alpha
        | Copyright 2017-2018  Dero Project
    </h6>
</div>
</body>
</html>
{{end}}
`
