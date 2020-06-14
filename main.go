package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
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
	SendCommand("password 123123", ts)
	SendCommand("motd Welcome!", ts)
	SendCommand("motd", ts)

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
			Password    string
			Seed        string
			Version     string
			MOTD        string
		}{
			Worldname:   "test",
			Players:     ts.Players(),
			PlayerCount: len(ts.Players()),
			Password:    ts.Password,
			Seed:        ts.Seed,
			Version:     ts.Version,
			MOTD:        ts.MOTD,
		}
		if err := t.Execute(w, data); err != nil {
			log.Output(1, err.Error())
		}
	})

	http.HandleFunc("/api/player/kick/", func(w http.ResponseWriter, r *http.Request) {
		LogInfo(ts, "Received kick request: "+r.RequestURI)
		pn := strings.TrimPrefix(r.RequestURI, "/api/player/kick/")

		if plr := ts.Player(pn); plr != nil {
			plr.Kick("Kicked by the internet")
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
		}
	})

	http.HandleFunc("/api/player/ban/", func(w http.ResponseWriter, r *http.Request) {
		pn := strings.TrimPrefix(r.RequestURI, "/api/player/ban/")
		LogInfo(ts, "Received ban request: "+r.RequestURI)

		if plr := ts.Player(pn); plr != nil {
			// plr.Ban("Banned from the internet")
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
		}
	})

	http.HandleFunc("/api/server/password/", func(w http.ResponseWriter, r *http.Request) {
		u, _ := url.Parse(r.RequestURI)
		p := strings.TrimPrefix(u.Path, "/api/server/password")
		p = strings.TrimPrefix(p, "/")
		LogInfo(ts, "Received password request: "+p)

		if p == "" {
			w.WriteHeader(200)
			w.Write([]byte(ts.Password))
		} else {
			SendCommand("password "+p, ts)
			w.WriteHeader(200)
		}
	})

	http.HandleFunc("/api/server/start", func(w http.ResponseWriter, r *http.Request) {
	})

	http.HandleFunc("/api/server/stop", func(w http.ResponseWriter, r *http.Request) {
	})

	http.HandleFunc("/api/server/say/", func(w http.ResponseWriter, r *http.Request) {
		LogOutput(ts, "Sending message: "+r.RequestURI)
		u, _ := url.Parse(r.RequestURI)
		SendCommand("say "+strings.TrimPrefix(u.Path, "/api/server/say/"), ts)
	})

	http.HandleFunc("/api/server/motd/", func(w http.ResponseWriter, r *http.Request) {
		u, _ := url.Parse(r.RequestURI)
		m := strings.TrimPrefix(u.Path, "/api/server/motd")
		m = strings.TrimPrefix(m, "/")
		LogInfo(ts, "Received motd request: "+m)

		if m == "" {
			w.Write([]byte(ts.MOTD))
		} else {
			SendCommand("motd "+m, ts)
			SendCommand("motd", ts)
		}
	})

	http.HandleFunc("/api/server/time/", func(w http.ResponseWriter, r *http.Request) {
		LogOutput(ts, "Received time request: "+r.RequestURI)
		u, _ := url.Parse(r.RequestURI)
		t := strings.TrimPrefix(u.Path, "/api/server/time")
		set := ""
		switch t {
		case "/", "":
			SendCommand("time", ts)
			return
		case "/dawn":
			set = "dawn"
		case "/noon":
			set = "noon"
		case "/dusk":
			set = "dusk"
		case "/midnight":
			set = "midnight"
		default:
			LogDebug(ts, "Invalid time sent: "+t)
			w.WriteHeader(404)
			return
		}

		if set != "" {
			SendCommand("say Setting time to "+set, ts)
			SendCommand(set, ts)
		}
	})

	http.HandleFunc("/api/server/settle", func(w http.ResponseWriter, r *http.Request) {
		LogInfo(ts, "Settling liquids")
		SendCommand("settle", ts)
		w.WriteHeader(200)
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
