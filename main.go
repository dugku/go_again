package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
)

func start_parsing(filePath string, m *Match) error {
	s := State{}
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open demo %s: %w", filePath, err)
	}
	defer f.Close()

	config := demoinfocs.ParserConfig{
		IgnorePacketEntitiesPanic: true,
	}

	p := demoinfocs.NewParserWithConfig(f, config)

	p.RegisterNetMessageHandler(func(info *msg.CSVCMsg_ServerInfo) {
		m.MapName = info.GetMapName()
		log.Printf("ServerInfo — map=%s tickInterval=%.6f maxClients=%d",
			info.GetMapName(), info.GetTickInterval(), int(info.GetMaxClients()))
	})
	state(m, &s, p)
	match_started(m, &s, p)
	econ_controller(m, &s, p)
	team_switch(p, m, &s)
	team_scoring(m, &s, p)
	round_end_controller(m, &s, p)
	player_getter(m, &s, p)
	kill_controller(m, &s, p)
	bom_planted(p, m, &s)
	flash_logic(p, m, &s)
	nade_dmg(p, m, &s)
	get_pres_round_kills(m, &s, p)

	for {
		more, err := p.ParseNextFrame()
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, demoinfocs.ErrUnexpectedEndOfDemo) {
			if strings.Contains(err.Error(), "packet entities") {
				s.SkippedFrames++
				continue
			}
			return fmt.Errorf("parse %s: %w", filePath, err)
		}
		if !more || errors.Is(err, io.EOF) {
			break
		}

		gs := p.GameState()
		if gs == nil || !s.RoundOngoing {
			continue
		}

		if gs == nil || !s.RoundOngoing || gs.IsFreezetimePeriod() {
			continue
		}

		tick := gs.IngameTick()
		if tick-s.LastCapturedTick < 16 {
			continue
		}
		s.LastCapturedTick = tick

		capture_positions(m, &s, p)
	}

	if s.SkippedFrames > 0 {
		log.Printf("WARNING: skipped %d problematic frames — positional data may have gaps", s.SkippedFrames)
	}

	return nil
}

func write_json(v any, path string) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func match_id_from_path(demPath string) string {
	base := filepath.Base(demPath)                      // "eyeballers-vs-faze-m1-ancient.dem"
	return strings.TrimSuffix(base, filepath.Ext(base)) // "eyeballers-vs-faze-m1-ancient"
}

func main() {
	demPath := "/Users/uggh/Desktop/go_again/eyeballers-vs-faze-m2-mirage.dem"

	m := Match{
		MatchID: match_id_from_path(demPath),
		Round:   make([]Rounds, 0),
	}

	err := start_parsing(demPath, &m)
	if err != nil {
		log.Printf("Parser Error: %v", err)
	}

	err3 := write_json(&m, "./out_json/thing.json")
	if err3 != nil {
		log.Printf("%v", err3)
	}

	err2 := write_json(m.Positions, "./out_json/positions.json")
	if err != nil {
		log.Printf("%v", err2)
	}
}
