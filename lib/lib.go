package lib

import (
	"context"
	"io"
	"log"

	"github.com/gorilla/websocket"
)

func File2WS(ctx context.Context, cancel func(), src io.Reader, dst *websocket.Conn) error {
	defer cancel()
	for {
		if ctx.Err() != nil {
			return nil
		}
		b := make([]byte, 32*1024)
		if n, err := src.Read(b); err != nil {
			return err
		} else {
			b = b[:n]
		}
		//log.Printf("->ws %d bytes: %q", len(b), string(b))
		if err := dst.WriteMessage(websocket.BinaryMessage, b); err != nil {
			log.Println(err)
			return err
		}
	}
}
