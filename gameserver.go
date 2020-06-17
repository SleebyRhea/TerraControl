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
	LoginMessager
	PasswordLockable
	Seeded
	Websocketer
}

// OutputSender sends output from a GameServer to a channel
type OutputSender interface {
	SetSendChannel(chan []byte)
}

// Playable - Define an object that can track the players that have joined
type Playable interface {
	Player(string) Player
	Players() []Player
	NewPlayer(string, string) Player
	RemovePlayer(string) bool
}

// Player - Define a player than can join a server and has various details
// regarding its connection tracked
type Player interface {
	SetIP(string)
	Name() string
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

// Websocketer is an object that is able to output to the Guis websocket ub
type Websocketer interface {
	WSOutput() chan []byte
}

// Versioned is an interface to objects with verisons
type Versioned interface {
	SetVersion(string)
	Version() string
}

// PasswordLockable is an interface to an object that can have its password set
type PasswordLockable interface {
	Password() string
	SetPassword(string)
}

// LoginMessager is an interface to an object that can have an MOTD set
type LoginMessager interface {
	MOTD() string
	SetMOTD(string)
}

// Seeded is an interface to an object that has a seed
type Seeded interface {
	Seed() string
	SetSeed(string)
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
