package main

import (
	//"fmt"
	"log"
	//"math/rand/v2"
	"strconv"
	"strings"
	"time"

	//"github.com/mitchellh/hashstructure/v2"
	"golang.org/x/exp/slices"
	//"golang.org/x/tools/go/analysis/passes/printf"
)

const MOVE_TIME_GRACE_SECONDS = 4
const BOT_TIME_LIMIT = time.Second * time.Duration(3)
const PLAYER_TIME_LIMIT = time.Second * time.Duration(39)
const ENDGAME_TIME_LIMIT = time.Second * time.Duration(12)
const NEW_ROUND_FIRST_PLAYER_BUFFER = time.Second * time.Duration(1)

// Drop players who do not make a move in 5 minutes
const PLAYER_PING_TIMEOUT = time.Minute * time.Duration(-5)

const WAITING_MESSAGE = "Waiting for more players"

var botNames = []string{"Clyd", "Jim", "Kirk", "Hulk", "Fry", "Meg", "Grif", "GPT"}

// For simplicity on the 8bit side (using switch statement), using a single character for each key.
// DOUBLE CHECK that letter isn't already in use on the object!
// Double characters are used for the list objects (validMoves and players)

type validMove struct {
	Move string `json:"m"`
	Name string `json:"n"`
}

type Cell int64

const (
	CELL_EMPTY Cell = 0
	CELL_BLACK Cell = 1
	CELL_WHITE Cell = 2
)

type board_cell struct {
	cell Cell
}

type Status int64

const (
	STATUS_INACTIVE Status = -1
	STATUS_WAITING  Status = 0
	STATUS_PLAYING  Status = 1
	STATUS_LEFT     Status = 3
)

type Player struct {
	Name   string `json:"n"`
	Status Status `json:"s"`
	Bet    int    `json:"b"`
	Move   string `json:"m"`

	// Internal
	isBot    bool
	lastPing time.Time
}

type GameState struct {
	// External (JSON)
	LastResult   string      `json:"l"`
	Round        int         `json:"r"`
	ActivePlayer int         `json:"a"`
	MoveTime     int         `json:"m"`
	Viewing      int         `json:"v"`
	ValidMoves   []validMove `json:"vm"`
	Players      []Player    `json:"pl"`
	Board_str    string      `json:"bd"`
	//hash         string   	 `json:"z"` // external later

	// Internal
	board         []board_cell
	gameOver      bool
	clientPlayer  int
	table         string
	moveExpires   time.Time
	serverName    string
	registerLobby bool
}

// Used to send a list of available tables
type GameTable struct {
	Table      string `json:"t"`
	Name       string `json:"n"`
	CurPlayers int    `json:"p"`
	MaxPlayers int    `json:"m"`
}

// Used for indicating version of API and Server
type VersionTable struct {
	Api    string `json:"av"`
	Server string `json:"sv"`
}

func initializeGameServer() {

	// Append BOT to botNames array
	for i := 0; i < len(botNames); i++ {
		botNames[i] = botNames[i] + " BOT"
	}
}

func board_to_string(state *GameState) string {
	board_str := ""
	log.Printf("board_to_string")
	log.Printf("state.board %p", state.board)
	log.Printf("board len %d\n", len(state.board))

	for pos := 0; pos < len(state.board); pos++ {
		board_piece := "."

		if state.board[pos].cell == CELL_BLACK {
			board_piece = "B"
		}

		if state.board[pos].cell == CELL_WHITE {
			board_piece = "W"
		}

		board_str += board_piece
	}
	return board_str
}

/*

func random_board(state *GameState) {
	for pos := 0; pos < len(state.board); pos++ {
		r := rand.IntN(10)
		if r > 3 {
			state.board[pos].cell = CELL_BLACK
		}

		if r > 6 {
			state.board[pos].cell = CELL_WHITE
		}
	}
}
*/

func display_board(state *GameState) {

	//state.Board_str = board_to_string(state)
	board := "\bBoard:\n 12345678\n"
	pos := 0

	for y := 0; y < 8; y++ {
		board = board + strconv.Itoa(y+1)
		for x := 0; x < 8; x++ {
			board_piece := "."

			if state.board[pos].cell == CELL_BLACK {
				board_piece = "B"
			}

			if state.board[pos].cell == CELL_WHITE {
				board_piece = "W"
			}

			board += board_piece
			pos += 1
		}
		board = board + "\n"
	}

	log.Printf("%s 12345678", board)

}



