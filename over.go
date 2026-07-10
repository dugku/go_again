package main

import "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

type PositionRecord struct {
	MatchID      string
	MatchMap     string
	Round        int
	Tick         int
	Time         float64
	SteamID      uint64
	Side         int
	X            float64
	Y            float64
	ViewX        float32
	ViewY        float32
	IsFreezeTime bool
}

type PlayerPosition struct {
	PlayerSteamId uint64
	X             float64
	Y             float64

	XViewDir float32
	YViewdir float32
}

type State struct {
	RoundOngoing     bool
	Round            int
	TeamASide        int
	TeamBSide        int
	LastCapturedTick int
	SkippedFrames    int
	WarmupKills      []events.Kill
}

type Match struct {
	MatchID   string
	Round     []Rounds
	MapName   string
	date      string
	Players   map[uint64]PlayerStats
	Positions []PositionRecord `json:"-"`
}

type Rounds struct {
	//Round Scores
	RoundNum   int
	TeamAScore int
	TeamBScore int

	//times
	RoundStartTime int
	RoundEndTime   int

	//Round end reasons
	RoundEndReason string

	//Econ Stats
	EconA       int
	EconB       int
	CTEquipVal  int
	TEquipVal   int
	TypeOfBuyCT string
	TypeOfBuyT  string

	//For KAST maybe idk
	SurvivorsTeamA []uint64
	SurvivorsTeamB []uint64

	//Kills
	FirstKillCount int
	RoundKills     []Kills
	IsOneVX        bool
	OneVXCount     int

	//Bomb Stuff
	BombPlanted     bool
	PlayerPlanted   string
	BombPlantedSite string

	//Ticks []Tick

	TeamAAlive   []uint64
	TeamBAlive   []uint64
	ClutcherID   uint64
	ClutcherTeam int
}

type Kills struct {
	//Time Stats
	TimeOfKill int //This will need to be converted to seconds because this needs to be in JSON yeah
	Tick       int //to ensure consistency.
	//Personal Identifiers
	VictimID  uint64
	KillerID  uint64
	AsistorID uint64
	//Poisitions
	KillerX float64
	KillerY float64
	VictimX float64
	VictimY float64 //Shoudln't need z since it has nothing to do with mapping anything to 2d coordinates.
	//Assistor..?
	WasAssisted bool
	//Weapon Stats
	WeaponName  string
	WeaponClass int
	//Blinded Stats
	KillerFlashed bool
	VictimFlashed bool
	//Team
	KillerTeam int
	VictimTeam int
	//Misc
	IsOpening       bool
	IsHeadshot      bool
	IsWallbang      bool
	IsNoscope       bool
	IsThroughSmoke  bool
	IsFlashAssisted bool
}

type PlayerStats struct {
	Username string
	SteamID  uint64

	//More Stats
	Kills               int
	Deaths              int
	Assists             int
	HS                  int
	HeadshotPercent     float64
	ADR                 float64
	KAST                float64
	eKAST               float64 //I will need to figureout how to do that because how does one eco adjust stuff like wtf..?
	Impact              float64
	KDRatio             float64
	FirstKill           int
	FirstDeath          int
	FKDiff              int
	Round2k             int
	Round3k             int
	Round4k             int
	Round5k             int
	TotalDmg            int
	TradeKills          int
	TradeDeaths         int
	CTKills             int
	TKills              int
	EffectiveFlashes    int
	AvgFlashDuration    int //Need to convert to seconds for int and stuff
	WeaponKills         map[int]int
	WeaponKillsHeadshot map[int]int
	AvgDist             float64
	TotalDist           float64
	FlashesThrown       int
	TotalUtilDamage     int
	AvgKillRnd          float64
	AvgDeathsRnd        float64
	AvgAssistsRnd       float64
	AvgNadeDmg          float64
	AvgInferDmg         float64
	RoundsSurvived      int
	RoundTraded         int
	RoundContrid        []int
	InfernoDmg          int
	NadeDmg             int
	TotalOpening        int
	OpeningAttSuccess   int
	OpeingAttpFail      int
	OpeningPercent      float64
	OpeningAttpPercent  float64
	OpeningRoundsWon    int
	OpeningWinPercent   float64
	OneVsOne            int
	OneVsTwo            int
	OneVsThree          int
	OneVsFour           int
	OneVsFive           int
	HeadToHead          map[uint64]int
}
