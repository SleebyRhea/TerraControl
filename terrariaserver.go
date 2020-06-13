package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Player - Defines a player that has connected to the server at some point
type Player struct {
	ip   net.IP
	name string
}

// IP - Returns the IP address that the player used to connect this session
func (p *Player) IP() net.IP {
	return p.ip
}

// Name - Return the name of the player object
func (p *Player) Name() string {
	return p.name
}

// TerrariaServer - Terraria server definition
type TerrariaServer struct {
	Cmd    *exec.Cmd
	Stdin  io.Writer
	Stdout io.Reader

	// Loggable
	loglevel int
	uuid     string

	// Commandable
	commandqueue    chan string
	commandcount    int
	commandqueuemax int

	players []*Player
}

// Start -
func (s *TerrariaServer) Start() error {
	var err error

	if s.Stdin, err = s.Cmd.StdinPipe(); err != nil {
		return err
	}

	if s.Stdout, err = s.Cmd.StdoutPipe(); err != nil {
		return err
	}

	s.commandqueue = make(chan string, 500)
	s.commandcount = 0
	s.commandqueuemax = 500

	ready := make(chan struct{})

	// Refactor these two goroutines to exit gracefully when the
	// server is stopped to avoid stale goroutines
	go superviseTerrariaOut(s, ready)
	go func() {
		for {
			select {
			case cmd := <-s.commandqueue:
				b := convertString(cmd)
				b.WriteTo(s.Stdin)
				s.commandcount = s.commandcount - 1
			}
		}
	}()

	if err = s.Cmd.Start(); err != nil {
		return err
	}

	<-ready
	return nil
}

// Stop -
func (s *TerrariaServer) Stop() error {
	LogOutput(s, "Stopping Terraria server")
	done := make(chan error)

	SendCommand("exit", s)
	go func() { done <- s.Cmd.Wait() }()

	LogDebug(s, "Waiting for Terraria to exit")
	select {
	case <-time.After(30 * time.Second):
		s.Cmd.Process.Kill()
		return errors.New("terraria took too long to exit, killed")
	case err := <-done:
		LogInfo(s, "Terraria server has been stopped")
		if err != nil {
			return err
		}
		return nil
	}
}

// Restart -
func (s *TerrariaServer) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}

	if err := s.Start(); err != nil {
		return err
	}

	return nil
}

// Status -
func (s *TerrariaServer) Status() (int, error) {
	if s.Cmd.ProcessState != nil {
		return 1, nil
	}

	if s.Cmd.Process != nil {
		return 0, nil
	}

	return 2, errors.New("Process entered an unknown state")
}

// IsUp -
func (s *TerrariaServer) IsUp() bool {
	if s.Cmd.ProcessState != nil {
		return false
	}

	if s.Cmd.Process != nil {
		return true
	}

	return false
}

/**********/
/* Logger */
/**********/

// UUID -
func (s *TerrariaServer) UUID() string {
	return s.uuid
}

// Loglevel -
func (s *TerrariaServer) Loglevel() int {
	return s.loglevel
}

// SetLoglevel -
func (s *TerrariaServer) SetLoglevel(l int) {
	s.loglevel = l
}

/***************/
/* Commandable */
/***************/

// EnqueueCommand -
func (s *TerrariaServer) EnqueueCommand(c string) {
	if s.commandcount < s.commandqueuemax-1 {
		s.commandqueue <- c + "\n"
		s.commandcount = s.commandcount + 1
	} else {
		LogWarning(s, "Attempted to run more than the maximum amount of commands!")
	}
}

/********/
/* Main */
/********/

// NewPlayer - Add a player to the list of players if it isn't already present
func (s *TerrariaServer) NewPlayer(n, ips string) *Player {
	var plr *Player
	if plr = s.Player(n); plr == nil {
		plr = &Player{name: n}
	}
	plr.ip = net.ParseIP(ips)
	s.players = append(s.players, plr)
	LogInfo(s, "New player logged: "+plr.Name())
	return plr
}

// RemovePlayer - Removes a player from the list of players
func (s *TerrariaServer) RemovePlayer(n string) bool {
	for i, p := range s.players {
		if p.Name() == n {
			LogInfo(s, "Removing "+p.Name())
			s.players = append(s.players[:i], s.players[i+1:]...)
			return true
		}
	}
	return false
}

