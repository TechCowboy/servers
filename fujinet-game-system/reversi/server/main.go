package main

import (
	//"fmt"
	//"crypto/rand"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"net"

	"github.com/gin-gonic/gin"
)

/*
http://localhost:8080/tables
http://localhost:8080/state?table=bot2c&player=TechCowboy

*/
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
	log.Print("Starting reversi server...")

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

	ifaces, err := net.Interfaces()
	// handle err
	if err != nil {

	}
	
	for _, i := range ifaces {
    	addrs, err := i.Addrs()
		// handle err
		if err != nil {

		}
    	
    	for _, addr := range addrs {
        	var ip net.IP
        	switch v := addr.(type) {
        		case *net.IPNet:
                	ip = v.IP
        		case *net.IPAddr:
                	ip = v.IP
        	}
        // process IP address
		log.Printf("%s\n", ip.String())
    }
}

	router.Run(":" + port)
}

// Api Request steps
// 1. Get state
// 2. Game Logic
// 3. Save state
// 4. Return client centric state

// request pattern
// 1. get state (locks the state)
//   A. Start a function that updates table statef
//   B. Defer unlocking the state until the current "state updating" function is complete
//   C. If state is not nil, perform logic
// 2. Serialize and return results

// Executes a move for the client player, if that player is currently active



func apiMove(c *gin.Context) {

	log.Printf("********************************************************\n")
	log.Printf("***********************apiMove**************************\n")
	log.Printf("********************************************************\n")

	state, unlock := getState(c)
	func() {
		defer unlock()

		if state != nil {
			for i := 0; i < 2; i++ {
				player := state.Players[i]
				color := "BLACK"
				if player.color == CELL_WHITE {
					color = "WHITE"
				}
				log.Printf("player[%d]: %s %s\n", i, player.Name, color)

			}
			log.Printf("clientPlayer:%d  ActivePlayer: %d\n", state.clientPlayer, state.ActivePlayer)
			// Access check - only move if the client is the active player
			if state.clientPlayer == state.ActivePlayer {
				move := strings.ToUpper(c.Param("move"))
				log.Printf("MOVE: '%s'\n", move)
				requested_move := move[2:]

				requested_move = requested_move[:len(requested_move)-1]

				if state.isValidMove(requested_move) {
					state.performMove(requested_move)
					state.nextValidPlayer()
				} else {
					log.Printf("Not a Valid Move\n")
				}
				saveState(state)
				state = state.createClientState()
			} else {
				log.Printf("MOVE: Not active player\n")
			}
		}
	}()

	serializeResults(c, state)
}

// Steps forward and returns the updated state
func apiState(c *gin.Context) {
	hash := c.Query("hash")
	state, unlock := getState(c)
	log.Println("apiState")

	func() {
		defer unlock()

		if state != nil {

			if state.clientPlayer > ANONYMOUS_CLIENT {

				state.runGameLogic()
				saveState(state)
			}

			state = state.createClientState()
		}
	}()

	// Check if passed in hash matches the state
	if state != nil && len(hash) > 0 && hash == state.hash {
		serializeResults(c, "1")
		return
	}

	serializeResults(c, state)
}

// Drop from the specified table
func apiLeave(c *gin.Context) {
	state, unlock := getState(c)

	func() {
		defer unlock()

		if state != nil {
			if state.clientPlayer > ANONYMOUS_CLIENT {
				state.clientLeave()
				state.updateLobby()
				saveState(state)
			}
		}
	}()
	serializeResults(c, "bye")
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
	table := c.Query("table")

	if table == "" {
		table = "default"
	}
	table = strings.ToLower(table)
	player := c.Query("player")

	log.Printf("getState table:'%s' player:'%s'\n", table, player)
	// Lock by the table so to avoid multiple threads updating the same table state
	unlock := tableMutex.Lock(table)

	// Load state
	value, ok := stateMap.Load(table)

	var state *GameState

	if ok {
		stateCopy := *value.(*GameState)
		state = &stateCopy
		state.setClientPlayerByName(player)
	} else {
		log.Printf("getState Not OK\n")
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
	createTable("Bot Room A - 1 bot", "bot1a", 1, true)
	createTable("Bot Room B - 1 bot", "bot2b", 1, true)
	createTable("Bot Room C - 1 bot", "bot2c", 1, true)

	// For client developers, create hidden tables for each # of bots (for ease of testing with a specific # of players in the game)
	// These will not update the lobby

	//for i := 1; i < 8; i++ {
	//	createTable(fmt.Sprintf("Dev Room - %d bots", i), fmt.Sprintf("dev%d", 2), 2, false)
	//}

}

func createTable(serverName string, table string, botCount int, registerLobby bool) {
	log.Printf("createTable %s  %s  %d\n", serverName, table, botCount)

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
