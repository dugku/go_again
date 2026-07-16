package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

	//config := demoinfocs.ParserConfig{
	//gnorePacketEntitiesPanic: true,
	//}

	p := demoinfocs.NewParser(f)

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
				if s.SkippedFrames == 0 {
					log.Printf("first packet-entities failure: %v", err)
				}
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
	splitted_string := strings.Split(demPath, "/")

	match_id := splitted_string[4]

	temp := strings.Split(splitted_string[5], "-")
	temp2 := strings.Split(temp[len(temp)-1], ".")
	map_match := temp2[0]

	full := match_id + "-" + map_match

	return full
}

func get_files(rootPath string) []string {
	stringPaths := make([]string, 0)
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".dem") {
			stringPaths = append(stringPaths, path)
			backOne := filepath.Dir(path)
			full := filepath.Join(backOne, "data")
			err2 := os.Mkdir(full, 0755)
			if err2 != nil {
				log.Printf("Failed to create data dir at %s: %v", full, err2)
			}
		}
		return nil
	})

	if err != nil {
		log.Println(err)
	}

	return stringPaths
}

func main() {
	//demPath := "/mnt/c/Users/Mike/git_dirs/go_again/lag-vs-overtake-sector-m1-mirage.dem"
	dir_path := "/mnt/e/dems"

	pa := get_files(dir_path)
	for _, p := range pa {
		fmt.Printf("Parsing %s\n", p)
		m := Match{
			MatchID: match_id_from_path(p),
			Round:   make([]Rounds, 0),
		}

		err := start_parsing(p, &m)
		if err != nil {
			log.Printf("Parser Err: %v", err)
			continue
		}

		write_path_data := match_id_from_path(p)
		fmt.Println(filepath.Join(filepath.Dir(p), "data", write_path_data+".json"))
		write_err := write_json(&m, filepath.Join(filepath.Dir(p), "data", write_path_data+".json"))
		if write_err != nil {
			log.Printf("%v", write_err)
		}

		pos_write_err := write_json(&m.Positions, filepath.Join(filepath.Dir(p), "data", write_path_data+"_positions.json"))
		if pos_write_err != nil {
			log.Printf("%v", pos_write_err)
		}
	}

	//delete_dems(pa)
}