func createGameState(playerCount int, registerLobby bool) *GameState {

	log.Printf("createGameState %d\n", playerCount)
	board := []board_cell{}

	for cell := 0; cell < 64; cell++ {
		bc := board_cell{cell: CELL_EMPTY}
		board = append(board, bc)
	}

	state := GameState{}
	state.board = board
	state.Round = 0
	state.ActivePlayer = int(CELL_BLACK)
	state.registerLobby = registerLobby
	state.Board_str = board_to_string(&state)

	// Pre-populate player pool with bots
	for i := 0; i < playerCount; i++ {
		state.addPlayer(botNames[i], true)
	}

	if playerCount < 2 {
		state.LastResult = WAITING_MESSAGE
	}

	return &state
}

func (state *GameState) newRound() {

	log.Printf("newRound\n")

	// Drop any players that left last round
	state.dropInactivePlayers(true, false)

	log.Printf("newRound 2\n")
	// Check if multiple players are still playing
	if state.Round > 0 {
		playersLeft := 0
		for _, player := range state.Players {
			if player.Status == STATUS_PLAYING {
				playersLeft++
			}
		}

		if playersLeft < 2 {
			state.endGame(false)
			return
		}
	} else {
		if len(state.Players) < 2 {
			return
		}
	}

	log.Printf("newRound 3\n")
	state.Round++

	// Clear pot at start so players can anti
	if state.Round == 1 {

		state.gameOver = false
	}

	log.Printf("newRound 4\n")
	// Reset players for this round
	for i := 0; i < len(state.Players); i++ {

		// Get pointer to player
		player := &state.Players[i]

		// Reset player's last move/bet for this round
		player.Move = ""
	}

	log.Printf("newRound 5 (%d)\n", state.Round)

	// First round of a new game? put on initial chips
	if state.Round == 0 {

		state.board[3+3*8].cell = CELL_BLACK
		state.board[4+3*8].cell = CELL_WHITE
		state.board[3+4*8].cell = CELL_WHITE
		state.board[4+4*8].cell = CELL_BLACK

		state.Board_str = board_to_string(state)

		if state.LastResult == WAITING_MESSAGE {
			state.LastResult = ""
		}
	}

	log.Printf("newRound 6\n")

	// if it's curretly black's turn, switch to white, etc.
	if state.ActivePlayer == int(CELL_BLACK) {
		state.ActivePlayer = int(CELL_WHITE)
	} else {
		state.ActivePlayer = int(CELL_BLACK)
	}

	log.Printf("newRound 7\n")
	state.resetPlayerTimer(true)

	log.Printf("newRound 8\n")
}

func (state *GameState) addPlayer(playerName string, isBot bool) {
	//log.Printf("addPlayer %s\n", playerName)

	newPlayer := Player{
		Name:   playerName,
		Status: 0,
		isBot:  isBot,
	}

	state.Players = append(state.Players, newPlayer)
}

func (state *GameState) setClientPlayerByName(playerName string) {

	log.Printf("setClientPlayer %s\n", playerName)

	// If no player name was passed, simply return. This is an anonymous viewer.
	if len(playerName) == 0 {
		state.clientPlayer = -1
		return
	}
	state.clientPlayer = slices.IndexFunc(state.Players, func(p Player) bool { return strings.EqualFold(p.Name, playerName) })

	// If a new player is joining, remove any old players that timed out to make space
	if state.clientPlayer < 0 {
		// Drop any players that left to make space
		state.dropInactivePlayers(false, true)
	}

	// Add new player if there is room
	if state.clientPlayer < 0 && len(state.Players) < 2 {
		state.addPlayer(playerName, false)
		state.clientPlayer = len(state.Players) - 1

		// Set the ping for this player so they are counted as active when updating the lobby
		state.playerPing()

		// Update the lobby with the new state (new player joined)
		state.updateLobby()
	}

	// Extra logic if a player is requesting
	if state.clientPlayer > 0 {

		// In case a player returns while they are still in the "LEFT" status (before the current game ended), add them back in as waiting
		if state.Players[state.clientPlayer].Status == STATUS_LEFT {
			state.Players[state.clientPlayer].Status = STATUS_WAITING
		}
	}
}

func (state *GameState) endGame(abortGame bool) {
	// The next request for /state will start a new game

	if abortGame {
		state.gameOver = true
		log.Printf("abort game")
		state.ActivePlayer = -1
		state.Round = 5
	}

}

