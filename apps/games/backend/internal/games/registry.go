package games

import (
	"games/internal/games/battleships"
	"games/internal/games/connect4"
	"games/internal/games/lightcycles"
	"games/internal/games/memory"
	"games/internal/games/nim"
	"games/internal/games/pong"
	"games/internal/games/tictactoe"
	"games/internal/games/wordle"
)

// Registry maps game IDs to their Game implementations.
var Registry = map[string]Game{
	"tictactoe":   &tictactoe.Game{},
	"connect4":    &connect4.Game{},
	"memory":      &memory.Game{},
	"nim":         &nim.Game{},
	"pong":        &pong.Game{},
	"lightcycles": &lightcycles.Game{},
	"wordle":      &wordle.Game{},
	"battleships": &battleships.Game{},
}

// GameMeta holds display metadata for a game.
type GameMeta struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Emoji       string `json:"emoji"`
	Color       string `json:"color"`
	RealTime    bool   `json:"real_time"`
}

// Meta maps game IDs to their display metadata.
var Meta = map[string]GameMeta{
	"tictactoe":   {ID: "tictactoe", Name: "Tic Tac Toe", Description: "Classic 3×3 grid. Place your mark, line up three to win.", Emoji: "⭕", Color: "#e91e63", RealTime: false},
	"connect4":    {ID: "connect4", Name: "Connect 4", Description: "Drop pieces into the grid. Connect four in a row to win.", Emoji: "🔴", Color: "#f44336", RealTime: false},
	"memory":      {ID: "memory", Name: "Memory Match", Description: "Flip cards and find matching pairs. Most pairs wins.", Emoji: "🃏", Color: "#9c27b0", RealTime: false},
	"nim":         {ID: "nim", Name: "Nim", Description: "Take 1-3 sticks per turn. The player who takes the LAST stick loses.", Emoji: "🪄", Color: "#ff9800", RealTime: false},
	"pong":        {ID: "pong", Name: "Pong", Description: "Classic paddle battle. Ball speeds up after every hit. First to 7 wins.", Emoji: "🏓", Color: "#2196f3", RealTime: true},
	"lightcycles": {ID: "lightcycles", Name: "Light Cycles", Description: "Leave a glowing trail. Force your opponent to crash into it.", Emoji: "⚡", Color: "#00bcd4", RealTime: true},
	"wordle":      {ID: "wordle", Name: "Wordle Duel", Description: "Guess the same 5-letter word. First to solve wins!", Emoji: "🔤", Color: "#4caf50", RealTime: false},
	"battleships": {ID: "battleships", Name: "Battleships", Description: "Place your fleet and sink the opponent's ships. Hit = shoot again!", Emoji: "⚓", Color: "#1565c0", RealTime: false},
}
