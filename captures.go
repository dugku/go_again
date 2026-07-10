package main

import (
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
)

func capture_positions(m *Match, s *State, p demoinfocs.Parser) {
	gs := p.GameState()

	base := PositionRecord{
		MatchID:      m.MatchID,
		MatchMap:     m.MapName,
		Round:        s.Round,
		Tick:         gs.IngameTick(),
		Time:         p.CurrentTime().Seconds(),
		IsFreezeTime: gs.IsFreezetimePeriod(),
	}

	if ct := gs.TeamCounterTerrorists(); ct != nil {
		capture_team(m, ct.Members(), base)
	}
	if tt := gs.TeamTerrorists(); tt != nil {
		capture_team(m, tt.Members(), base)
	}
}

func capture_team(m *Match, players []*common.Player, base PositionRecord) {
	for _, pl := range players {
		if pl == nil || !pl.IsAlive() {
			continue
		}
		rec := base // copy the shared fields
		rec.SteamID = pl.SteamID64
		rec.Side = int(pl.Team)
		rec.X = pl.Position().X
		rec.Y = pl.Position().Y
		rec.ViewX = pl.ViewDirectionX()
		rec.ViewY = pl.ViewDirectionY()
		m.Positions = append(m.Positions, rec)
	}
}
