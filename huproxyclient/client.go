// Copyright 2017-2022 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aus/proxyplease"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	huproxy "github.com/google/huproxy/lib"
)

var (
	writeTimeout = flag.Duration("write_timeout", 10*time.Second, "Write timeout")
	basicAuth    = flag.String("auth", "", "HTTP Basic Auth in @<filename> or <username>:<password> format.")
	certFile     = flag.String("cert", "", "Certificate Auth File")
	keyFile      = flag.String("key", "", "Certificate Key File")
	verbose      = flag.Bool("verbose", false, "Verbose.")
	insecure     = flag.Bool("insecure_conn", false, "Skip certificate validation")
)

func secretString(s string) (string, error) {
	if strings.HasPrefix(s, "@") {
		fn := s[1:]
		st, err := os.Stat(fn)
		if err != nil {
			return "", err
		}
		p := st.Mode() & os.ModePerm
		if p&0177 > 0 {
			return "", fmt.Errorf("valid permissions for %q is %0o, was %0o", fn, 0600, p)
		}
		b, err := ioutil.ReadFile(fn)
		return strings.TrimSpace(string(b)), err
	}
	return s, nil
}

func dialError(url string, resp *http.Response, err error) {
	if resp != nil {
		extra := ""
		if *verbose {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Warningf("Failed to read HTTP body: %v", err)
			}
			extra = "Body:\n" + string(b)
		}
		log.Fatalf("%s: HTTP error: %d %s\n%s", err, resp.StatusCode, resp.Status, extra)

	}
	log.Fatalf("Dial to %q fail: %v", url, err)
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatalf("Want exactly one arg")
	}
	url := flag.Arg(0)

	if *verbose {
		log.Infof("huproxyclient %s", huproxy.Version)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dialContext := proxyplease.NewDialContext(proxyplease.Proxy{})

	dialer := websocket.Dialer{
		// It's not documented if handshake timeout defaults.
		HandshakeTimeout: websocket.DefaultDialer.HandshakeTimeout,
		NetDialContext:   dialContext,
	}

	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: *insecure,
	}
	head := map[string][]string{}

	// Add basic auth.
	if *basicAuth != "" {
		ss, err := secretString(*basicAuth)
		if err != nil {
			log.Fatalf("Error reading secret string %q: %v", *basicAuth, err)
		}
		a := base64.StdEncoding.EncodeToString([]byte(ss))
		head["Authorization"] = []string{
			"Basic " + a,
		}
	}

	// Load client cert
	if *certFile != "" {
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Fatal(err)
		}

		dialer.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	conn, resp, err := dialer.Dial(url, head)
	if err != nil {
		dialError(url, resp, err)
	}
	defer conn.Close()

	// websocket -> stdout
	go func() {
		for {
			mt, r, err := conn.NextReader()
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			if mt != websocket.BinaryMessage {
				log.Fatal("non-binary websocket message received")
			}
			if _, err := io.Copy(os.Stdout, r); err != nil {
				log.Errorf("Reading from websocket: %v", err)
				cancel()
			}
		}
	}()

	// stdin -> websocket
	// TODO: NextWriter() seems to be broken.
	if err := huproxy.File2WS(ctx, cancel, os.Stdin, conn); err == io.EOF {
		if err := conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(*writeTimeout)); err == websocket.ErrCloseSent {
		} else if err != nil {
			log.Errorf("Error sending 'close' message: %v", err)
		}
	} else if err != nil {
		log.Errorf("reading from stdin: %v", err)
		cancel()
	}

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
