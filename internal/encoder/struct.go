package encoder

// enum struct S_Demo_FileHeader {
//   int teamColor[4];
//   int steamId64; //for crosshair
//   int binaryFormatVersion;
//   int recordEndTime;
//   char recordName[MAX_RECORD_NAME_LENGTH];
//   int tickCount;
//   int tickRate;
//   int bookmarkCount;
//   float initialPosition[3];
//   float initialAngles[3];
//   ArrayList bookmarks;
//   ArrayList frames;
// }

type FrameInitInfo struct {
  PlayerName string
  Position   [3]float32
  Angles     [2]float32
  // CrosshairCode
}

// enum struct S_Demo_PlayerData {
//   int health;
//   // int flash_duration_time_remaining
//   bool has_helmet;
//   int armor;

//   int grenadeType; //EqSmoke, EqMolotov, EqDecoy, ...
//   float grenadeStartPos[3];
//   float grenadeStartVel[3];

//   ArrayList inventory;
//   // int ammoLeft;
//   int activeWeapon; // go inside S_Demo_FrameInfo ?(ammoLeft too) 

//   int money;
//   int on_ground;
// }

// enum struct S_Demo_FrameInfo {
//   float playerPos[3];
//   int playerButtons;
//   // on_ground flags
//   // score_assists
//   // score_deaths
//   // COPY DEMO VARIABLES SERVER: AIRTIME SV_GRAVITY ETC ETC
//   // flash_duration_time_remaining??? or correct nades only
//   float playerAngles[2];
//   // spec_show_xray, sv_competitive_official_5v5, sv_specnoclip(?), mp_forcecamera(?)
//   int playerData; // both of these will work as additionalfields, they change its value -> it will be asigned to the bot
// }

type FrameInfo struct {
  PlayerButtons     int32
  PlayerOrigin      [3]float32
  PlayerAngles      [2]float32
  PlayerVelocity    [3]float32
  ExtraData         int32

  Health            int32
  Helmet            bool
  Armor             int32
  OnGround          bool

  GrenadeType       int32
  GrenadeStartPos   [3]float32
  GrenadeStartVel   [3]float32

  Inventory         int32
  ActiveWeapon      int32

  Money             int32
  // Chat_Message
  IsScoped          int32
  // Score_Points      int32
  // Score_Kills       int32
  // Score_Deaths      int32
  // Score_Assists     int32
  // Score_MVP         int32

}