package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// This started as a sync.Map but could revert back to a map since a keyed mutex is being used
// to restrict state reading/setting to one thread at a time
var stateMap sync.Map
var tables []GameTable = []GameTable{}

var tableMutex KeyedMutex

type KeyedMutex struct {
	mutexes sync.Map // Zero value is empty and ready for use
}

func (m *KeyedMutex) Lock(key string) func() {
	key = strings.ToLower(key)
	value, _ := m.mutexes.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()
	return func() { mtx.Unlock() }
}

func main() {
	log.Print("Starting Reversi server...")

	// Set environment flags
	UpdateLobby = os.Getenv("GO_PROD") == "1"

	if UpdateLobby {
		log.Printf("This instance will update the lobby at " + LOBBY_ENDPOINT_UPSERT)
	}

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listing on port %s", port)

	router := gin.Default()

	router.GET("/version", apiVersion)

	router.GET("/view", apiView)

	router.GET("/state", apiState)
	router.POST("/state", apiState)

	router.GET("/move/:move", apiMove)
	router.POST("/move/:move", apiMove)

	router.GET("/leave", apiLeave)
	router.POST("/leave", apiLeave)

	router.GET("/tables", apiTables)
	router.GET("/updateLobby", apiUpdateLobby)

	//	router.GET("/REFRESHLOBBY", apiRefresh)

	initializeGameServer()
	initializeTables()

	router.Run(":" + port)
}

// Api Request steps
// 1. Get state
// 2. Game Logic
// 3. Save state
// 4. Return client centric state

// request pattern
// 1. get state (locks the state)
//   A. Start a function that updates table state
//   B. Defer unlocking the state until the current "state updating" function is complete
//   C. If state is not nil, perform logic
// 2. Serialize and return results

// Executes a move for the client player, if that player is currently active
func apiMove(c *gin.Context) {

	state, unlock := getState(c)
	func() {
		defer unlock()

		if state != nil {
			// Access check - only move if the client is the active player
			if state.clientPlayer == state.ActivePlayer {
				move := strings.ToUpper(c.Param("move"))
				state.performMove(move)
				saveState(state)
				state = state.createClientState()
			}
		}
	}()

	serializeResults(c, state)
}

// Steps forward and returns the updated state
func apiState(c *gin.Context) {
	//hash := c.Query("hash")
	log.Printf("apiState 1")
	state, unlock := getState(c)
	log.Printf("apiState 2")

	func() {
		log.Printf("apiState 3")

		defer unlock()

			log.Printf("apiState 4")

		if state != nil {
				log.Printf("apiState 5")

			if state.clientPlayer >= 0 {
					log.Printf("apiState 6")

				state.runGameLogic()
					log.Printf("apiState 7")

				saveState(state)
					log.Printf("apiState 8")

			}
				log.Printf("apiState 9")

			state = state.createClientState()
				log.Printf("apiState 10")

		}
	}()

		log.Printf("apiState 11")

	// Check if passed in hash matches the state
	//if state != nil && len(hash) > 0 && hash == state.hash {
	//	serializeResults(c, "1")
	//	return
	//}
	log.Printf("apiState 12")

	state.Board_str = board_to_string(state)
		log.Printf("apiState 13")

	state.ValidMoves = valid_moves(state)
		log.Printf("apiState 14")

	display_board(state)
	log.Printf("apiState 15")

	serializeResults(c, state)
		log.Printf("apiState 16")

}

// Drop from the specified table
func apiLeave(c *gin.Context) {
	state, unlock := getState(c)

	func() {
		defer unlock()

		if state != nil {
			if state.clientPlayer >= 0 {
				state.clientLeave()
				state.updateLobby()
				saveState(state)
			}
		}
	}()
	serializeResults(c, "bye")
}

// Return the api and server version
func apiVersion(c *gin.Context) {
	versionOutput := []VersionTable{}

	//versionOutput := append(versionOutput, ["Api": "0.0", "Server":"0.0"])

	serializeResults(c, versionOutput)
}

// Returns a view of the current state without causing it to change. For debugging side-by-side with a client
func apiView(c *gin.Context) {

	state, unlock := getState(c)
	func() {
		defer unlock()

		if state != nil {
			state = state.createClientState()
		}
	}()

	serializeResults(c, state)
}

// Returns a list of real tables with player/slots for the client
// If passing "dev=1", will return developer testing tables instead of the live tables
func apiTables(c *gin.Context) {
	returnDevTables := c.Query("dev") == "1"

	tableOutput := []GameTable{}
	for _, table := range tables {
		value, ok := stateMap.Load(table.Table)
		if ok {
			state := value.(*GameState)
			if (returnDevTables && !state.registerLobby) || (!returnDevTables && state.registerLobby) {
				humanPlayerSlots, humanPlayerCount := state.getHumanPlayerCountInfo()
				table.CurPlayers = humanPlayerCount
				table.MaxPlayers = humanPlayerSlots
				tableOutput = append(tableOutput, table)
			}
		}
	}
	serializeResults(c, tableOutput)
}

// Forces an update of all tables to the lobby - useful for adhoc use if the Lobby restarts or loses info
func apiUpdateLobby(c *gin.Context) {
	for _, table := range tables {
		value, ok := stateMap.Load(table.Table)
		if ok {
			state := value.(*GameState)
			state.updateLobby()
		}
	}

	serializeResults(c, "Lobby Updated")
}

// Gets the current game state for the specified table and adds the player id of the client to it
func getState(c *gin.Context) (*GameState, func()) {

	log.Printf("getState")

	table := c.Query("table")

	if table == "" {
		table = "default"
	}
	table = strings.ToLower(table)
	player := c.Query("player")

	log.Printf("state table: '" + table + "' player: '" + player + "'")

	// Lock by the table so to avoid multiple threads updating the same table state
	unlock := tableMutex.Lock(table)

	// Load state
	value, ok := stateMap.Load(table)

	var state *GameState

	if ok {
		stateCopy := *value.(*GameState)
		state = &stateCopy
		state.setClientPlayerByName(player)
	}

	return state, unlock
}

func saveState(state *GameState) {
	stateMap.Store(state.table, state)
}

func initializeTables() {

	// Create the real servers (hard coded for now)
	createTable("The Green Door", "green", 0, true)
	createTable("The Enlighted", "enlighted", 0, true)
	createTable("Bot Room A - 2 bots", "bot2a", 2, true)
	createTable("Bot Room B - 2 bots", "bot2b", 2, true)
	createTable("Bot Room C - 2 bots", "bot2 ", 2, true)

	// For client developers, create hidden tables for each # of bots (for ease of testing with a specific # of players in the game)
	// These will not update the lobby

	for i := 1; i < 8; i++ {
		createTable(fmt.Sprintf("Dev Room - %d bots", i), fmt.Sprintf("dev%d", i), i, false)
	}

}

func createTable(serverName string, table string, botCount int, registerLobby bool) {
	state := createGameState(botCount, registerLobby)
	state.table = table
	state.serverName = serverName
	saveState(state)
	state.updateLobby()

	tables = append([]GameTable{{Table: table, Name: serverName}}, tables...)

	if UpdateLobby {
		time.Sleep(time.Millisecond * time.Duration(100))
	}
}
