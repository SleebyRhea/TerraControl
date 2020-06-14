package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Player - Defines a player that has connected to the server at some point
type Player struct {
	ip     net.IP
	name   string
	server *TerrariaServer
}

// IP - Returns the IP address that the player used to connect this session
func (p *Player) IP() net.IP {
	return p.ip
}

// Name - Return the name of the player object
func (p *Player) Name() string {
	return p.name
}

// Kick - Kick a player
func (p *Player) Kick(r string) {
	SendCommand(sprintf("say Kicking player: \"%s\". %s.", p.Name(), r), p.server)
	SendCommand("kick "+p.Name(), p.server)
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

	// PlayerInfo
	players  []*Player
	messages [][2]string

	// Game Info
	worldfile  string
	configfile string
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
				time.Sleep(time.Second / 2)
				b := convertString(cmd)
				b.WriteTo(s.Stdin)
				LogDebug(s, "Ran: "+cmd)
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

// Player - Return a player object that matches the string given
func (s *TerrariaServer) Player(n string) *Player {
	for _, p := range s.players {
		if p.Name() == n {
			return p
		}
	}

	return nil
}

// Players - Returns the players that are currently in-game
func (s *TerrariaServer) Players() []*Player {
	return s.players
}

// NewPlayer - Add a player to the list of players if it isn't already present
func (s *TerrariaServer) NewPlayer(n, ips string) *Player {
	var plr *Player
	if plr = s.Player(n); plr == nil {
		plr = &Player{name: n, server: s}
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

// ChatMessages - Return the total number of message that are logged
func (s *TerrariaServer) ChatMessages() [][2]string {
	return s.messages
}

// NewChatMessage - Return the total number of message that are logged
func (s *TerrariaServer) NewChatMessage(msg, name string) {
	s.messages = append(s.messages, [2]string{name, msg})
}

// Time - Returns the current game time
// TODO: Actually finish this.
func (s *TerrariaServer) Time() string {
	SendCommand("time", s)
	return ""
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

/**************/
/* Goroutines */
/**************/

func superviseTerrariaOut(s *TerrariaServer, ready chan struct{}) {
	LogDebug(s, "Started Terraria supervisor")
	logOut := LogInit
	scanner := bufio.NewScanner(s.Stdout)
	cch := make(chan string, 0)
	pch := make(chan string, 0)

	go superviseTerrariaConnects(s, cch, pch)

	for scanner.Scan() {
		// Strip the prefix that terraria outputs on a newline
		out := scanner.Text()
		out = strings.TrimPrefix(out, ":")
		out = strings.TrimPrefix(out, " ")

		switch out {
		case "Server started":
			logOut(s, "Terraria server INIT completed")
			logOut = LogOutput
			ready <- struct{}{}

		default:
			switch e := EventType(out, s); e {
			case eventConnection:
				re := gameEvents[e]
				m := re.FindStringSubmatch(out)
				go func() {
					cch <- m[1]
					LogDebug(s, sprintf("Passed new connection information: %s", m[1]))
				}()

			case eventPlayerJoin:
				SendCommand("playing", s)

			case eventPlayerLeft:
				name := strings.TrimSuffix(out, " has left.")
				s.RemovePlayer(name)

			case eventPlayerInfo:
				go func() { pch <- out }()
				m := gameEvents[e].FindStringSubmatch(out)
				plr := s.NewPlayer(m[1], m[2])
				if IsNameIllegal(plr.Name()) {
					plr.Kick("Name is not allowed")
				}

			case eventPlayerChat:
				m := gameEvents[e].FindStringSubmatch(out)
				s.NewChatMessage(m[2], m[1])
				logOut(s, out)

			case eventPlayerBoot:
				m := gameEvents[e].FindStringSubmatch(out)
				LogInfo(s, sprintf("Failed connection: %s [%s]", m[1], m[2]))

			default:
				// Just log it and move on
				logOut(s, out)
			}
		}
	}
}

func superviseTerrariaConnects(s *TerrariaServer, cch chan string, pch chan string) {
	newconnections := make(map[string]time.Time)
	stale := make(map[string]int)
	conRe := gameEvents[eventPlayerInfo]

	for {
		select {
		case <-time.After(5 * time.Second):
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

			for ip, cnt := range stale {
				if cnt > 25 {
					LogWarning(s, "Possible DoS taking place!")
					LogWarning(s, sprintf("IP: %s | Stale Connections: %d", ip, cnt))
				}
				delete(stale, ip)
			}

		case c := <-cch:
			LogDebug(s, "Adding channeled connection to list")
			if _, ok := newconnections[c]; ok {
				LogWarning(s, "Extra connection found for IP: "+c)
				if num, ok := stale[c]; ok {
					stale[c] = num + 1
				} else {
					stale[c] = 1
				}

			}
			newconnections[c] = time.Now()

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
		}
	}
}