// Emulates simplified player/logic for 5 card stud
func (state *GameState) runGameLogic() {
	state.playerPing()

	// We can't play a game until there are at least 2 players
	if len(state.Players) < 2 {
		// Reset the round to 0 so the client knows there is no active game being run
		state.Round = 0
		log.Printf("too few players")
		state.ActivePlayer = -1
		return
	}

	// Very first call of state? Initialize first round but do not play for any BOTs
	if state.Round == 0 {
		state.newRound()
		return
	}

	//isHumanPlayer := state.ActivePlayer == state.clientPlayer

	if state.gameOver {

		// Create a new game if the end game delay is past
		if int(time.Until(state.moveExpires).Seconds()) < 0 {
			state.dropInactivePlayers(false, false)
			state.Round = 0
			state.gameOver = false
			state.newRound()
		}
		return
	}

	// Check if only one player is left
	playersLeft := 0
	for _, player := range state.Players {
		if player.Status == STATUS_PLAYING {
			playersLeft++
		}
	}

	// If only one player is left, just end the game now
	if playersLeft == 1 {
		state.endGame(false)
		return
	}

	// If there is no active player, we are done
	if state.ActivePlayer < 0 {
		log.Printf("No active player")
		return
	}

	// Bounds check - clamp the move to the end of the array if a higher move is desired.
	// This may occur if a bot wants to call, but cannot, due to limited funds.
	//if choice > len(moves)-1 {
	//	choice = len(moves) - 1
	//}

	//move := moves[choice]

	//state.performMove(move.Move, true)

}

// Drop players that left or have not pinged within the expected timeout
func (state *GameState) dropInactivePlayers(inMiddleOfGame bool, dropForNewPlayer bool) {
	cutoff := time.Now().Add(PLAYER_PING_TIMEOUT)
	players := []Player{}
	currentPlayerName := ""
	if state.clientPlayer > -1 {
		currentPlayerName = state.Players[state.clientPlayer].Name
	}

	for _, player := range state.Players {
		if len(state.Players) > 0 && player.Status != STATUS_LEFT && (inMiddleOfGame || player.isBot || player.lastPing.Compare(cutoff) > 0) {
			players = append(players, player)
		}
	}

	// If one player is left, don't drop them within the round, let the normal game end take care of it
	if inMiddleOfGame && len(players) == 1 {
		return
	}

	// Store if players were dropped, before updating the state player array
	playersWereDropped := len(state.Players) != len(players)

	if playersWereDropped {
		state.Players = players
	}

	// If a new player is joining, don't bother updating anything else
	if dropForNewPlayer {
		return
	}

	// Update the client player index in case it changed due to players being dropped
	if len(players) > 0 {
		state.clientPlayer = slices.IndexFunc(players, func(p Player) bool { return strings.EqualFold(p.Name, currentPlayerName) })
	}

	// If only one player is left, we are waiting for more
	if len(state.Players) < 2 {
		state.LastResult = WAITING_MESSAGE
	}

	// If any player state changed, update the lobby
	if playersWereDropped {
		state.updateLobby()
	}

}

func (state *GameState) clientLeave() {
	if state.clientPlayer < 0 {
		return
	}
	player := &state.Players[state.clientPlayer]

	player.Status = STATUS_LEFT
	player.Move = "LEFT"

	// Check if no human players are playing. If so, end the game
	playersLeft := 0
	for _, player := range state.Players {
		if player.Status == STATUS_PLAYING && !player.isBot {
			playersLeft++
		}
	}

	// If the last player dropped, stop the game and update the lobby
	if playersLeft == 0 {
		state.endGame(true)
		state.dropInactivePlayers(false, false)
		return
	}
}

// Update player's ping timestamp. If a player doesn't ping in a certain amount of time, they will be dropped from the server.
func (state *GameState) playerPing() {
	state.Players[state.clientPlayer].lastPing = time.Now()
}

// Performs the requested move for the active player, and returns true if successful
func (state *GameState) performMove(move string, internalCall ...bool) bool {

	if len(internalCall) == 0 || !internalCall[0] {
		state.playerPing()
	}

	// Get pointer to player
	player := &state.Players[state.ActivePlayer]

	// Sanity check if player is still in the game. Unless there is a bug, they should never be active if their status is != PLAYING
	if player.Status != STATUS_PLAYING {
		return false
	}

	// Only perform move if it is a valid move for this player
	if !slices.ContainsFunc(state.getValidMoves(), func(m validMove) bool { return m.Move == move }) {
		return false
	}

	player.Move = ""
	state.nextValidPlayer()

	return true
}

