package main

import (
	"net"
	"regexp"
)

var gameServers []GameServer
var illegalNamesRe []*regexp.Regexp

// GameServer - A GameServer describes an interface to a full GameServer
type GameServer interface {
	Commandable
	Versioned
	Loggable
	Playable
	Server
}

// Playable - Define an object that can track the players that have joined
type Playable interface {
	Player(string) Player
	NewPlayer(string, string) Player
	RemovePlayer(string) bool
}

// Player - Define a player than can join a server and has various details
// regarding its connection tracked
type Player interface {
	Name() string
	SetIP(string)
	Kick(string)
	Ban(string)
	IP() net.IP
}

// Server -
type Server interface {
	IsUp() bool
	Stop() error
	Start() error
	Restart() error
}

// Versioned -
type Versioned interface {
	SetVersion(string)
	Version() string
}

// Commandable - A Commandable object must implement the function EnqueueCommand
type Commandable interface {
	EnqueueCommand(string)
}

// SendCommand - Send a command to a Commandable() object
func SendCommand(s string, cs Commandable) {
	cs.EnqueueCommand(s)
}

// RegisterIllegalName = Register a name/regex that is not permitted to be used.
func RegisterIllegalName(re string) {
	illegalNamesRe = append(illegalNamesRe, regexp.MustCompile(re))
}

// IsNameIllegal - Determine if a given name is not permitted to be used
func IsNameIllegal(s string) bool {
	for _, re := range illegalNamesRe {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func init() {
	// Prepare our application data
	gameServers = make([]GameServer, 0)
	illegalNamesRe = make([]*regexp.Regexp, 0)

	// Ban names that can mess with our Regex and confuse players
	RegisterIllegalName("^(\\s$|^[<>\\[\\]\\(\\)\\|]|[<>\\[\\]\\(\\)\\|]$)")
	RegisterIllegalName("^([aA]dmin|[sS]ystem|[sS]erver|[sS]uper[aA]dmin)")
}
