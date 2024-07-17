package main

import (
	"fmt"
	"log"

	//"math"
	//"math/rand"
	//"sort"
	"strconv"
	"strings"
	"time"

	//"github.com/cardrank/cardrank"
	"github.com/mitchellh/hashstructure/v2"
	"golang.org/x/exp/slices"
)

const MAX_PLAYERS = 2
const BOARD_SIZE = 8
const WAITING_TURN = 0
const FIRST_TURN = 1
const ENDING_TURN = 64
const ANONYMOUS_CLIENT = -1

const MOVE_TIME_GRACE_SECONDS = 4
const BOT_TIME_LIMIT = time.Second * time.Duration(3)
const PLAYER_TIME_LIMIT = time.Second * time.Duration(39)
const ENDGAME_TIME_LIMIT = time.Second * time.Duration(12)
const NEW_ROUND_FIRST_PLAYER_BUFFER = time.Second * time.Duration(1)

// Drop players who do not make a move in 5 minutes
const PLAYER_PING_TIMEOUT = time.Minute * time.Duration(-5)

const WAITING_MESSAGE = "Waiting for more players"

var moveLookup = map[string]string{
	"MV": "MOVE",
}

var botNames = []string{"Hal", "Rusty", "Prime", "Torque", "Spark", "Volt", "Robo", "Data"}

var botCounter = 0
var numBots = len(botNames)

// For simplicity on the 8bit side (using switch statement), using a single character for each key.
// DOUBLE CHECK that letter isn't already in use on the object!
// Double characters are used for the list objects (validMoves and players)

type validMove struct {
	Move string `json:"m"`
	Name string `json:"n"`
}

// Move represents a move with row and column
type Move struct {
	Row, Col int
}

type Status int64

const (
	STATUS_WAITING Status = 0
	STATUS_PLAYING Status = 1
	STATUS_LEFT    Status = 3
)

type Cell int64

const (
	CELL_EMPTY Cell = 0
	CELL_BLACK Cell = 1
	CELL_WHITE Cell = 2
)

type board_cell struct {
	cell Cell
}

type Player struct {
	Name   string `json:"n"`
	Status Status `json:"s"`
	Move   string `json:"m"`
	Color  string `json:"c"`

	// Internal
	isBot    bool
	color    Cell
	lastPing time.Time
}

type GameState struct {
	// External (JSON)
	LastResult   string      `json:"l"`
	Turn         int         `json:"t"`
	ActivePlayer int         `json:"a"`
	MoveTime     int         `json:"m"`
	Viewing      int         `json:"v"`
	ValidMoves   []validMove `json:"vm"`
	Players      []Player    `json:"pl"`
	Board_str    string      `json:"bd"`

	// Internal
	board         []board_cell
	gameOver      bool
	clientPlayer  int
	table         string
	moveExpires   time.Time
	serverName    string
	registerLobby bool
	hash          string //   `json:"z"` // external later
}

// Used to send a list of available tables
type GameTable struct {
	Table      string `json:"t"`
	Name       string `json:"n"`
	CurPlayers int    `json:"p"`
	MaxPlayers int    `json:"m"`
}

