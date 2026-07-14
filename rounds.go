package main

import (
	"fmt"
	"log"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

var RoundEndReasonCategory = map[events.RoundEndReason]string{
	events.RoundEndReasonBombDefused:         "bomb_defused",
	events.RoundEndReasonTargetBombed:        "target_bombed",
	events.RoundEndReasonTargetSaved:         "time_expired",
	events.RoundEndReasonCTWin:               "ct_elimination_ct_won",
	events.RoundEndReasonTerroristsWin:       "t_elimination_t_won",
	events.RoundEndReasonCTSurrender:         "surrender",
	events.RoundEndReasonTerroristsSurrender: "surrender",
}

// This will determine the round num so kinda important to get this right
// this will also add the round i think
func state(m *Match, s *State, p demoinfocs.Parser) {
	fmt.Println("In State")
	gs := p.GameState()

	p.RegisterEventHandler(func(e events.RoundStart) {
		s.RoundOngoing = true
		s.Round += 1

		round := Rounds{}
		round.RoundStartTime = int(p.CurrentTime().Seconds())

		TeamA := gs.Team(common.Team(s.TeamASide)).Members()
		TeamB := gs.Team(common.Team(s.TeamBSide)).Members()

		if TeamA == nil || TeamB == nil {
			log.Println("The Team obj is nil in state uh no..")
			return
		}

		append_team_alive(&round, s, TeamA)
		append_team_alive(&round, s, TeamB)

		m.Round = append(m.Round, round)

	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		s.RoundOngoing = false

	})
}

func append_team_alive(r *Rounds, s *State, c []*common.Player) {
	for _, p := range c {
		steam_id := p.SteamID64

		if int(p.Team) == s.TeamASide {
			r.TeamAAlive = append(r.TeamAAlive, steam_id)
		}
		if int(p.Team) == s.TeamBSide {
			r.TeamBAlive = append(r.TeamBAlive, steam_id)
		}
	}
}

func match_started(m *Match, s *State, p demoinfocs.Parser) {
	p.RegisterEventHandler(func(e events.MatchStart) {
		fmt.Println("Here")
		gs := p.GameState()
		s.Round = 0
		m.Round = m.Round[:0]
		s.LastCapturedTick = 0
		players := gs.Participants().Playing()

		if m.Players == nil {
			m.Players = make(map[uint64]PlayerStats)
		}
		for _, pl := range players {
			if _, exists := m.Players[pl.SteamID64]; !exists {
				m.Players[pl.SteamID64] = player_get(pl)
			}
		}

		team_a := common.TeamCounterTerrorists
		team_b := common.TeamTerrorists
		//There are Demos where the team is nil so at the start we should check that no?
		//It could also be nil in the middle of the demo maybe?
		//Haven't seen that yet.
		if team_a == common.TeamUnassigned || team_b == common.TeamUnassigned {
			log.Printf("Unassigned team state encountered: TeamA = %v; TeamB = %v", team_a, team_b)
			return
		}

		s.TeamASide = int(team_a)
		s.TeamBSide = int(team_b)

	})
}

func team_scoring(m *Match, s *State, p demoinfocs.Parser) {
	p.RegisterEventHandler(func(e events.ScoreUpdated) {
		gs := p.GameState()
		if s.TeamASide == 0 || s.TeamBSide == 0 {
			return // MatchStart hasn't set sides yet
		}
		if s.Round <= 0 || s.Round > len(m.Round) {
			return
		}
		ta := gs.Team(common.Team(s.TeamASide))
		tb := gs.Team(common.Team(s.TeamBSide))
		if ta == nil || tb == nil {
			return
		}
		m.Round[s.Round-1].RoundNum = s.Round
		m.Round[s.Round-1].TeamAScore = ta.Score()
		m.Round[s.Round-1].TeamBScore = tb.Score()
	})
}

func team_switch(p demoinfocs.Parser, m *Match, s *State) {
	p.RegisterEventHandler(func(e events.TeamSideSwitch) {
		TempSide := s.TeamASide
		s.TeamASide = s.TeamBSide
		s.TeamBSide = TempSide

	})
}

func econ_controller(m *Match, s *State, p demoinfocs.Parser) {
	gs := p.GameState()

	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		if s.Round <= 0 || s.Round > len(m.Round) {
			return
		}
		Mround := &m.Round[s.Round-1]

		TeamA_EquipVal := gs.Team(common.Team(s.TeamASide)).CurrentEquipmentValue()
		TeamB_EquipVal := gs.Team(common.Team(s.TeamBSide)).CurrentEquipmentValue()

		Mround.CTEquipVal = TeamA_EquipVal
		Mround.TEquipVal = TeamB_EquipVal
		Mround.TypeOfBuyCT = assess_econ(TeamA_EquipVal)
		Mround.TypeOfBuyT = assess_econ(TeamB_EquipVal)
	})

}

