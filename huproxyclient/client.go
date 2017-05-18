package main

import (
	"context"
	"io/ioutil"
	"flag"
	"io"
	"log"
	"os"
	"time"
	"net/http"

	"github.com/gorilla/websocket"

	huproxy "github.com/google/huproxy/lib"
)

var (
	writeTimeout = flag.Duration("write_timeout", 10*time.Second, "Write timeout")
	verbose = flag.Bool("verbose", false, "Verbose.")
)


func dialError(url string, resp *http.Response, err error) {
	if resp != nil {
		extra := ""
		if *verbose {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read HTTP body: %v", err)
			}
			extra = "Body:\n" + string(b)
		}
		log.Fatalf("%s: HTTP error: %d %s\n%s", err,resp.StatusCode, resp.Status,extra)
		
	}
	log.Fatalf("Dial to %q fail: %v", url, err)
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatalf("Want exactly one arg")
	}
	url := flag.Arg(0)

	// log.Printf("huproxyclient")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dialer := websocket.Dialer{}
	head := map[string][]string{}

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
				log.Fatal("blah")
			}
			if _, err := io.Copy(os.Stdout, r); err != nil {
				log.Printf("Reading from websocket: %v", err)
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
				log.Printf("Error sending close message: %v", err)
			}
	} else if err != nil {
		log.Printf("reading from stdin: %v", err)
		cancel()
	}

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
