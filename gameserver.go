package main

import "time"

var gameServers = make([]GameServer, 0)

// GameServer -
type GameServer interface {
	IsUp() bool
	Stop() error
	Start() error
	Restart() error

	//Save() error
	//RunCommand(...string) error
}

// Commandable -
type Commandable interface {
	EnqueueCommand(string)
	RunCommand(string) error
	CommandQueue() *chan string
}

func superviseQueue(c Commandable) {
	q := c.CommandQueue()
	for {
		select {
		case command := <-*q:
			c.RunCommand(command)
			time.Sleep(time.Second)
		}
	}
}
