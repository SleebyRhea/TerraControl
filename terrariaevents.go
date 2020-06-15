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

func handleEventConnection(gs GameServer, in string, oc chan string) {
	ge := gameEventsMap["EventConnection"]
	m := ge.Capture.FindStringSubmatch(in)
	go func() {
		oc <- m[1]
		LogDebug(gs, sprintf("Passed new connection information: %s", m[1]))
	}()
}

func handleEventPlayerJoin(gs GameServer, in string, oc chan string) {
	SendCommand("playing", gs)
}

func handleEventPlayerLeft(gs GameServer, in string, oc chan string) {
	name := strings.TrimSuffix(in, " has left.")
	gs.RemovePlayer(name)
}

func handleEventPlayerInfo(gs GameServer, in string, oc chan string) {
	go func() { oc <- in }()
	m := gameEventsMap["EventPlayerInfo"].Capture.FindStringSubmatch(in)
	plr := gs.NewPlayer(m[1], m[2])
	if IsNameIllegal(plr.Name()) {
		plr.Kick("Name is not allowed")
	}
}

func handleEventPlayerChat(gs GameServer, in string, oc chan string) {
	logChat(gs, in)
}

func handleEventPlayerBoot(gs GameServer, in string, oc chan string) {
	m := gameEventsMap["EventPlayerBoot"].Capture.FindStringSubmatch(in)
	LogInfo(gs, sprintf("Failed connection: %s [%s]", m[1], m[2]))
}

func handleEventPlayerBan(gs GameServer, in string, oc chan string) {
	LogOutput(gs, in)
	// m := gameEvents[e].FindStringSubmatch(out)
	// LogInfo(s, sprintf("Banned IP: %s [%s]", m[1], m[2]))
}

func handleEventServerTime(gs GameServer, in string, oc chan string) {
	LogOutput(gs, in)
}

func handleEventServerSeed(gs GameServer, in string, oc chan string) {
	// m := gameEventsMap["EventServerSeed"].Capture.FindStringSubmatch(in)
	// gs.Seed = m[1]
}

func handleEventServerMOTD(gs GameServer, in string, oc chan string) {
	// m := gameEventsMap["EventServerMOTD"].Capture.FindStringSubmatch(in)
	// s.MOTD = m[1]
}

func handleEventServerPass(gs GameServer, in string, oc chan string) {
	//
	// s.Password = m[1]
}

func handleEventServerVers(gs GameServer, in string, oc chan string) {
	m := gameEventsMap["EventServerVers"].Capture.FindStringSubmatch(in)
	gs.SetVersion(m[1])
}
