package parser

import (
	"os"

	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
)

func Start() {
	filePath := "./demofiles/demo.dem"
	iFile, err := os.Open(filePath)
	checkError(err)

	iParser := dem.NewParser(iFile)
	defer iParser.Close()

	var attackTickMap map[int][]events.WeaponFire = make(map[int][]events.WeaponFire)
	var jumpTickMap map[int][]uint64 = make(map[int][]uint64)
	var (
		roundStarted      = 0
		roundInFreezetime = 0
		roundNum          = 0
	)

	iParser.RegisterEventHandler(func(e events.FrameDone) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()

		if roundInFreezetime == 0 {
			tPlayers := gs.TeamTerrorists().Members()
			ctPlayers := gs.TeamCounterTerrorists().Members()
			Players := append(tPlayers, ctPlayers...)
			for _, player := range Players {
				if player != nil {
					var addonButton int32 = 0
					if attackEvent, ok := attackTickMap[currentTick]; ok {
						for _, atEvent := range attackEvent {
							if atEvent.Shooter.SteamID64 == player.SteamID64 {
								addonButton |= IN_ATTACK
								break
							}
						}
					}
					if jumpList, ok := jumpTickMap[currentTick]; ok {
						for _, steamid := range jumpList {
							if steamid == player.SteamID64 {
								addonButton |= IN_JUMP
								break
							}
						}
					}
					parsePlayerFrame(player, addonButton)
				}
			}
			delete(attackTickMap, currentTick)
			delete(jumpTickMap, currentTick)
		}
	})

	iParser.RegisterEventHandler(func(e events.WeaponFire) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		attackTickMap[currentTick] = append(attackTickMap[currentTick], e)
	})

	iParser.RegisterEventHandler(func(e events.PlayerJump) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		jumpTickMap[currentTick] = append(jumpTickMap[currentTick], e.Player.SteamID64)
	})

	iParser.RegisterEventHandler(func(e events.RoundStart) {
		roundStarted = 1
		roundInFreezetime = 1
	})

	iParser.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		roundInFreezetime = 0
		roundNum += 1
		gs := iParser.GameState()
		tPlayers := gs.TeamTerrorists().Members()
		ctPlayers := gs.TeamCounterTerrorists().Members()
		Players := append(tPlayers, ctPlayers...)
		for _, player := range Players {
			if player != nil {
				// parse player
				parsePlayerInitFrame(player)
			}
		}
	})

	iParser.RegisterEventHandler(func(e events.RoundEnd) {
		if roundStarted == 0 {
			roundStarted = 1
			roundNum = 0
		}
		gs := iParser.GameState()
		tPlayers := gs.TeamTerrorists().Members()
		ctPlayers := gs.TeamCounterTerrorists().Members()
		Players := append(tPlayers, ctPlayers...)
		for _, player := range Players {
			if player != nil {
				// save to rec file
				saveToRecFile(player, int32(roundNum))
			}
		}
	})
	err = iParser.ParseToEnd()
	checkError(err)
}