func (state *GameState) resetPlayerTimer(newRound bool) {
	log.Printf("resetPlayerTimer 1\n")
	timeLimit := PLAYER_TIME_LIMIT

	log.Printf("resetPlayerTimer ActivePlayer: %d  - %d\n", state.ActivePlayer, len(state.Players))
	if state.Players[state.ActivePlayer].isBot {
		timeLimit = BOT_TIME_LIMIT
	}
	log.Printf("resetPlayerTimer 2\n")
	if newRound {
		log.Printf("newRound time limit")
		timeLimit += NEW_ROUND_FIRST_PLAYER_BUFFER
	}
	log.Printf("resetPlayerTimer 3\n")
	state.moveExpires = time.Now().Add(timeLimit)
	log.Printf("resetPlayerTimer 4\n")
}

func (state *GameState) nextValidPlayer() {
	// Move to next player
	state.ActivePlayer = (state.ActivePlayer + 1) % len(state.Players)

	// Skip over player if not in this game (joined late / folded)
	for state.Players[state.ActivePlayer].Status != STATUS_PLAYING {
		state.ActivePlayer = (state.ActivePlayer + 1) % len(state.Players)
	}
	state.resetPlayerTimer(false)
}

func (state *GameState) getValidMoves() []validMove {
	//moves := []validMove{}

	moves := valid_moves(state)

	return moves
}

// Creates a copy of the state and modifies it to be from the
// perspective of this client (e.g. player array, visible cards)
func (state *GameState) createClientState() *GameState {

	stateCopy := *state

	//setActivePlayer := false

	// Check if:
	// 1. The game is over,
	// 2. Only one player is left (waiting for another player to join)
	// 3. We are at the end of a round, where the active player has moved
	// This lets the client perform end of round/game tasks/animation
	if state.gameOver ||
		len(stateCopy.Players) < 2 ||
		(stateCopy.ActivePlayer > -1) {
		print("createClientState game over")
		stateCopy.ActivePlayer = -1
	}

	// Now, store a copy of state players, then loop
	// through and add to the state copy, starting
	// with this player first

	//statePlayers := stateCopy.Players
	stateCopy.Players = []Player{}

	// When on observer is viewing the game, the clientPlayer will be -1, so just start at 0
	// Also, set flag to let client know they are not actively part of the game
	start := state.clientPlayer
	if start < 0 {
		start = 0
		stateCopy.Viewing = 1
	} else {
		stateCopy.Viewing = 0
	}

	// Determine valid moves for this player (if their turn)
	if stateCopy.ActivePlayer == 0 {
		stateCopy.ValidMoves = state.getValidMoves()
	}

	// Determine the move time left. Reduce the number by the grace period, to allow for plenty of time for a response to be sent back and accepted
	stateCopy.MoveTime = int(time.Until(stateCopy.moveExpires).Seconds())

	if stateCopy.ActivePlayer > -1 {
		stateCopy.MoveTime -= MOVE_TIME_GRACE_SECONDS
	}

	// No need to send move time if the calling player isn't the active player
	if stateCopy.MoveTime < 0 || stateCopy.ActivePlayer != 0 {
		stateCopy.MoveTime = 0
	}

	// Compute hash - this will be compared with an incoming hash. If the same, the entire state does not
	// need to be sent back. This speeds up checks for change in state
	//stateCopy.hash = "0"
	//hash, _ := hashstructure.Hash(stateCopy, hashstructure.FormatV2, nil)
	//stateCopy.hash = fmt.Sprintf("%d", hash)

	return &stateCopy
}

func (state *GameState) updateLobby() {
	if !state.registerLobby {
		return
	}

	humanPlayerSlots, humanPlayerCount := state.getHumanPlayerCountInfo()

	// Send the total human slots / players to the Lobby
	sendStateToLobby(humanPlayerSlots, humanPlayerCount, true, state.serverName, "?table="+state.table)
}

// Return number of active human players in the table, for the lobby
func (state *GameState) getHumanPlayerCountInfo() (int, int) {

	log.Printf("Get human player count")

	humanAvailSlots := 8
	humanPlayerCount := 0
	cutoff := time.Now().Add(PLAYER_PING_TIMEOUT)

	for _, player := range state.Players {
		if player.isBot {
			humanAvailSlots--
		} else if player.Status != STATUS_LEFT && player.lastPing.Compare(cutoff) > 0 {
			humanPlayerCount++
		}
	}
	log.Printf("human avail %d, human players: %d\n", humanAvailSlots, humanPlayerCount)
	return humanAvailSlots, humanPlayerCount
}
