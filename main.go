package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ts := NewTerrariaServer("/home/andrew/1405/Windows/TerrariaServer.exe")

	if err := ts.Start(); err != nil {
		log.Output(1, err.Error())
		os.Exit(1)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc)
	SendCommand("help", ts)

	go func() {
		for {
			SendCommand("say Testing", ts)
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			<-time.After(10 * time.Second)
			LogInfo(ts, sprintf("Message logged: %d", len(ts.ChatMessages())))
		}
	}()

	for sig := range sc {
		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			fmt.Print("\r")
			log.Output(1, "Quitting")
			if err := ts.Stop(); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		default:
			log.Output(1, "Caught signal "+sig.String())
		}
	}
}
