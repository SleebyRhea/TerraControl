package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

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
}

// Start -
func (s *TerrariaServer) Start() error {
	var err error

	s.commandqueue = make(chan string, 500)
	s.commandcount = 0
	s.commandqueuemax = 500

	if s.Stdin, err = s.Cmd.StdinPipe(); err != nil {
		return err
	}

	if s.Stdout, err = s.Cmd.StdoutPipe(); err != nil {
		return err
	}

	ready := make(chan struct{})

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
		LogDebug(s, "Terraria has exited")
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

// Main //

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
	r := false
	scanner := bufio.NewScanner(s.Stdout)
	// messagerRe := regexp.MustCompile("^<([^<>])*>.*$")

	for scanner.Scan() {
		//Strip the prefix that terraria outputs on newline
		out := scanner.Text()
		out = strings.TrimPrefix(out, ": ")

		switch out {
		case ": Server started", "Server started":
			r = true
			ready <- struct{}{}
			LogInit(s, "Terraria server INIT completed")
		default:
			if r {
				LogOutput(s, out)
			} else {
				LogInit(s, out)
			}

			switch e := EventType(out, s); e {
			case 0: //Player login
				name := strings.TrimSuffix(out, " has joined.")
				SendCommand("playing", s)
				SendCommand(sprintf("say Hello there %s!", name), s)
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
