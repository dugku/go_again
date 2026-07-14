package main

import (
	"fmt"
	"log"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

var (
	player_map = make(map[uint64]int)
)

var (
	OneVX = map[int]string{
		1: "OneVsOne",
		2: "OneVsTwo",
		3: "OneVsThree",
		4: "OneVsFour",
		5: "OneVsFive",
	}
)

func player_get(c *common.Player) PlayerStats {
	return PlayerStats{
		Username: c.Name,
		SteamID:  c.SteamID64,

		Kills:               0,
		Deaths:              0,
		Assists:             0,
		HS:                  0,
		HeadshotPercent:     0,
		ADR:                 0,
		KAST:                0,
		eKAST:               0,
		KDRatio:             0,
		FirstKill:           0,
		FirstDeath:          0,
		FKDiff:              0,
		Round2k:             0,
		Round3k:             0,
		Round4k:             0,
		Round5k:             0,
		TotalDmg:            0,
		TradeKills:          0,
		TradeDeaths:         0,
		CTKills:             0,
		TKills:              0,
		EffectiveFlashes:    0,
		AvgFlashDuration:    0,
		WeaponKills:         all_weapons(),
		WeaponKillsHeadshot: all_weapons(),
		AvgDist:             0,
		TotalDist:           0,
		FlashesThrown:       0,
		TotalUtilDamage:     0,
		AvgKillRnd:          0,
		AvgDeathsRnd:        0,
		AvgAssistsRnd:       0,
		AvgNadeDmg:          0,
		AvgInferDmg:         0,
		RoundsSurvived:      0,
		RoundTraded:         0,
		RoundContrid:        []int{},
		InfernoDmg:          0,
		NadeDmg:             0,
		OpeningPercent:      0,
		OpeningAttpPercent:  0,
		OpeningRoundsWon:    0,
		OpeningWinPercent:   0,
		OneVsOne:            0,
		OneVsTwo:            0,
		OneVsThree:          0,
		OneVsFour:           0,
		OneVsFive:           0,
		HeadToHead:          make_head_to_head(),
	}
}

func all_weapons() map[int]int { return make(map[int]int) }

func make_head_to_head() map[uint64]int { return make(map[uint64]int) }

func player_getter(m *Match, s *State, p demoinfocs.Parser) {
	p.RegisterEventHandler(func(e events.RoundEnd) {
		gs := p.GameState()

		TeamA := gs.Team(common.Team(s.TeamASide))
		TeamB := gs.Team(common.Team(s.TeamBSide))

		if TeamA == nil || TeamB == nil {
			log.Println("The Team obj is nil skipping player getter")
			return
		}

		player_a := TeamA.Members()
		player_b := TeamB.Members()

		if player_a == nil || player_b == nil {
			log.Println("One or both of the member list is also nil")
			return
		}

		stat_setter(player_a, m, gs)
		stat_setter(player_b, m, gs)
	})
}

func stat_setter(c []*common.Player, m *Match, gs demoinfocs.GameState) {
	for i := range c {
		steam_id := c[i].SteamID64
		p, exists := m.Players[steam_id]
		if !exists {
			continue
		}
		//TIME FOR THE STAT HAHAHAHHA
		p.Kills = c[i].Kills()
		p.Deaths = c[i].Deaths()
		p.TotalDmg = c[i].TotalDamage()
		p.Assists = c[i].Assists()
		p.ADR = calc_adr(gs, c[i].TotalDamage())
		p.HeadshotPercent = calc_hs_per(p.HS, c[i].Kills())
		p.KDRatio = calc_kd(c[i].Kills(), c[i].Deaths())
		p.AvgKillRnd = calc_avg_per_round(p.Kills, gs.TotalRoundsPlayed())
		p.AvgDeathsRnd = calc_avg_per_round(p.Deaths, gs.TotalRoundsPlayed())
		p.AvgAssistsRnd = calc_avg_per_round(p.Assists, gs.TotalRoundsPlayed())
		p.Impact = calc_impatct(p.AvgKillRnd, p.AvgAssistsRnd)
		p.KAST = calc_kast(len(p.RoundContrid), gs.TotalRoundsPlayed())
		p.AvgNadeDmg = calc_avg_per_round(p.NadeDmg, gs.TotalRoundsPlayed())
		p.AvgInferDmg = calc_avg_per_round(p.InfernoDmg, gs.TotalRoundsPlayed())
		p.TotalOpening = p.OpeningAttSuccess + p.OpeingAttpFail
		p.OpeningAttpPercent = calc_open_percent(p.TotalOpening, gs.TotalRoundsPlayed())

		multi_check := c[i].Kills() - player_map[steam_id]

		switch {
		case multi_check == 2:
			p.Round2k++
		case multi_check == 3:
			p.Round3k++
		case multi_check == 4:
			p.Round4k++
		case multi_check == 5:
			p.Round5k++
		}

		m.Players[steam_id] = p
	}
}

func get_pres_round_kills(m *Match, s *State, p demoinfocs.Parser) {
	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		gs := p.GameState()

		teamA := gs.Team(common.Team(s.TeamASide))
		teamB := gs.Team(common.Team(s.TeamBSide))

		if teamA != nil {
			make_playerMap(teamA.Members())
		} else {
			log.Println("Warning: TeamA is nil; cannot retrieve members.")
		}

		if teamB != nil {
			make_playerMap(teamB.Members())
		} else {
			log.Println("Warning: TeamB is nil; cannot retrieve members.")
		}
	})
}

