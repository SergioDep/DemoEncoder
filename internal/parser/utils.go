package parser

import (
	encoder "github.com/hx-w/minidemo-encoder/internal/encoder"
	ilog "github.com/hx-w/minidemo-encoder/internal/logger"
	common "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/common"
)

var lastActiveWeapon map[string]int32 = make(map[string]int32)
var lastOnGround map[string]bool = make(map[string]bool)
var lastMoney map[string]int32 = make(map[string]int32)
var lastIsScoped map[string]bool = make(map[string]bool)

// Function to handle errors
func checkError(err error) {
	if err != nil {
		ilog.ErrorLogger.Println(err.Error())
	}
}

func parsePlayerInitFrame(player *common.Player) {
	iFrameInit := encoder.FrameInitInfo{
		PlayerName: player.Name,
	}
	iFrameInit.Position[0] = float32(player.Position().X)
	iFrameInit.Position[1] = float32(player.Position().Y)
	iFrameInit.Position[2] = float32(player.Position().Z)
	// Pay attention to XY, need to test
	iFrameInit.Angles[0] = float32(player.ViewDirectionY())
	iFrameInit.Angles[1] = float32(player.ViewDirectionX())

	encoder.InitPlayer(iFrameInit)
	delete(encoder.PlayerFramesMap, iFrameInit.PlayerName)
	delete(lastActiveWeapon, player.Name)
	delete(lastOnGround, player.Name)
	delete(lastMoney, player.Name)
	delete(lastIsScoped, player.Name)
}

func saveToRecFile(player *common.Player, roundNum int32) {
	var playerTeam string
	if (player.Team==common.TeamTerrorists) {
		playerTeam = "TT"
	} else {
		playerTeam = "CT"
	}
	encoder.WriteToRecFile(player.Name, playerTeam, roundNum)
}
