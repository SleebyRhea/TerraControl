package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Very temporary
func main() {
	out := make(chan []byte)
	hub := NewConnHub(out)

	ts := NewTerrariaServer(out, "D:\\Games\\GOG\\Windows\\Terraria\\TerrariaServer.exe")

	go hub.Start()

	serveHTTP(hub, ts, out)

	go func() {
		log.Output(1, "Starting webserver")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	if err := ts.Start(); err != nil {
		log.Output(1, err.Error())
		os.Exit(1)
	}

	log.Output(1, "Completed INIT. Waiting for termination signal")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc)

	for sig := range sc {
		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			fmt.Print("\r")
			log.Output(1, "Quitting")
			if ts.IsUp() {
				if err := ts.Stop(); err != nil {
					log.Fatal(err)
				}
			}
			os.Exit(0)
		default:
			log.Output(1, "Caught signal "+sig.String())
		}
	}
}

func convertString(str string) bytes.Buffer {
	b := *bytes.NewBuffer(make([]byte, 0))
	nul := []byte{0x0000}
	for _, c := range str {
		b.WriteRune(c)
		b.Write(nul)
	}
	log.Output(1, sprintf("[DEBUG] Converted string %q to [% x] ", str, b.Bytes()))
	return b
}