// Directions for capturing pieces
var directions = [BOARD_SIZE][2]int{
	{0, 1}, {1, 0}, {0, -1}, {-1, 0},
	{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
}

var lastTurn = -1

func initializeGameServer() {

	// Append BOT to botNames array
	for i := 0; i < len(botNames); i++ {
		botNames[i] = botNames[i] + " BOT"
	}
}

func createGameState(playerCount int, registerLobby bool) *GameState {

	board := []board_cell{}

	for cell := 0; cell < 64; cell++ {
		bc := board_cell{cell: CELL_EMPTY}
		board = append(board, bc)
	}

	state := GameState{}
	state.board = board
	state.Board_str = board_to_string(&state)
	state.Turn = WAITING_TURN
	state.ActivePlayer = -1
	state.registerLobby = registerLobby

	// Pre-populate player pool with bots
	for i := 0; i < playerCount; i++ {
		if botCounter > numBots {
			botCounter = 0
		}
		state.addPlayer(botNames[botCounter], true)
		botCounter++
	}

	log.Printf("Players Count: %d\n", playerCount)

	if playerCount < 2 {
		log.Printf("createGameState - waiting message\n")
		state.LastResult = WAITING_MESSAGE
	} else {
		state.LastResult = ""
	}

	return &state
}

func board_to_string(state *GameState) string {
	board_str := ""

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

func display_board(state *GameState) {

	board := "\nBoard:\n 12345678\n"
	pos := 0

	for y := 0; y < BOARD_SIZE; y++ {
		board = board + strconv.Itoa(y+1)
		for x := 0; x < BOARD_SIZE; x++ {
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

func display_moves(state *GameState, validMoves []validMove) {

	if state.Players[state.ActivePlayer].color == CELL_BLACK {
		log.Printf("\nMoves for Black\n")
	} else {
		log.Printf("\nMoves for White\n")
	}

	board := "\n 12345678\n"
	pos := 0

	for y := 0; y < BOARD_SIZE; y++ {
		board = board + strconv.Itoa(y+1)
		for x := 0; x < BOARD_SIZE; x++ {
			board_piece := "."

			if state.board[pos].cell == CELL_BLACK {
				board_piece = "B"
			}

			if state.board[pos].cell == CELL_WHITE {
				board_piece = "W"
			}

			if state.board[pos].cell == CELL_EMPTY {
				for i := 0; i < len(validMoves); i++ {
					tempPos, e := strconv.Atoi(validMoves[i].Move)
					if e != nil {
						log.Printf("Conversion error")
					}
					if tempPos == pos {
						board_piece = strconv.Itoa(i)
					}
				}
			}

			board += board_piece
			pos += 1
		}
		board = board + "\n"
	}

	log.Printf("%s 12345678", board)

}

// ApplyMove applies a move to the board
func (state *GameState) ApplyMove(row int, col int, player_color Cell) {
	log.Printf("ApplyMove\n")
	var opponent_color = CELL_BLACK

	state.board[row*BOARD_SIZE+col].cell = player_color
	if player_color == CELL_BLACK {
		opponent_color = CELL_WHITE
	}

	for _, dir := range directions {
		r, c := row+dir[0], col+dir[1]
		toFlip := []Move{}
		for r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && state.board[r*BOARD_SIZE+c].cell == opponent_color {
			toFlip = append(toFlip, Move{r, c})
			r += dir[0]
			c += dir[1]
		}
		if r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && state.board[r*BOARD_SIZE+c].cell == player_color {
			for _, flip := range toFlip {
				state.board[flip.Row*BOARD_SIZE+flip.Col].cell = player_color
			}
		}
	}
	log.Printf("In applymove\n")
	display_board(state)
	state.Board_str = board_to_string(state)
	log.Printf("Leaving applymove\n")
}

func (state *GameState) newGame() {

	state.gameOver = false
	lastTurn = -1

	bc := board_cell{cell: CELL_EMPTY}
	for cell := 0; cell < 64; cell++ {
		state.board[cell] = bc
	}
	state.board[3+3*BOARD_SIZE].cell = CELL_BLACK
	state.board[4+3*BOARD_SIZE].cell = CELL_WHITE
	state.board[3+4*BOARD_SIZE].cell = CELL_WHITE
	state.board[4+4*BOARD_SIZE].cell = CELL_BLACK

	state.Board_str = board_to_string(state)

	if state.ActivePlayer != -1 {
		player := &state.Players[state.ActivePlayer]
		if player.color == CELL_BLACK {
			player.color = CELL_WHITE
		} else {
			player.color = CELL_BLACK
		}

		for i := 0; i < len(state.Players); i++ {
			if i != state.ActivePlayer {
				if player.color == CELL_BLACK {
					state.Players[state.ActivePlayer].color = CELL_WHITE
				} else {
					state.Players[state.ActivePlayer].color = CELL_BLACK
				}
			}
		}
	} else {
		state.Players[0].color = CELL_BLACK
		state.Players[1].color = CELL_WHITE
	}

	for i := 0; i < len(state.Players); i++ {
		player := &state.Players[i]
		player.Status = STATUS_PLAYING
		if player.color == CELL_BLACK {
			state.ActivePlayer = i
			player.Color = "B"
		} else {
			player.Color = "W"
		}
	}

	state.LastResult = ""
	state.Turn = FIRST_TURN
	log.Printf("Active Player is %d\n", state.ActivePlayer)

}

func (state *GameState) newTurn() {

	// Drop any players that left last turn
	state.dropInactivePlayers(true, false)

	// Check if multiple players are still playing
	if state.Turn > WAITING_TURN {
		playersLeft := 0
		for _, player := range state.Players {
			if player.Status == STATUS_PLAYING {
				playersLeft++
			}
		}

		if playersLeft < 2 {
			log.Printf("playersLeft < 2 %d\n", playersLeft)
			state.endGame(false)
			return
		}
	} else {
		if len(state.Players) < 2 {
			log.Printf("TURN == WAITING_TURN && %d\n", len(state.Players))
			return
		}
	}

	log.Printf("len(state.Players): %d\n", len(state.Players))

	log.Printf("newTurn turn: %d\n", state.Turn)

	// First turn of a new game?
	if state.Turn == FIRST_TURN {

		if state.LastResult == WAITING_MESSAGE {
			state.LastResult = ""
		}
	}

	if lastTurn != state.Turn {
		state.resetPlayerTimer(true)
	}
	lastTurn = state.Turn

}

func (state *GameState) addPlayer(playerName string, isBot bool) {

	newPlayer := Player{
		Name:   playerName,
		Status: 0,
		isBot:  isBot,
	}

	state.Players = append(state.Players, newPlayer)
}

func (state *GameState) setClientPlayerByName(playerName string) {
	// If no player name was passed, simply return. This is an anonymous viewer.
	if len(playerName) == 0 {
		log.Printf("No name so Anonymous Client")
		state.clientPlayer = ANONYMOUS_CLIENT
		return
	}

	state.clientPlayer = slices.IndexFunc(state.Players, func(p Player) bool { return strings.EqualFold(p.Name, playerName) })

	if state.clientPlayer == -1 {
		log.Printf("%s - was not found\n", playerName)
	} else {
		log.Printf("%s - already exists clientPlayer:%d\n", playerName, state.clientPlayer)
	}
	// If a new player is joining, remove any old players that timed out to make space
	if state.clientPlayer <= ANONYMOUS_CLIENT {
		// Drop any players that left to make space
		state.dropInactivePlayers(false, true)
	}

	if state.clientPlayer == ANONYMOUS_CLIENT {
		log.Printf("%s will be added as a client if room\n", playerName)
	}

	// Add new player if there is room
	if state.clientPlayer <= ANONYMOUS_CLIENT &&
		len(state.Players) < MAX_PLAYERS {

		state.addPlayer(playerName, false)
		state.clientPlayer = len(state.Players) - 1
		log.Printf("There are now %d players, New player %s as client %d\n", len(state.Players), playerName, state.clientPlayer)

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

	// Hand rank details
	// Rank: SF, 4K, FH, F, S, 3K, 2P, 1P, HC

	log.Printf("Ending Game!\n")
	state.gameOver = true
	state.ActivePlayer = -1
	state.Turn = ENDING_TURN

	remainingPlayers := []int{}

	for index, player := range state.Players {
		if !abortGame && player.Status == STATUS_PLAYING {
			remainingPlayers = append(remainingPlayers, index)
		}

	}

	state.LastResult = ""

	state.moveExpires = time.Now().Add(ENDGAME_TIME_LIMIT)

}

// ********************************************************************************
// Emulates simplified player/logic for REVERSI
// ********************************************************************************
func (state *GameState) runGameLogic() {
	log.Printf("****runGameLogic****\n")
	state.playerPing()

	// We can't play a game until there are at least 2 players
	if len(state.Players) < 2 {
		// Reset the Turn to 0 so the client knows there is no active game being run
		state.Turn = WAITING_TURN
		state.ActivePlayer = -1
		return
	}

	// Very first call of state? Initialize first turn but do not play for any BOTs
	if state.Turn == WAITING_TURN {
		log.Printf("runGameLogic newGame\n")
		state.newGame()
		return
	}

	//isHumanPlayer := state.ActivePlayer == state.clientPlayer

	if state.gameOver {

		// Create a new game if the end game delay is past
		if int(time.Until(state.moveExpires).Seconds()) < 0 {
			state.dropInactivePlayers(false, false)
			state.Turn = WAITING_TURN
			state.gameOver = false
			state.newGame()
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
		log.Printf("Players Left == 1\n")
		state.endGame(false)
		return
	}

	// Check if we should start the next game. One of the following must be true

	if state.ActivePlayer > -1 {
		if state.Turn == ENDING_TURN {
			log.Printf("ActivePlayer >  -1\n")
			state.endGame(false)
		} else {
			log.Printf("runGameLogic newTurn ActivePlayer > -1\n")
			state.newTurn()
		}
		return
	}

	// Return if the move timer has not expired
	// Check timer if no active player, or the active player hasn't already left
	if state.ActivePlayer == -1 || state.Players[state.ActivePlayer].Status != STATUS_LEFT {
		moveTimeRemaining := int(time.Until(state.moveExpires).Seconds())

		if moveTimeRemaining > 0 {
			log.Printf("runGameLogic moveTimeRemaining\n")
			return
		}
	}

	// If there is no active player, we are done
	if state.ActivePlayer < 0 {
		log.Printf("runGameLogic ActivePlayer < 0\n")
		return
	}

	// Force a move for this player or BOT if they are in the game
	if state.Players[state.ActivePlayer].Status == STATUS_PLAYING {
		moves := state.getValidMoves()

		// If this is a bot, pick the best move using some simple logic (sometimes random)
		if state.Players[state.ActivePlayer].isBot {
			log.Printf("Active Player isBot\n")
			if len(moves) > 0 {
				log.Printf("Forcing Move")
				state.performMove(moves[0].Move)
			}
		}

		if !state.Players[state.ActivePlayer].isBot {
			log.Printf("Active Player isHuman\n")
			if len(moves) > 0 {
				log.Printf("Forcing Move")
				state.performMove(moves[0].Move)
			}
		}

	}

}

// Drop players that left or have not pinged within the expected timeout
func (state *GameState) dropInactivePlayers(inMiddleOfGame bool, dropForNewPlayer bool) {
	cutoff := time.Now().Add(PLAYER_PING_TIMEOUT)
	players := []Player{}
	currentPlayerName := ""
	if state.clientPlayer > ANONYMOUS_CLIENT {
		currentPlayerName = state.Players[state.clientPlayer].Name
	}

	for _, player := range state.Players {
		if len(state.Players) > 0 && player.Status != STATUS_LEFT && (inMiddleOfGame || player.isBot || player.lastPing.Compare(cutoff) > 0) {
			players = append(players, player)
		}
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
		log.Printf("dropInactivePlayers - waiting message\n")
		state.LastResult = WAITING_MESSAGE
	}

	// If any player state changed, update the lobby
	if playersWereDropped {
		state.updateLobby()
	}

}

func (state *GameState) clientLeave() {
	if state.clientPlayer <= ANONYMOUS_CLIENT {
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
		log.Printf("Players left == 0\n")
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

	log.Printf("Perform Move '%s'\n", move)

	if len(internalCall) == 0 || !internalCall[0] {
		state.playerPing()
	}

	// Get pointer to player
	player := &state.Players[state.ActivePlayer]

	// Sanity check if player is still in the game. Unless there is a bug, they should never be active if their status is != PLAYING
	if player.Status != STATUS_PLAYING {
		log.Printf("Not Playing")
		return false
	}

	// Only perform move if it is a valid move for this player
	//if !slices.ContainsFunc(state.getValidMoves(), func(m validMove) bool { return m.Move == move }) {
	//	log.Printf("No Valid Move")
	//	return false
	//}

	board_pos, err := strconv.Atoi(move)
	if err != nil {
		log.Printf("error converting '%s' to int\n", move)
		return false
	}
	row := board_pos / BOARD_SIZE
	col := board_pos - (row * BOARD_SIZE)

	state.ApplyMove(row, col, player.color)
	player.Move = moveLookup[move]
	log.Printf("Going to next player\n")
	state.Turn++
	state.nextValidPlayer()

	return true
}

func (state *GameState) resetPlayerTimer(newRound bool) {
	timeLimit := PLAYER_TIME_LIMIT

	if state.Players[state.ActivePlayer].isBot {
		timeLimit = BOT_TIME_LIMIT
	}

	if newRound {
		timeLimit += NEW_ROUND_FIRST_PLAYER_BUFFER
	}

	state.moveExpires = time.Now().Add(timeLimit)
}

func (state *GameState) nextValidPlayer() {
	log.Printf("CurrentPlayer: %d\n", state.ActivePlayer)
	// Move to next player
	state.ActivePlayer = (state.ActivePlayer + 1) % len(state.Players)

	// Skip over player if not in this game (joined late / folded)
	for state.Players[state.ActivePlayer].Status != STATUS_PLAYING {
		state.ActivePlayer = (state.ActivePlayer + 1) % len(state.Players)
	}
	state.resetPlayerTimer(false)
	log.Printf("NewPlayer: %d\n", state.ActivePlayer)
}

/***********************************************
 * Calculates which squares are valid moves    *
 * for player. Valid moves are recorded in the *
 * moves array                                 *
 **********************************************/

func (state *GameState) getValidMoves() []validMove {
	var rowdelta int = 0 /* Row increment around a square    */
	var coldelta int = 0 /* Column increment around a square */
	var x int = 0        /* Row index when searching         */
	var y int = 0        /* Column index when searching      */
	var player_color Cell = CELL_BLACK
	var opponent_color Cell = CELL_WHITE
	var move validMove
	var moves []validMove

	if state.ActivePlayer == -1 {
		return moves
	}

	if state.Players[state.ActivePlayer].color == CELL_BLACK {
		log.Printf("Active Player (%s) is Black", state.Players[state.ActivePlayer].Name)
		player_color = CELL_BLACK
		opponent_color = CELL_WHITE
	} else {
		log.Printf("Active Player (%s) is White", state.Players[state.ActivePlayer].Name)
		player_color = CELL_WHITE
		opponent_color = CELL_BLACK
	}

	/* Find squares for valid moves.                           */
	/* A valid move must be on a blank square and must enclose */
	/* at least one opponent square between two player squares */
	for row := 0; row < BOARD_SIZE; row++ {
		for col := 0; col < BOARD_SIZE; col++ {

			if state.board[row*BOARD_SIZE+col].cell != CELL_EMPTY { /* Is it a blank square?  */
				continue /* No - so on to the next */
			}

			/* Check all the squares around the blank square  */
			/* for the opponents counter                      */
			for rowdelta = -1; rowdelta <= 1; rowdelta++ {
				for coldelta = -1; coldelta <= 1; coldelta++ {
					/* Don't check outside the array, or the current square */
					if row+rowdelta < 0 || row+rowdelta >= BOARD_SIZE ||
						col+coldelta < 0 || col+coldelta >= BOARD_SIZE ||
						(rowdelta == 0 && coldelta == 0) {
						continue
					}

					/* Now check the square */
					if state.board[(row+rowdelta)*BOARD_SIZE+col+coldelta].cell == opponent_color {
						/* If we find the opponent, move in the delta direction  */
						/* over opponent counters searching for a player counter */
						x = row + rowdelta /* Move to          */
						y = col + coldelta /* opponent square  */

						/* Look for a player square in the delta direction */
						for {
							y += rowdelta /* Go to next square */
							x += coldelta /* in delta direction*/

							/* If we move outside the array, give up */
							if x < 0 || x >= BOARD_SIZE || y < 0 || y >= BOARD_SIZE {
								break
							}
							/* If we find a blank square, give up */
							if state.board[x+y*BOARD_SIZE].cell == CELL_EMPTY {
								break
							}

							/*  If the square has a player counter */
							/*  then we have a valid move          */
							if state.board[x+y*BOARD_SIZE].cell == player_color {
								log.Printf("valid move: %d [%d,%d]\n", row*BOARD_SIZE+col, row+1, col+1)

								move.Move = strconv.Itoa(row*BOARD_SIZE + col)
								move.Name = strconv.Itoa(row+1) + "," + strconv.Itoa(col+1)
								moves = append(moves, move) /* Mark as valid */
								break                       /* Go check another square    */
							}
						}
					} // if
				} // for coldelta
			} // for rowdelta
		} // for col
	} // for row

	display_moves(state, moves)
	return moves
}

// Creates a copy of the state and modifies it to be from the
// perspective of this client (e.g. player array, visible cards)
func (state *GameState) createClientState() *GameState {

	// Check if:
	// 1. The game is over,

	if state.gameOver ||
		len(state.Players) < 2 {
		state.ActivePlayer = -1
	}

	// When on observer is viewing the game, the clientPlayer will be -1, so just start at 0
	// Also, set flag to let client know they are not actively part of the game
	start := state.clientPlayer
	if start < 0 {
		start = 0
		state.Viewing = 1
	} else {
		state.Viewing = 0
	}

	// Determine valid moves for this player (if their turn)
	//if state.ActivePlayer == 0 {

	state.ValidMoves = state.getValidMoves()

	//}

	// Determine the move time left. Reduce the number by the grace period, to allow for plenty of time for a response to be sent back and accepted
	state.MoveTime = int(time.Until(state.moveExpires).Seconds())

	if state.ActivePlayer > -1 {
		log.Println("Reducing timer")
		state.MoveTime -= MOVE_TIME_GRACE_SECONDS
		log.Printf("Time = %d  ActivePlayer = %d\n", state.MoveTime, state.ActivePlayer)
	}

	// No need to send move time if the calling player isn't the active player
	if state.MoveTime < 0 || state.ActivePlayer != 0 {
		log.Printf("Time adjusted to zero")
		state.MoveTime = 0
	}

	if state.ActivePlayer == 0 {
		if len(state.ValidMoves) == 0 {
			log.Printf("ActivePlayer No moves")
			state.nextValidPlayer()
		} else {
			log.Printf("ActivePlayer Force Moved: %s", state.ValidMoves[0].Name)
			state.performMove(state.ValidMoves[0].Move)
		}
	} else {
		log.Printf("Not an active player")
		if (state.ActivePlayer >= 0) && (state.MoveTime == 0) {

			if len(state.ValidMoves) > 0 {
				log.Printf("ActivePlayer Force Moved: %s", state.ValidMoves[0].Name)
				state.performMove(state.ValidMoves[0].Move)
			}
		}

	}
	// Compute hash - this will be compared with an incoming hash. If the same, the entire state does not
	// need to be sent back. This speeds up checks for change in state
	state.hash = "0"
	hash, _ := hashstructure.Hash(state, hashstructure.FormatV2, nil)
	state.hash = fmt.Sprintf("%d", hash)

	return state
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
	humanAvailSlots := MAX_PLAYERS
	humanPlayerCount := 0
	cutoff := time.Now().Add(PLAYER_PING_TIMEOUT)

	for _, player := range state.Players {
		if player.isBot {
			humanAvailSlots--
		} else if player.Status != STATUS_LEFT && player.lastPing.Compare(cutoff) > 0 {
			humanPlayerCount++
		}
	}
	return humanAvailSlots, humanPlayerCount
}
