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

// Very temporary
func main() {
	out := make(chan []byte)
	hub := NewConnHub(out)

	// ts := NewTerrariaServer(out, "/home/andrew/1405/Linux/TerrariaServer.exe")
	ts := NewTerrariaServer(out, "D:\\Games\\GOG\\Windows\\Terraria\\TerrariaServer.exe")
	// https://stackoverflow.com/questions/43601359/how-do-i-serve-css-and-js-in-go
	// Am thief. Credit to @RayfenWindspear :D

	go hub.Start()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		LogOutput(ts, "Received connection to /admin")
		t := template.Must(template.ParseFiles("templates/admin.html"))
		data := struct {
			Worldname   string
			Players     []Player
			PlayerCount int
			Password    string
			Seed        string
			Version     string
			MOTD        string
		}{
			Worldname:   "test",
			Players:     ts.Players(),
			PlayerCount: len(ts.Players()),
			Password:    ts.Password(),
			Seed:        ts.Seed(),
			Version:     ts.Version(),
			MOTD:        ts.MOTD(),
		}

		if err := t.Execute(w, data); err != nil {
			log.Output(1, err.Error())
			LogHTTP(ts, 500, r)
		}

		LogHTTP(ts, 200, r)
	})

	http.HandleFunc("/api/player/kick/", func(w http.ResponseWriter, r *http.Request) {
		LogInfo(ts, "Received kick request: "+r.RequestURI)
		pn := strings.TrimPrefix(r.RequestURI, "/api/player/kick/")
		rc := 403

		if plr := ts.Player(pn); plr != nil {
			plr.Kick("Kicked by the internet")
			rc = 200
		} else {
			rc = 404
		}

		w.WriteHeader(rc)
		LogHTTP(ts, rc, r)
	})

	http.HandleFunc("/api/player/ban/", func(w http.ResponseWriter, r *http.Request) {
		pn := strings.TrimPrefix(r.RequestURI, "/api/player/ban/")
		var (
			rc  = 403
			msg = "Banned from the internet"
		)

		if plr := ts.Player(pn); plr != nil {
			rc = 200
			plr.Ban(msg)
		} else {
			rc = 404
		}

		LogHTTP(ts, rc, r)
		w.WriteHeader(rc)
	})

	http.HandleFunc("/api/server/password/", func(w http.ResponseWriter, r *http.Request) {
		u, _ := url.Parse(r.RequestURI)
		p := strings.TrimPrefix(u.Path, "/api/server/password")
		p = strings.TrimPrefix(p, "/")

		if p == "" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(200)
			w.Write([]byte(ts.password))
			SendCommand("password "+p, ts)
		}

		LogHTTP(ts, 200, r)
	})

	http.HandleFunc("/api/server/start/", func(w http.ResponseWriter, r *http.Request) {
		if ts.IsUp() {
			LogHTTP(ts, 403, r)
			w.WriteHeader(403)
			return
		}

		if err := ts.Start(); err != nil {
			log.Fatal(err)
		}

		LogHTTP(ts, 200, r)
		w.WriteHeader(200)
	})

	http.HandleFunc("/api/server/stop/", func(w http.ResponseWriter, r *http.Request) {
		if ts.IsUp() {
			LogHTTP(ts, 200, r)
			go func() { ts.Stop() }()
			w.WriteHeader(200)
			return
		}

		LogHTTP(ts, 400, r)
		w.WriteHeader(400)
	})

	http.HandleFunc("/api/server/status/", func(w http.ResponseWriter, r *http.Request) {
		if ts.IsUp() {
			LogHTTP(ts, 200, r)
			w.WriteHeader(200)
		}

		LogHTTP(ts, 400, r)
		w.WriteHeader(400)
	})

	http.HandleFunc("/api/server/restart/", func(w http.ResponseWriter, r *http.Request) {
		if err := ts.Restart(); err != nil {
			LogHTTP(ts, 500, r)
			w.WriteHeader(500)
			log.Fatal(err)
		}

		LogHTTP(ts, 200, r)
		w.WriteHeader(200)
	})

	http.HandleFunc("/api/server/say/", func(w http.ResponseWriter, r *http.Request) {
		LogOutput(ts, "Sending message: "+r.RequestURI)
		u, _ := url.Parse(r.RequestURI)
		SendCommand("say "+strings.TrimPrefix(u.Path, "/api/server/say/"), ts)
		LogHTTP(ts, 200, r)
	})

	http.HandleFunc("/api/server/motd/", func(w http.ResponseWriter, r *http.Request) {
		u, _ := url.Parse(r.RequestURI)
		m := strings.TrimPrefix(u.Path, "/api/server/motd")
		m = strings.TrimPrefix(m, "/")

		w.WriteHeader(200)
		if m == "" {
			w.Write([]byte(ts.motd))
		} else {
			SendCommand("motd "+m, ts)
			SendCommand("motd", ts)
		}

		LogHTTP(ts, 200, r)
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
			LogHTTP(ts, 404, r)
			w.WriteHeader(404)
			return
		}

		if set != "" {
			SendCommand("say Setting time to "+set, ts)
			SendCommand(set, ts)
		}

		w.WriteHeader(200)
		LogHTTP(ts, 200, r)
	})

	http.HandleFunc("/api/server/settle/", func(w http.ResponseWriter, r *http.Request) {
		LogInfo(ts, "Settling liquids", out)
		SendCommand("settle", ts)
		w.WriteHeader(200)
		LogHTTP(ts, 200, r)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

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