func make_playerMap(c []*common.Player) {
	for _, v := range c {
		player_map[v.SteamID64] = v.Kills()
	}
}

func kill_controller(m *Match, s *State, p demoinfocs.Parser) {
	p.RegisterEventHandler(func(e events.Kill) {
		gs := p.GameState()
		if e.Killer == nil || e.Victim == nil {
			return
		}

		if p.GameState().IsWarmupPeriod() {
			s.WarmupKills = append(s.WarmupKills, e)
			return
		}

		if e.IsHeadshot && e.Weapon != nil {
			add_headshot(e.Killer, e.Weapon.Type, m)
		}

		//ar assistor string
		var assistor_id uint64
		var IsAssisted bool
		if e.Assister != nil {
			//assistor = e.Assister.Name
			assistor_id = e.Assister.SteamID64
			IsAssisted = true
		}

		if e.Killer.PlayerPawnEntity() == nil {
			fmt.Println("Player Pawn is nil.")
		}

		if s.Round > 0 && s.Round <= len(m.Round) {
			round := &m.Round[s.Round-1]
			open_kill := len(round.RoundKills) == 0
			/*
				Since these are not slices couldn't I just get the opening kill after processing round.
				Because Slices keep their order after appending.?
			*/
			weaponName := "unknown"
			if e.Weapon != nil {
				weaponName = e.Weapon.String()
				update_weapon_kill(e.Killer, e.Weapon.Type, m)
			}

			kill := Kills{
				TimeOfKill: int(p.CurrentTime().Seconds()),
				Tick:       gs.IngameTick(),

				VictimID:  e.Victim.SteamID64,
				KillerID:  e.Killer.SteamID64,
				AsistorID: assistor_id,

				KillerX: e.Killer.Position().X,
				KillerY: e.Killer.Position().Y,
				VictimX: e.Victim.Position().X,
				VictimY: e.Victim.Position().Y,

				WasAssisted: IsAssisted,
				WeaponName:  weaponName,
				//WeaponClass: e.Killer.ActiveWeapon().Class(),

				KillerFlashed: e.Killer.IsBlinded(),
				VictimFlashed: e.Victim.IsBlinded(),

				KillerTeam: int(e.Killer.Team),
				VictimTeam: int(e.Victim.Team),

				IsOpening:  open_kill,
				IsHeadshot: e.IsHeadshot,
				IsWallbang: e.IsWallBang(),
				//IsNoscope: e.NoScope(),
				IsThroughSmoke:  e.ThroughSmoke,
				IsFlashAssisted: e.AssistedFlash,
			}

			m.Round[s.Round-1].RoundKills = append(m.Round[s.Round-1].RoundKills, kill)
			open_kill = false

			if e.Killer.Team != e.Victim.Team {
				if pl, ok := m.Players[e.Killer.SteamID64]; ok {
					pl.HeadToHead[e.Victim.SteamID64]++
					m.Players[e.Killer.SteamID64] = pl
				}
			}

			roundkill_contrib(m, s, p, e)
			one_vs_number(m, s, p, e)
			detect_clutch(&m.Round[s.Round-1], s)
		}
	})
}

