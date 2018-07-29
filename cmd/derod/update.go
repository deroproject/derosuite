package main

//import "fmt"
import "net"
import "time"
import "io"

//import "io/ioutil"
//import "net/http"
import "context"
import "strings"
import "math/rand"
import "encoding/base64"
import "encoding/json"
import "runtime/debug"
import "encoding/binary"

//import "crypto/tls"

import "github.com/blang/semver"
import "github.com/miekg/dns"
import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/globals"

/* this needs to be set on update.dero.io. as TXT record,  in encoded form as base64
 *
{ "version" : "1.0.2",
  "message" : "\n\n\u001b[32m This is a mandatory update\u001b[0m",
  "critical" : ""
}

base64 eyAidmVyc2lvbiIgOiAiMS4wLjIiLAogIm1lc3NhZ2UiIDogIlxuXG5cdTAwMWJbMzJtIFRoaXMgaXMgYSBtYW5kYXRvcnkgdXBkYXRlXHUwMDFiWzBtIiwgCiJjcml0aWNhbCIgOiAiIiAKfQ==



TXT record should be set as update=eyAidmVyc2lvbiIgOiAiMS4wLjIiLAogIm1lc3NhZ2UiIDogIlxuXG5cdTAwMWJbMzJtIFRoaXMgaXMgYSBtYW5kYXRvcnkgdXBkYXRlXHUwMDFiWzBtIiwgCiJjcml0aWNhbCIgOiAiIiAKfQ==
*/

func check_update_loop() {

	for {

		if config.DNS_NOTIFICATION_ENABLED {

			globals.Logger.Debugf("Checking update..")
			check_update()
		}
		time.Sleep(2 * 3600 * time.Second) // check every 2 hours
	}

}

// wrapper to make requests using proxy
func dialContextwrapper(ctx context.Context, network, address string) (net.Conn, error) {
	return globals.Dialer.Dial(network, address)
}

type socks_dialer net.Dialer

func (d *socks_dialer) Dial(network, address string) (net.Conn, error) {
	globals.Logger.Infof("Using our dial")
	return globals.Dialer.Dial(network, address)
}

func (d *socks_dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	globals.Logger.Infof("Using our context dial")
	return globals.Dialer.Dial(network, address)
}

func dial_random_read_response(in []byte) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			rlog.Warnf("Recovered while checking updates, Stack trace below", r)
			rlog.Warnf("Stack trace  \n%s", debug.Stack())
		}
	}()

	// since we may be connecting through socks, grab the remote ip for our purpose rightnow
	//conn, err := globals.Dialer.Dial("tcp", "208.67.222.222:53")
	//conn, err := net.Dial("tcp", "8.8.8.8:53")
	random_feeder := rand.New(globals.NewCryptoRandSource())                          // use crypto secure resource
	server_address := config.DNS_servers[random_feeder.Intn(len(config.DNS_servers))] // choose a random server cryptographically
	conn, err := net.Dial("tcp", server_address)

	//conn, err := tls.Dial("tcp", remote_ip.String(),&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		rlog.Warnf("Dial failed err %s", err.Error())
		return
	}

	defer conn.Close() // close connection at end

	// upgrade connection TO TLS ( tls.Dial does NOT support proxy)
	//conn = tls.Client(conn, &tls.Config{InsecureSkipVerify: true})

	rlog.Tracef(1, "Sending %d bytes", len(in))

	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(len(in)))
	conn.Write(buf[:]) // write length in bigendian format

	conn.Write(in) // write data

	// now we must wait for response to arrive
	var frame_length_buf [2]byte

	conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	nbyte, err := io.ReadFull(conn, frame_length_buf[:])
	if err != nil || nbyte != 2 {
		// error while reading from connection we must disconnect it
		rlog.Warnf("Could not read DNS length prefix err %s", err)
		return
	}

	frame_length := binary.BigEndian.Uint16(frame_length_buf[:])
	if frame_length == 0 {
		// most probably memory DDOS attack, kill the connection
		rlog.Warnf("Frame length is too small")
		return
	}

	out = make([]byte, frame_length)

	conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	data_size, err := io.ReadFull(conn, out)
	if err != nil || data_size <= 0 || uint16(data_size) != frame_length {
		// error while reading from connection we must kiil it
		rlog.Warnf("Could not read  DNS data size  read %d, frame length %d err %s", data_size, frame_length, err)
		return

	}
	out = out[:frame_length]

	return
}