func assess_econ(team_econ int) string {
	FullBuy := 20000
	HalfBuy := 10000
	SemiEco := 5000

	switch {
	case team_econ >= FullBuy:
		return "Full Buy"
	case team_econ >= HalfBuy:
		return "Half Buy"
	case team_econ >= SemiEco:
		return "Force Buy"
	default:
		return "Eco"
	}
}

// Why this isn't in state is because it'll probably be more clean in here maybe idk.
func round_end_controller(m *Match, s *State, p demoinfocs.Parser) {
	p.RegisterEventHandler(func(e events.RoundEnd) {
		if m == nil || s == nil {
			return
		}

		gs := p.GameState()
		if gs == nil {
			return
		}
		if gs.IsWarmupPeriod() {
			return
		}

		if s.Round <= 0 {
			return
		}

		if s.Round > len(m.Round) {
			return
		}

		teamA := common.Team(s.TeamASide)
		teamB := common.Team(s.TeamBSide)

		if teamA == common.TeamUnassigned || teamA == common.TeamSpectators {
			return
		}

		if teamB == common.TeamUnassigned || teamB == common.TeamSpectators {
			return
		}

		if teamA == teamB {
			return
		}

		players_a := gs.Team(teamA).Members()
		players_b := gs.Team(teamB).Members()

		if len(players_a) == 0 || len(players_b) == 0 {
			return
		}
		currRound := &m.Round[s.Round-1]

		round_end_reason(m, s, p, e)
		round_contributed(m, s, p, players_a, gs)
		round_contributed(m, s, p, players_b, gs)
		trade_kill_cal(m, s, p)
		award_clutch(m, currRound, e.Winner)
		opening_kills(m, s, p)
		opening_kill_win(m, s, p, e)
		team_side_kills(m, s, p)
		remove_dups(p, m, players_a)
		remove_dups(p, m, players_b)
	})

	p.RegisterEventHandler(func(e events.RoundEndOfficial) {
		s.RoundOngoing = false
	})
}

func round_end_reason(m *Match, s *State, p demoinfocs.Parser, e events.RoundEnd) {
	reason := e.Reason

	if s.Round <= 0 || s.Round > len(m.Round) {
		return
	}
	round := &m.Round[s.Round-1]
	cate, ok := RoundEndReasonCategory[reason]
	if !ok {
		cate = "Other"
	}
	round.RoundEndReason = cate

}

func round_contributed(m *Match, s *State, p demoinfocs.Parser, c []*common.Player, gs demoinfocs.GameState) {
	if s.Round <= 0 || s.Round > len(m.Round) {
		return
	}
	round := &m.Round[s.Round-1]
	for _, p := range c {
		if p.IsAlive() {
			id := p.SteamID64

			//get the players team can't fuck this up or else the stat is not gonna work ohno
			if int(p.Team) == s.TeamASide {
				round.SurvivorsTeamA = append(round.SurvivorsTeamA, id)
			} else {
				round.SurvivorsTeamB = append(round.SurvivorsTeamB, id)
			}

			player, exists := m.Players[id]
			if !exists {
				continue
			}
			player.RoundsSurvived++
			player.RoundContrid = append(player.RoundContrid, s.Round)
			m.Players[id] = player
		}
	}
}

