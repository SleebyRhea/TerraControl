package main

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
)

// TerrariaServer - Terraria server definition
type TerrariaServer struct {
	Cmd    *exec.Cmd
	Stdin  *bufio.Writer
	Stdout *bufio.Scanner

	loglevel int
	uuid     string

	commandqueue chan string
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

// Stop -
func (s *TerrariaServer) Stop() error {
	if _, err := s.Stdin.WriteString("\nexit\n"); err != nil {
		return err
	}

	if err := s.Cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// Start -
func (s *TerrariaServer) Start() error {
	var (
		stdin  io.WriteCloser
		stdout io.ReadCloser
		err    error
	)

	if stdin, err = s.Cmd.StdinPipe(); err != nil {
		return err
	}

	if stdout, err = s.Cmd.StdoutPipe(); err != nil {
		return err
	}

	s.Stdin = bufio.NewWriter(stdin)
	s.Stdout = bufio.NewScanner(stdout)

	if err = s.Start(); err != nil {
		return err
	}

	go superviseTerrariaOut(s)
	go superviseQueue(s)

	return nil
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

// CommandQueue -
func (s *TerrariaServer) CommandQueue() *chan string {
	return &s.commandqueue
}

// RunCommand -
func (s *TerrariaServer) RunCommand(c string) error {
	if _, err := s.Stdin.WriteString(c); err != nil {
		return err
	}
	return nil
}

// Main //

// NewTerrariaServer -
func NewTerrariaServer() *TerrariaServer {
	t := &TerrariaServer{
		uuid:         "terraria112312",
		loglevel:     3,
		commandqueue: make(chan string, 1000)}
	gameServers = append(gameServers, t)
	return t
}

func superviseTerrariaOut(t *TerrariaServer) {
	scanner := t.Stdout
	for scanner.Scan() {
		LogOutput(t, scanner.Text())
	}
}