func check_update() {

	// add panic handler, in case DNS acts rogue and tries to attack
	defer func() {
		if r := recover(); r != nil {
			rlog.Warnf("Recovered while checking updates, Stack trace below", r)
			rlog.Warnf("Stack trace  \n%s", debug.Stack())
		}
	}()

	if !config.DNS_NOTIFICATION_ENABLED { // if DNS notifications are disabled bail out
		return
	}

	/*  var u update_message
	    u.Version = "2.0.0"
	    u.Message = "critical msg txt\x1b[35m should \n be in RED"

	     globals.Logger.Infof("testing %s",u.Message)

	     j,err := json.Marshal(u)
	     globals.Logger.Infof("json format %s err %s",j,err)
	*/
	/*extract_parse_version("update=eyAidmVyc2lvbiIgOiAiMS4xLjAiLCAibWVzc2FnZSIgOiAiXG5cblx1MDAxYlszMm0gVGhpcyBpcyBhIG1hbmRhdG9yeSB1cGdyYWRlIHBsZWFzZSB1cGdyYWRlIGZyb20geHl6IFx1MDAxYlswbSIsICJjcml0aWNhbCIgOiAiIiB9")

	  return
	*/

	m1 := new(dns.Msg)
	// m1.SetEdns0(65000, true), dnssec probably leaks current timestamp, it's disabled until more invetigation
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{config.DNS_UPDATE_CHECK, dns.TypeTXT, dns.ClassINET}

	packed, err := m1.Pack()
	if err != nil {
		globals.Logger.Warnf("Error which packing DNS query for program update err %s", err)
		return
	}

	/*

			// setup a http client
			httpTransport := &http.Transport{}
			httpClient := &http.Client{Transport: httpTransport}
			// set our socks5 as the dialer
			httpTransport.Dial = globals.Dialer.Dial



		        packed_base64:= base64.RawURLEncoding.EncodeToString(packed)
		response, err := httpClient.Get("https://1.1.1.1/dns-query?ct=application/dns-udpwireformat&dns="+packed_base64)

		_ = packed_base64

		if err != nil {
		    rlog.Warnf("error making DOH request err %s",err)
		    return
		}

		defer response.Body.Close()
		        contents, err := ioutil.ReadAll(response.Body)
		        if err != nil {
		            rlog.Warnf("error reading DOH response err %s",err)
		            return
		}
	*/

	contents, err := dial_random_read_response(packed)
	if err != nil {
		rlog.Warnf("error reading response from DNS server err %s", err)
		return

	}

	rlog.Debugf("DNS response length from DNS server %d bytes", len(contents))

	err = m1.Unpack(contents)
	if err != nil {
		rlog.Warnf("error decoding DOH response err %s", err)
		return

	}

	for i := range m1.Answer {
		if t, ok := m1.Answer[i].(*dns.TXT); ok {

			// replace any spaces so as records could be joined

			rlog.Tracef(1, "Process record %+v", t.Txt)
			joined := strings.Join(t.Txt, "")
			extract_parse_version(joined)

		}
	}

	//globals.Logger.Infof("response %+v err ",m1,err)

}

type update_message struct {
	Version  string `json:"version"`
	Message  string `json:"message"`
	Critical string `json:"critical"` // always broadcasted, without checks for version
}

// our version are TXT record of following format
// version=base64 encoded json
func extract_parse_version(str string) {

	strl := strings.ToLower(str)
	if !strings.HasPrefix(strl, "update=") {
		rlog.Tracef(1, "Skipping record %s", str)
		return
	}

	parts := strings.SplitN(str, "=", 2)
	if len(parts) != 2 {
		return
	}
	rlog.Tracef(1, "parts %s", parts[1])

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		rlog.Tracef(1, "Could NOT decode base64 update message %s", err)
		return
	}

	var u update_message
	err = json.Unmarshal(data, &u)

	//globals.Logger.Infof("data %+v", u)

	if err != nil {
		rlog.Tracef(1, "Could NOT decode json update message %s", err)
		return
	}

	uversion, err := semver.ParseTolerant(u.Version)
	if err != nil {
		rlog.Tracef(1, "Could NOT update version %s", err)
	}

	current_version := config.Version
	current_version.Pre = current_version.Pre[:0]
	current_version.Build = current_version.Build[:0]

	// give warning to update the daemon
	if u.Message != "" && err == nil { // check semver
		if current_version.LT(uversion) {
			if current_version.Major != uversion.Major { // if major version is different give extract warning
				globals.Logger.Infof("\033[31m CRITICAL MAJOR update, please upgrade ASAP.\033[0m")
			}

			globals.Logger.Infof("%s", u.Message) // give the version upgrade message
			globals.Logger.Infof("\033[33mCurrent Version %s \033[32m-> Upgrade Version %s\033[0m ", current_version.String(), uversion.String())
		}
	}

	if u.Critical != "" { // give the critical upgrade message
		globals.Logger.Infof("%s", u.Critical)
	}

}