// low kay can get this one done yet bevause need the kill function hm..
func trade_kill_cal(m *Match, s *State, p demoinfocs.Parser) {
	if s.Round <= 0 || s.Round > len(m.Round) {
		return
	}
	r := &m.Round[s.Round-1]
	//tradeWindowTicks := int(5 * p.TickRate())

	for i := 0; i < len(r.RoundKills); i++ {
		current := r.RoundKills[i]
		for j := i - 1; j >= 0; j-- {
			//if the victim of the next kill is the killer of the current (previous in this case)
			//then if the time of that kill is less than 5 secondsish then that is a trade kill?
			//of course the time of subjective for the trade kill since everyone has different ideas on that.
			prev := r.RoundKills[j]
			if current.TimeOfKill-prev.TimeOfKill > 5 {
				break
			}

			if current.KillerTeam != prev.VictimTeam {
				continue
			}
			if current.VictimID != prev.KillerID {
				continue
			}

			if killer, exist := m.Players[current.KillerID]; exist {
				killer.TradeKills++
				m.Players[current.KillerID] = killer
			}
			if vict, exists := m.Players[prev.VictimID]; exists {
				vict.TradeDeaths++
				vict.RoundContrid = append(vict.RoundContrid, s.Round)
				m.Players[prev.VictimID] = vict
			}
			break
		}

	}
}

func team_side_kills(m *Match, s *State, p demoinfocs.Parser) {
	if s.Round <= 0 || s.Round > len(m.Round) {
		return
	}

	r := &m.Round[s.Round-1]

	for i := 0; i < len(r.RoundKills); i++ {
		val := r.RoundKills[i]

		killer_id := val.KillerID

		player, exists := m.Players[killer_id]
		if !exists {
			continue
		}

		if val.KillerID == val.VictimID {
			continue
		}

		if val.KillerTeam == 3 {
			player.CTKills++
		}

		if val.KillerTeam == 2 {
			player.TKills++
		}

		m.Players[killer_id] = player
	}
}

func opening_kills(m *Match, s *State, p demoinfocs.Parser) {
	if s.Round <= 0 || s.Round > len(m.Round) {
		return
	}

	r := &m.Round[s.Round-1]

	l := len(r.RoundKills)

	//since in a rare instance where this can be nil we need to
	//guard against it.
	if l == 0 {
		return
	}

	first_kill := r.RoundKills[0]
	//This also includes teamamtes sadly... But will change later probably..
	if first_kill.IsOpening == true {
		f_k_id := first_kill.KillerID
		vict_id := first_kill.VictimID

		player, exists := m.Players[f_k_id]
		if !exists {
			return
		}
		player.FirstKill++
		player.OpeningAttSuccess++
		m.Players[f_k_id] = player

		vict, v_ex := m.Players[vict_id]
		if !v_ex {
			return
		}
		vict.FirstDeath++
		vict.OpeingAttpFail++
		m.Players[vict_id] = vict
	}

}

func opening_kill_win(m *Match, s *State, p demoinfocs.Parser, e events.RoundEnd) {
	if s.Round <= 0 || s.Round > len(m.Round) {
		return
	}

	r := &m.Round[s.Round-1]

	l := len(r.RoundKills)

	//since in a rare instance where this can be nil we need to
	//guard against it.
	if l == 0 {
		return
	}

	first_kill := r.RoundKills[0]

	if first_kill.IsOpening == true {
		killer_t := first_kill.KillerTeam

		if killer_t == int(e.Winner) {
			killer_id := first_kill.KillerID
			player, exists := m.Players[killer_id]
			if !exists {
				return
			}

			player.OpeningRoundsWon++
			m.Players[killer_id] = player
		}
	}
}

func bom_planted(p demoinfocs.Parser, m *Match, s *State) {
	p.RegisterEventHandler(func(e events.BombPlanted) {
		//gs := p.GameState()

		if m == nil || s == nil {
			log.Println("This is whats wrong", m, s)
			return
		}

		if s.Round <= 0 || s.Round > len(m.Round) {
			log.Println("Something within the round is wrong.")
			return
		}

		if e.Player == nil {
			log.Println("Something with player is wrong ")
			return
		}

		round_info := &m.Round[s.Round-1]

		round_info.BombPlanted = true
		round_info.PlayerPlanted = e.Player.Name
		round_info.BombPlantedSite = string(e.Site)
	})
}

func remove_dups(p demoinfocs.Parser, m *Match, c []*common.Player) {

	for _, pl := range c {
		player_id := pl.SteamID64

		player, exists := m.Players[player_id]
		if !exists {
			continue
		}

		seen := make(map[int]bool)
		var result []int

		for _, val := range player.RoundContrid {
			if !seen[val] {
				// If we haven't seen this value yet, append it to result
				seen[val] = true
				result = append(result, val)
			}
		}

		player.RoundContrid = result
		m.Players[player_id] = player
	}
}
