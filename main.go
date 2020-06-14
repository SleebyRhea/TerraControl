package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	ts := NewTerrariaServer("/home/andrew/1405/Windows/TerrariaServer.exe")

	if err := ts.Start(); err != nil {
		log.Output(1, err.Error())
		os.Exit(1)
	}

	SendCommand("help", ts)

	// https://stackoverflow.com/questions/43601359/how-do-i-serve-css-and-js-in-go
	// Am thief. Credit to @RayfenWindspear :D
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		LogOutput(ts, "Received connection to /admin")
		t := template.Must(template.ParseFiles("templates/admin.html"))
		data := struct {
			Worldname   string
			Players     []*Player
			PlayerCount int
		}{
			Worldname:   "test",
			Players:     ts.Players(),
			PlayerCount: len(ts.Players())}
		if err := t.Execute(w, data); err != nil {
			log.Output(1, err.Error())
		}
	})

	http.HandleFunc("/api/player/kick/", func(w http.ResponseWriter, r *http.Request) {
		LogOutput(ts, "Received kick request: "+r.RequestURI)
		pn := strings.TrimPrefix(r.RequestURI, "/api/player/kick/")
		if plr := ts.Player(pn); plr != nil {
			plr.Kick("Kicked by the internet")
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
		}
	})

	http.HandleFunc("/api/server/say/", func(w http.ResponseWriter, r *http.Request) {
		LogOutput(ts, "Received kick request: "+r.RequestURI)
		pn := strings.TrimPrefix(r.RequestURI, "/api/server/say/")
		SendCommand("say "+pn, ts)
	})

	go func() { log.Fatal(http.ListenAndServe(":8080", nil)) }()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc)

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
