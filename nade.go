package main

import (
	"log"
	"time"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

func flash_logic(p demoinfocs.Parser, m *Match, s *State) {
	p.RegisterEventHandler(func(e events.PlayerFlashed) {
		if e.Attacker == nil || e.Player == nil {
			return
		}

		// enemy flashes only — this also excludes self-flashes,
		// since you're on your own team
		if e.Attacker.Team == e.Player.Team {
			return
		}

		// flashing a corpse doesn't help anyone
		if !e.Player.IsAlive() {
			return
		}

		if e.FlashDuration() < 2*time.Second {
			return
		}

		id := e.Attacker.SteamID64
		pl, exists := m.Players[id]
		if !exists {
			return
		}
		pl.EffectiveFlashes++
		m.Players[id] = pl
	})
}

func nade_dmg(p demoinfocs.Parser, m *Match, s *State) {
	p.RegisterEventHandler(func(e events.PlayerHurt) {
		//magic numbers oh no....
		if e.Player == nil {
			return
		}

		if e.Attacker == nil {
			log.Println("Something is wrong with PlayerNade")
			return
		}

		if e.Weapon == nil {
			return
		}

		if e.Weapon.Type == 506 {
			att_id := e.Attacker.SteamID64

			player, exists := m.Players[att_id]
			if !exists {
				return
			}

			player.NadeDmg += e.HealthDamageTaken
			m.Players[att_id] = player
		}

		if e.Weapon.Type == 503 {
			att_id := e.Attacker.SteamID64

			player, exists := m.Players[att_id]

			if !exists {
				return
			}

			player.InfernoDmg += e.HealthDamageTaken
			m.Players[att_id] = player
		}

		if e.Weapon.Type == 502 {
			att_id := e.Attacker.SteamID64

			player, exists := m.Players[att_id]

			if !exists {
				return
			}

			player.InfernoDmg += e.HealthDamageTaken
			m.Players[att_id] = player
		}
	})
}