func roundkill_contrib(m *Match, s *State, p demoinfocs.Parser, e events.Kill) {
	kill_id := e.Killer.SteamID64
	pl, exists := m.Players[kill_id]
	if !exists {
		return
	}
	pl.RoundContrid = append(pl.RoundContrid, s.Round)
	m.Players[kill_id] = pl

	if e.Assister != nil {
		if player_ex, ok := m.Players[e.Assister.SteamID64]; ok {
			player_ex.RoundContrid = append(player_ex.RoundContrid, s.Round)
			m.Players[e.Assister.SteamID64] = player_ex
		}
	}
}

func update_weapon_kill(c *common.Player, weapon_type common.EquipmentType, m *Match) {
	if weapon_type == 407 {
		return
	}

	player_id := c.SteamID64
	player, exists := m.Players[player_id]
	if !exists {
		return
	}

	player.WeaponKills[int(weapon_type)]++
	m.Players[player_id] = player
}

func one_vs_number(m *Match, s *State, p demoinfocs.Parser, e events.Kill) {
	round_info := &m.Round[s.Round-1]

	if int(e.Victim.Team) == s.TeamASide {
		vict_id := e.Victim.SteamID64
		round_info.TeamAAlive = delete_players(vict_id, round_info.TeamAAlive)
	}
	if int(e.Victim.Team) == s.TeamBSide {
		vict_id := e.Victim.SteamID64
		round_info.TeamBAlive = delete_players(vict_id, round_info.TeamBAlive)
	}

}

func delete_players(victId uint64, alive []uint64) []uint64 {
	for i, id := range alive {
		if victId == id {
			return append(alive[:i], alive[i+1:]...)
		}
	}

	return alive
}

func detect_clutch(r *Rounds, s *State) {
	if r.IsOneVX {
		return
	}

	a, b := len(r.TeamAAlive), len(r.TeamBAlive)

	switch {
	case a == 1 && b >= 1:
		r.IsOneVX = true
		r.OneVXCount = b
		r.ClutcherID = r.TeamAAlive[0]
		r.ClutcherTeam = s.TeamASide
	case b == 1 && a >= 1:
		r.IsOneVX = true
		r.OneVXCount = a
		r.ClutcherID = r.TeamBAlive[0]
		r.ClutcherTeam = s.TeamBSide
	}
}

func award_clutch(m *Match, r *Rounds, winner common.Team) {
	if !r.IsOneVX || r.ClutcherTeam != int(winner) {
		return
	}
	player, exists := m.Players[r.ClutcherID]
	if !exists {
		return
	}
	switch r.OneVXCount {
	case 1:
		player.OneVsOne++
	case 2:
		player.OneVsTwo++
	case 3:
		player.OneVsThree++
	case 4:
		player.OneVsFour++
	case 5:
		player.OneVsFive++
	}

	m.Players[r.ClutcherID] = player
}

func add_headshot(c *common.Player, w common.EquipmentType, m *Match) {
	player_id := c.SteamID64
	player, exists := m.Players[player_id]
	if !exists {
		return
	}

	player.HS++
	//player.WeaponKillHS[int(w)]++
	m.Players[player_id] = player
}
