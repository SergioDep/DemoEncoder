package encoder

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"time"
)

var __MAGIC__ int32 = -559038737
var __FORMAT_VERSION__ int8 = 3

var SvTickRate int32 = 128
var bufMap map[string]*bytes.Buffer = make(map[string]*bytes.Buffer)
var PlayerFramesMap map[string][]FrameInfo = make(map[string][]FrameInfo)

var EXTRA_PLAYERDATA_HEALTH       int32 = 1 << 0
var EXTRA_PLAYERDATA_HELMET       int32 = 1 << 1
var EXTRA_PLAYERDATA_ARMOR        int32 = 1 << 2
var EXTRA_PLAYERDATA_ON_GROUND    int32 = 1 << 3
var EXTRA_PLAYERDATA_GRENADE      int32 = 1 << 4
var EXTRA_PLAYERDATA_INVENTORY    int32 = 1 << 5
var EXTRA_PLAYERDATA_EQUIPWEAPON  int32 = 1 << 6
var EXTRA_PLAYERDATA_MONEY        int32 = 1 << 7
var EXTRA_PLAYERDATA_CHAT         int32 = 1 << 8
var EXTRA_PLAYERDATA_IS_SCOPED    int32 = 1 << 9

var saveDir string

func InitDemo(fileName string, mapName string) {
	fileName = strip(fileName)
	saveDir = "./output/" + mapName + "/" + fileName[:len(fileName)-3]
	if ok, _ := PathExists(saveDir); !ok {
		os.MkdirAll(saveDir, os.ModePerm)
	}
}

func InitPlayer(initFrame FrameInitInfo) {
	// fmt.Println("(InitPlayer)saving: ", initFrame.PlayerName)
	if bufMap[initFrame.PlayerName] == nil {
		bufMap[initFrame.PlayerName] = new(bytes.Buffer)
	} else {
		bufMap[initFrame.PlayerName].Reset()
	}
	// FileHeaders

	// step.1 MAGIC NUMBER
	WriteToBuf(initFrame.PlayerName, __MAGIC__)

	// step.2 VERSION
	WriteToBuf(initFrame.PlayerName, __FORMAT_VERSION__)

	// step.3 timestamp
	WriteToBuf(initFrame.PlayerName, int32(time.Now().Unix()))

	// step.4 name length
	WriteToBuf(initFrame.PlayerName, int8(len(initFrame.PlayerName)))

	// step.5 name
	WriteToBuf(initFrame.PlayerName, []byte(initFrame.PlayerName))

	// step.6 initial position
	for idx := 0; idx < 3; idx++ {
		WriteToBuf(initFrame.PlayerName, float32(initFrame.Position[idx]))
	}

	// step.7 initial angle
	for idx := 0; idx < 2; idx++ {
		WriteToBuf(initFrame.PlayerName, initFrame.Angles[idx])
	}
	// ilog.InfoLogger.Println("Inicializado correctamente: ", initFrame.PlayerName)
}

func WriteToRecFile(playerName string, playerTeam string, roundNum int32) {
	subDir := saveDir + "/" + strconv.Itoa(int(roundNum))
	if ok, _ := PathExists(subDir); !ok {
		os.MkdirAll(subDir, os.ModePerm)
	}

	playerNameFile := strip(playerName)

	fileName := subDir + "/" + playerTeam + "_" + playerNameFile + ".rec"

	file, err := os.Create(fileName) // Crea archivo, "binbin" es el nombre del archivo
	if err != nil {
		// fmt.Println("Error al crear archivo", err.Error())
		return
	}
	defer file.Close()

	// step.8 tick count
	var tickCount int32 = int32(len(PlayerFramesMap[playerName]))
	// fmt.Println("round: ", roundNum, " jugador: ", fileName, ", tick count: ", tickCount, "tick rate: ", SvTickRate)
	WriteToBuf(playerName, tickCount)

	// step.8.5 tick rate

	WriteToBuf(playerName, SvTickRate)

	// step.9 bookmark count
	WriteToBuf(playerName, int32(0))

	// step.10 all bookmark
	// ignore

	// step.11 all tick frame
	for _, frame := range PlayerFramesMap[playerName] {
		WriteToBuf(playerName, frame.PlayerButtons)
		for idx := 0; idx < 3; idx++ {
			WriteToBuf(playerName, frame.PlayerOrigin[idx])
		}
		for idx := 0; idx < 2; idx++ {
			WriteToBuf(playerName, frame.PlayerAngles[idx])
		}
		for idx := 0; idx < 3; idx++ {
			WriteToBuf(playerName, frame.PlayerVelocity[idx])
		}
		WriteToBuf(playerName, frame.ExtraData)
		// Extra Player Data
		if frame.ExtraData&EXTRA_PLAYERDATA_HEALTH != 0 {
			WriteToBuf(playerName, frame.Health)
		}
		if frame.ExtraData&EXTRA_PLAYERDATA_HELMET != 0 {
			WriteToBuf(playerName, frame.Helmet)
		}
		if frame.ExtraData&EXTRA_PLAYERDATA_ARMOR != 0 {
			WriteToBuf(playerName, frame.Armor)
		}
		if frame.ExtraData&EXTRA_PLAYERDATA_ON_GROUND != 0 {
			WriteToBuf(playerName, frame.OnGround)
		}
		if frame.ExtraData&EXTRA_PLAYERDATA_GRENADE != 0 {
			WriteToBuf(playerName, frame.GrenadeType)
			for idx := 0; idx < 3; idx++ {
				WriteToBuf(playerName, frame.GrenadeStartPos[idx])
			}
			for idx := 0; idx < 3; idx++ {
				WriteToBuf(playerName, frame.GrenadeStartVel[idx])
			}
		}
		// if frame.ExtraData&EXTRA_PLAYERDATA_INVENTORY != 0 {
		// 	WriteToBuf(playerName, frame.Health)
		// }
		if frame.ExtraData&EXTRA_PLAYERDATA_EQUIPWEAPON != 0 {
			WriteToBuf(playerName, frame.ActiveWeapon)
		}
		if frame.ExtraData&EXTRA_PLAYERDATA_MONEY != 0 {
			WriteToBuf(playerName, frame.Money)
		}
		// if frame.ExtraData&EXTRA_PLAYERDATA_CHAT != 0 {
		// 	WriteToBuf(playerName, frame.ChatMessage)
		// }
    if frame.ExtraData&EXTRA_PLAYERDATA_IS_SCOPED != 0 {
      WriteToBuf(playerName, frame.IsScoped)
    }
	}

	delete(PlayerFramesMap, playerName)
	file.Write(bufMap[playerName].Bytes())
	// ilog.InfoLogger.Printf("[ronda %d] Rec guardado correctamente: %s.rec\n", roundNum, playerName)
}

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}
