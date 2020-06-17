package main

import "strings"

func init() {
	ipReString := "[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}"

	RegisterGameEventHandler("EventConnection",
		"^("+ipReString+"):[0-9]{1,5} is connecting...$",
		handleEventConnection)
	RegisterGameEventHandler("EventPlayerJoin",
		"^(.{1,20}) has joined.$",
		handleEventPlayerJoin)
	RegisterGameEventHandler("EventPlayerLeft",
		"^(.{1,20}) has left.$",
		handleEventPlayerLeft)
	RegisterGameEventHandler("EventPlayerInfo",
		"^(.{1,20}) \\(("+ipReString+"):[0-9]{1,5}\\)$",
		handleEventPlayerInfo)
	RegisterGameEventHandler("EventPlayerChat",
		"^<(.{1,20})> (.*)$",
		handleEventPlayerChat)
	RegisterGameEventHandler("EventPlayerBoot",
		"^("+ipReString+"):[0-9]{1,5} was booted: (.*)$",
		handleEventPlayerBoot)
	RegisterGameEventHandler("EventPlayerBan",
		"^("+ipReString+"):[0-9]{1,5} was banned: (.*)$",
		handleEventPlayerBan)
	RegisterGameEventHandler("EventServerTime",
		"^Time: (.?:..)([AP]M)$",
		handleEventServerTime)
	RegisterGameEventHandler("EventServerSeed",
		"^World Seed: (.*)$",
		handleEventServerSeed)
	RegisterGameEventHandler("EventServerMOTD",
		"^MOTD: (.*)$",
		handleEventServerMOTD)
	RegisterGameEventHandler("EventServerPass",
		"^Password: (.*)$",
		handleEventServerPass)
	RegisterGameEventHandler("EventServerVers",
		"^Terraria Server v(.*)$",
		handleEventServerVers)
	RegisterGameEventHandler("EventNone",
		".*",
		defaultEventHandler)
}

func handleEventConnection(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	m := e.Capture.FindStringSubmatch(in)
	go func() { oc <- m[1] }()
}

func handleEventPlayerJoin(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	SendCommand("playing", gs)
	LogInfo(gs, in, gs.WSOutput())
}

func handleEventPlayerLeft(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	name := strings.TrimSuffix(in, " has left.")
	gs.RemovePlayer(name)
}

func handleEventPlayerInfo(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	go func() { oc <- in }()
	m := e.Capture.FindStringSubmatch(in)
	plr := gs.NewPlayer(m[1], m[2])
	if IsNameIllegal(plr.Name()) {
		plr.Kick("Name is not allowed")
	}
}

func handleEventPlayerChat(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	LogChat(gs, in, gs.WSOutput())
}

func handleEventPlayerBoot(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	m := e.Capture.FindStringSubmatch(in)
	LogInfo(gs, sprintf("Failed connection: %s [%s]", m[1], m[2]), gs.WSOutput())
}

func handleEventPlayerBan(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	LogInfo(gs, in, gs.WSOutput())
}

func handleEventServerTime(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	LogOutput(gs, in)
}

func handleEventServerSeed(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	m := e.Capture.FindStringSubmatch(in)
	gs.SetSeed(m[1])
}

func handleEventServerMOTD(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	m := e.Capture.FindStringSubmatch(in)
	gs.SetMOTD(m[1])
}

func handleEventServerPass(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	m := e.Capture.FindStringSubmatch(in)
	gs.SetPassword(m[1])
}

func handleEventServerVers(gs GameServer, e *GameEvent, in string,
	oc chan string) {
	m := e.Capture.FindStringSubmatch(in)
	gs.SetVersion(m[1])
}
