package parser

import (
	"fmt"
	"os"
	"math"
	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	encoder "github.com/hx-w/minidemo-encoder/internal/encoder"
	common "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
)

func Start(fileName string) {
	filePath := "./demofiles/" + fileName
	iFile, err := os.Open(filePath)
	checkError(err)

	iParser := dem.NewParser(iFile)
	defer iParser.Close()

	iParser.ParseHeader()

	iHeader := iParser.Header()
	encoder.InitDemo(fileName, iHeader.MapName)
	encoder.SvTickRate = int32(math.Round((&iHeader).FrameRate()))
	demoLength := iHeader.PlaybackTicks

	var (
		roundNum 						int32 = 0
		roundNumLast 				int32 = 0
		roundValidPlayers 	[]*common.Player
		roundCorrupted 			bool = false
		roundStartedTick 		int = -1
	)

	// Tick Maps

	// Used to record the WeaponAttack event under a certain Tick and handle it in FrameDone
	var attackTickMap map[int][]events.WeaponFire = make(map[int][]events.WeaponFire)
	iParser.RegisterEventHandler(func(e events.WeaponFire) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		attackTickMap[currentTick] = append(attackTickMap[currentTick], e)
	})

	var jumpTickMap map[int][]uint64 = make(map[int][]uint64)
	iParser.RegisterEventHandler(func(e events.PlayerJump) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		jumpTickMap[currentTick] = append(jumpTickMap[currentTick], e.Player.SteamID64)
	})

	var itemDropTickMap map[int][]events.ItemDrop = make(map[int][]events.ItemDrop)
	iParser.RegisterEventHandler(func(e events.ItemDrop) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		itemDropTickMap[currentTick] = append(itemDropTickMap[currentTick], e)
	})

	var playerHurtTickMap map[int][]events.PlayerHurt = make(map[int][]events.PlayerHurt)
	iParser.RegisterEventHandler(func(e events.PlayerHurt) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		playerHurtTickMap[currentTick] = append(playerHurtTickMap[currentTick], e)
	})

	// iParser.RegisterEventHandler(func(e events.weaponReload) {
	// 	gs := iParser.GameState()
	// 	currentTick := gs.IngameTick()
	// 	// e.Player
	// })

	// var playerFallDamageTickMap map[int][]events.playerFallDamage = make(map[int][]events.playerFallDamage)
	// iParser.RegisterEventHandler(func(e events.playerFallDamage) {
	// 	gs := iParser.GameState()
	// 	currentTick := gs.IngameTick()
	// 	playerFallDamageTickMap[currentTick] = append(playerFallDamageTickMap[currentTick], e)
	// })

	var grenadeThrownTickMap map[int][]*common.GrenadeProjectile = make(map[int][]*common.GrenadeProjectile)
	iParser.RegisterEventHandler(func(e events.GrenadeProjectileThrow) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		if e.Projectile.Thrower != nil {
			grenadeThrownTickMap[currentTick] = append(grenadeThrownTickMap[currentTick], e.Projectile)
		}
	})

	// The preparation time is over and officially started
	// put it on roundStart, not freezetime, its saving its freezetime location, not sure, read above
	// RoundFreezetimeEnd
	// RoundStart
	iParser.RegisterEventHandler(func(e events.RoundStart) { //RoundFreezetimeEnd
		// Save on next tick
			gs := iParser.GameState()
			currentTick := gs.IngameTick()
			roundStartedTick = currentTick
			// tPlayers := gs.TeamTerrorists().Members()
			// ctPlayers := gs.TeamCounterTerrorists().Members()
			// roundValidPlayers = append(tPlayers, ctPlayers...)
			// fmt.Println("saving players tick: ", currentTick, roundValidPlayers)
			// for _, player := range roundValidPlayers {
			// 	if player != nil {
			// 		// parse player
			// 		parsePlayerInitFrame(player)
			// 	}
			// }
	})

	// iParser.RegisterEventHandler(func(e events.RoundEnd) {
	// 	gs := iParser.GameState()
	// 	tScore := gs.TeamTerrorists().Score()
	// 	ctScore := gs.TeamCounterTerrorists().Score()
	// 	fmt.Println("round ended")
	// })

	// End of round, excluding free time
	iParser.RegisterEventHandler(func(e events.RoundEndOfficial) {
		gs := iParser.GameState()
		tScore := gs.TeamTerrorists().Score()
		ctScore := gs.TeamCounterTerrorists().Score()
		roundNum = int32(ctScore+tScore)
		if roundCorrupted || (roundNum == 0 && roundNumLast == 0) {
			// more than 10 round difference jump, this demo is f up, only consider this round
			fmt.Println("ignoring round: CT", ctScore, " - T", tScore)
		} else {
				if (math.Abs(float64(roundNum-roundNumLast))>10) {
					// more than 10 round difference jump, this demo is f up, only consider this round
					roundCorrupted = true
					// consider it as last round, give the winner a +
					roundNum = roundNumLast + 1
					// TODO: find the winner team?
				}
				for _, player := range roundValidPlayers {
					if player != nil {
						// save to rec file
						saveToRecFile(player, roundNum)
					}
				}
				fmt.Println("saved round", roundNum," round result: CT", ctScore, " - T", tScore)
		}
		roundNumLast = roundNum
	})

	i := iParser.GameState().IngameTick()

	// ///////////jump///////////
	// for i < 147400-100 {
	// 	iParser.ParseNextFrame()
	// 	i = iParser.GameState().IngameTick()
	// }
	// ///////////jump///////////

	for i < demoLength {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		tPlayers := gs.TeamTerrorists().Members()
		ctPlayers := gs.TeamCounterTerrorists().Members()
		Players := append(tPlayers, ctPlayers...)
		if ((roundStartedTick > 0) && (currentTick > roundStartedTick)) {
			roundValidPlayers = Players
			fmt.Println("saving players: ", roundValidPlayers)
			for _, player := range roundValidPlayers {
				if player != nil {
					// parse player
					parsePlayerInitFrame(player)
				}
			}
			roundStartedTick = -1
		}
		for _, player := range Players {
			if player != nil {
				// Parse all the events to buttons
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

				if dropList, ok := itemDropTickMap[currentTick]; ok {
					for _, eventItemDrop := range dropList {
						if eventItemDrop.Player.SteamID64 == player.SteamID64 {
							addonButton |= IN_SCORE
							break
						}
					}
				}

				if !player.IsAlive() {
					continue
				}

				iFrameInfo := new(encoder.FrameInfo)
				iFrameInfo.ExtraData = 0

				if playerHurtList, ok := playerHurtTickMap[currentTick]; ok {
					for _, EventPlayerHurt := range playerHurtList {
						if EventPlayerHurt.Player.SteamID64 == player.SteamID64 {
							iFrameInfo.Health = int32(EventPlayerHurt.Health)
							iFrameInfo.ExtraData |= encoder.EXTRA_PLAYERDATA_HEALTH
							iFrameInfo.Armor = int32(EventPlayerHurt.Armor)
							iFrameInfo.ExtraData |= encoder.EXTRA_PLAYERDATA_ARMOR
							break
						}
					}
				}

				if grenadeThrownList, ok := grenadeThrownTickMap[currentTick]; ok {
					for _, grenadeProjectile := range grenadeThrownList {
						if grenadeProjectile.Thrower.SteamID64 == player.SteamID64 {
							iFrameInfo.GrenadeType = int32(grenadeProjectile.WeaponInstance.Type)
							iFrameInfo.GrenadeStartPos[0] = float32(grenadeProjectile.Position().X)
							iFrameInfo.GrenadeStartPos[1] = float32(grenadeProjectile.Position().Y)
							iFrameInfo.GrenadeStartPos[2] = float32(grenadeProjectile.Position().Z)
							iFrameInfo.GrenadeStartVel[0] = float32(grenadeProjectile.Velocity().X)
							iFrameInfo.GrenadeStartVel[1] = float32(grenadeProjectile.Velocity().Y)
							iFrameInfo.GrenadeStartVel[2] = float32(grenadeProjectile.Velocity().Z)
							iFrameInfo.ExtraData |= encoder.EXTRA_PLAYERDATA_GRENADE
							break
						}
					}
				}

				var currWeaponID int32
				if player.ActiveWeapon() != nil {
					currWeaponID = int32(WeaponStr2ID(player.ActiveWeapon().String()))
				}

				var IsFirstFrame bool = (len(encoder.PlayerFramesMap[player.Name]) == 0)
				if ((lastActiveWeapon[player.Name] != currWeaponID) || IsFirstFrame) {
					iFrameInfo.ActiveWeapon = currWeaponID
					iFrameInfo.ExtraData |= encoder.EXTRA_PLAYERDATA_EQUIPWEAPON
					lastActiveWeapon[player.Name] = iFrameInfo.ActiveWeapon
				}
				if ((lastOnGround[player.Name] != player.Flags().OnGround()) || IsFirstFrame) {
					iFrameInfo.OnGround = bool(player.Flags().OnGround())
					iFrameInfo.ExtraData |= encoder.EXTRA_PLAYERDATA_ON_GROUND
					lastOnGround[player.Name] = iFrameInfo.OnGround
				}
				if ((lastMoney[player.Name] != int32(player.Money())) || IsFirstFrame) {
					iFrameInfo.Money = int32(player.Money())
					iFrameInfo.ExtraData |= encoder.EXTRA_PLAYERDATA_MONEY
					lastMoney[player.Name] = iFrameInfo.Money
				}
				if ((lastOnGround[player.Name] != player.IsScoped()) || IsFirstFrame) {
					iFrameInfo.IsScoped = bool(player.IsScoped())
					iFrameInfo.ExtraData |= encoder.EXTRA_PLAYERDATA_IS_SCOPED
					lastIsScoped[player.Name] = iFrameInfo.IsScoped
				}

				iFrameInfo.PlayerButtons = ButtonConvert(player, addonButton)
				iFrameInfo.PlayerOrigin[0] = float32(player.Position().X)
				iFrameInfo.PlayerOrigin[1] = float32(player.Position().Y)
				iFrameInfo.PlayerOrigin[2] = float32(player.Position().Z)
				iFrameInfo.PlayerAngles[0] = player.ViewDirectionY()
				iFrameInfo.PlayerAngles[1] = player.ViewDirectionX()
				iFrameInfo.PlayerVelocity[0] = float32(player.Velocity().X)
				iFrameInfo.PlayerVelocity[1] = float32(player.Velocity().Y)
				iFrameInfo.PlayerVelocity[2] = float32(player.Velocity().Z) //?

				encoder.PlayerFramesMap[player.Name] = append(encoder.PlayerFramesMap[player.Name], *iFrameInfo)
			}
		}
		iParser.ParseNextFrame()
		i = iParser.GameState().IngameTick()
	}

	// Save Last Round
	gs := iParser.GameState()
	tScore := gs.TeamTerrorists().Score()
	ctScore := gs.TeamCounterTerrorists().Score()
	roundNum = int32(ctScore+tScore)
	if roundCorrupted || (roundNum == 0 && roundNumLast == 0) {
		// more than 10 round difference jump, this demo is f up, only consider this round
		fmt.Println("ignoring round: CT", ctScore, " - T", tScore)
	} else {
			if (math.Abs(float64(roundNum-roundNumLast))>10) {
				// more than 10 round difference jump, this demo is f up, only consider this round
				roundCorrupted = true
				// consider it as last round, give the winner a +
				roundNum = roundNumLast + 1
				// TODO: find the winner team?
			}
			for _, player := range roundValidPlayers {
				if player != nil {
					// save to rec file
					saveToRecFile(player, roundNum)
				}
			}
			fmt.Println("saved round", roundNum," round result: CT", ctScore, " - T", tScore)
	}
	// roundNumLast = roundNum
	// rountCanStart = true
	
	fmt.Printf("demo length (in ticks): %v\n", demoLength)
	err = iParser.ParseToEnd()
	checkError(err)
}