// Player - Return a player object that matches the string given
func (s *TerrariaServer) Player(n string) *Player {
	for _, p := range s.players {
		if p.Name() == n {
			return p
		}
	}

	return nil
}

// NewTerrariaServer -
func NewTerrariaServer(path string, args ...string) *TerrariaServer {
	t := &TerrariaServer{
		uuid: "terraria",
		Cmd: exec.Command(path,
			"-autocreate", "3", "-world", "C:\\Users\\Andrew Wyatt\\Documents\\My Games\\Terraria\\Worlds\\World11.wld", "-secure",
			"-players", "8", "-pass", "123123", "-noupnp")}

	t.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	t.SetLoglevel(3)

	gameServers = append(gameServers, t)
	return t
}

func superviseTerrariaOut(s *TerrariaServer, ready chan struct{}) {
	LogDebug(s, "Started Terraria supervisor")
	logOut := LogInit
	scanner := bufio.NewScanner(s.Stdout)
	// messagerRe := regexp.MustCompile("^<([^<>])*>.*$")

	// Channel has a bufer of 100 in case too many connections are occurring at
	// at once.
	cch := make(chan string, 0)
	pch := make(chan string, 0)
	go connectionSupervisor(s, cch, pch)

	for scanner.Scan() {
		//Strip the prefix that terraria outputs on newline
		out := scanner.Text()
		out = strings.TrimPrefix(out, ": ")

		switch out {
		case "Server started":
			logOut(s, "Terraria server INIT completed")
			logOut = LogOutput
			ready <- struct{}{}

		default:
			switch e := EventType(out, s); e {
			case eventConnection:
				re := gameEvents[eventConnection]
				m := re.FindStringSubmatch(out)
				go func() {
					cch <- m[1]
					LogDebug(s, sprintf("Passed new connection information: %s", m[1]))
				}()

			case eventPlayerLogin:
				name := strings.TrimSuffix(out, " has joined.")
				SendCommand("playing", s)
				SendCommand(sprintf("say Hello there %s!", name), s)

			case eventPlayerLeft:
				name := strings.TrimSuffix(out, " has left.")
				s.RemovePlayer(name)

			case eventPlayerInfo:
				go func() {
					pch <- out
					LogDebug(s, sprintf("Passed player information: %s", out))
				}()
				m := gameEvents[eventPlayerInfo].FindStringSubmatch(out)
				s.NewPlayer(m[1], m[2])

			default:
				// Just log it and move on
				logOut(s, out)
			}
		}
	}
}

func connectionSupervisor(s *TerrariaServer, cch chan string, pch chan string) {
	newconnections := make(map[string]time.Time)
	stale := make(map[string]int)
	conRe := gameEvents[eventPlayerInfo]

	for {
		select {
		case c := <-cch:
			LogDebug(s, "Adding channeled connection to list")
			if _, ok := newconnections[c]; ok {
				LogWarning(s, "Stale connection found for IP: "+c)
				if num, ok := stale[c]; ok {
					stale[c] = num + 1
				} else {
					stale[c] = 1
				}
			}
			newconnections[c] = time.Now()
			//Timeout of 10 seconds on new connections
		case plr := <-pch:
			LogDebug(s, "Received player info: "+plr)
			m := conRe.FindStringSubmatch(plr)
			ip := m[2]
			name := m[1]
			if _, ok := newconnections[ip]; ok {
				delete(newconnections, ip)
				LogDebug(s, sprintf("Removed connection for IP: %s [%s]", ip, name))
			}

			if _, ok := stale[ip]; ok {
				delete(stale, ip)
				LogDebug(s, "Cleared stale connection count for IP: "+ip)
			}
		case <-time.After(5 * time.Second):
			if len(newconnections) < 1 {
				break
			}

			for ip, t := range newconnections {
				now := time.Now()
				if now.Sub(t) > 30*time.Second {
					LogWarning(s, "Stale connection found for IP: "+ip)
					delete(newconnections, ip)
					if num, ok := stale[ip]; ok {
						stale[ip] = num + 1
					} else {
						stale[ip] = 1
					}
				}
			}
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
