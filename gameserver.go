package main

import "regexp"

var gameServers = make([]GameServer, 0)
var gameEvents = make([]*regexp.Regexp, 0)
var illegalNamesRe = make([]*regexp.Regexp, 0)

const (
	eventConnection = 0
	eventPlayerJoin = 1
	eventPlayerLeft = 2
	eventPlayerInfo = 3
	eventPlayerChat = 4
	eventPlayerBoot = 5
	eventPlayerBan  = 6

	eventServerTime = 7
	eventServerSeed = 8
	eventServerMOTD = 9
	eventServerPass = 10
	eventServerVers = 11
)

// GameServer -
type GameServer interface {
	IsUp() bool
	Stop() error
	Start() error
	Restart() error
}

// Commandable -
type Commandable interface {
	EnqueueCommand(string)
}

// SendCommand -
func SendCommand(s string, cs Commandable) {
	cs.EnqueueCommand(s)
}

// EventType -
func EventType(s string, gs GameServer) int {
	for i, re := range gameEvents {
		if re.MatchString(s) {
			return i
		}
	}
	return -1
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
	ipReString := "[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}"
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^("+ipReString+"):[0-9]{1,5} is connecting...$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^(.{1,20}) has joined.$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^(.{1,20}) has left.$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^(.{1,20}) \\(("+ipReString+"):[0-9]{1,5}\\)$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^<(.{1,20})> (.*)$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^("+ipReString+"):[0-9]{1,5} was booted: (.*)$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^("+ipReString+"):[0-9]{1,5} was banned: (.*)$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^Time: (.?:..)([AP]M)$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^World Seed: (.*)$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^MOTD: (.*)$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^Password: (.*)$"))
	gameEvents = append(gameEvents, regexp.MustCompile(
		"^Terraria Server v(.*)$"))
	illegalNamesRe = append(illegalNamesRe, regexp.MustCompile(
		"^(\\s$|^[<>\\[\\]\\(\\)\\|\\]|[<>\\[\\]\\(\\)\\|\\]$|[aA]dmin|[sS]ystem|[sS]erver|[sS]uper[aA]dmin)"))
}
