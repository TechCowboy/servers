# Reversi/Othello Server
This is an Othello server written in GO. This was adapted from Eric Carr's 5 card stud server.  

It currently provides:
* Multiple concurrent games (tables) via the `?table=[Alphanumeric value]` url parameter
* Bots that simulate players 
* Auto moves for players that do not move in time 
* Auto drops players that have not interacted with the server after some time (timed out)

## Accessing the Game Server API

Clone and run the server locally:
    ```
    go run .
    ```


## Basic Flow

Game Server:
Tells Lobby Server what game it is, it's app_id, and the tables available. 
The Lobby Client grabs that info from the Lobby Server and the platforms it supports.
When the user selects a game and table, the Lobby Client writes to the app key (in a predefined structure) and launches the game client
The game client grabs it's app key and uses that server/table and begins "gaming"

You can read http://lobby.fujinet.online/docs - scroll down to "So you're a programmer and want to use Lobby Server?" - the first bit is hosting the lobby server, which IMO should be at the bottom, as most people want to know how to get their server/game in the lobby.

So what happens when tell CONFIG to go to the lobby?
It puts the lobby client found on fujinet.online in slot one and boots 

A game client is expected to:

1. Call `/tables` to present a list of tables to join.
2. There is no specific call to join a table. Simply retrieving the state will cause the player to join that table.
2. In a loop:
    A. Call `/state?player=X&table=Y` to retrieve that latest state
    B. Call `/move/"POSITION"?player=X&table=Y` to place a move if it is the current player's turn.  Position is calculated by row*8 + col
3. If the player wishes to exit the game, the client should call `/leave?player=X&table=Y`


## Retrieving the Table List

Retrieve the list of tables by calling `/tables`. 
___
**DEVELOPER TIP** - Call `/tables?dev=1` to retrieve a list of hidden tables for developer usage. You can test your client using these "dev*" tables without impacting live player facing games on the public server.
___

A list of objects with the following properties will be returned:

* `t` - Table id. Pass this as the `table` url parameter to other calls.
* `n` - Friendly name of table to show in a list for the player to choose
* `p` - Number of players currently connected. 0 if none.
* `m` - Number of max available player slots available.

Example response of `/tables` call
```
json
[{
    "t":"basement",
    "n":"The Basement",
    "p":3,
    "m":8
},{
    "t":"ai2",
    "n":"AI Room - 2 bots",
    "p":0,
    "m":6
}, ...]
```

These tables are psuedo real time. Call `/state` will run any housekeeping tasks (bot or player auto-move, deal card, proceed with dealing). Since a call to `/state` is required to advance the game, a table with bots in it will not actually play until one or more clients are connected and calling `/state`. Each player has a limited amount of time to make a move before the server makes a move on their behalf. BOTs take a second to move.

* The game is over when the active player is -1 is sent. The next game will begin automatically after a few seconds.
* The game is waiting on more players when **round 0** is sent.
* Clients should call `/leave` when a player exits the game or table, rather than rely on the server to eventually drop the player due to inactivity.

You can view the state as-is by calling `/view`.

## Api paths

* `/state` - Advance forward (AI/Game Logic) and return updated state as compact json
* `/move/[position]` - Apply your player's move and return updated state as compact json. 
* `/leave` - Leave the table. Each client should call this when a player exits the game
* `/view?table=N` - View the current state as-is without advancing, as formatted json. Useful for debugging in a browser alongside the client. **NOTE:** If you call this for an uninitated game, a different randomly initiated game will be returned every time. Only `table` query parameter is required.
* `/tables` - Returns a list of available REAL tables along with player information. No query parameters are required
* `/updateLobby` - Use to manually force a refresh of state to the Lobby. No query parameters are required.

All paths accept GET or POST for ease of use.

## Query parameters

### Required
All paths require the query parameters below, unless otherwise specified.
* `TABLE=[Alphanumeric]` - **Required** - Use to play in an isolated game. Case insensitive.
* `PLAYER=[Alphanumeric]` - **Required for Real** - Player's name. Treated as case insensitive unique ID.

### Optional
* `RAW=1` - **Optional** - Use to return key[byte 0]value[byte 0] pairs instead of json output - similar to FujiNet json parsing, with 0x00 used as delimiter instead of line end
* `UC=1` - **Optional** - Use with raw, to make the result data upper case
* `LC=1` - **Optional** - Use with raw, to make the result data lower case

## State structure
This is focused on a low nested structure and speed of parsing for 8-bit clients.

A client centric state is returned. This means that your client will only see the values of cards it is meant to see, and the player array will always start with your client's player first, though all clients will see all players in the same order.

#### Json Properties
Keys are single character, lower case, to make parsing easier on 8-bit clients. Array keys are 2 character.

* `l` - Will be filled with text when round=`64` to signal the current game is over. e.g. "So and so won", or when round=`0` to indicate waiting for more players to join.
* `r` - The current round (1-5). Round 5 means the game has ended and pot awarded to winning player(s).
* `p` - The current value of the pot for the current game
* `a` - The currently active player. Your client is always player 0. This will be `-1` at the end of a round (or end of game) to allow the client to show the last move before starting the next round.
* `m` - Move time - Number of seconds remaining for current player to make their move, or until the next game will start. If a player does not send a move within this time, the server will auto-move for them (post/check if possible, otherwise a fold)
* `v` - Viewing - If all player spots are full, your client's player will not join the game, but instead view the game as a spectator.  In this case, this will be `1` to indicate that you are only viewing. Otherwise, this will be `0` during normal play. 
* `vm` - An array of Valid Moves
    * `m` - The move code to send to `/move`
    * `n` - The friendly name of the move to show onscreen in the client
* `pl` - An array of player objects
    * `n` - Name - The name of the player, or `You` for the client
    * `s` - Status - The player's current in-game status
        * 0 - Just joined, waiting to play the next game
        * 1 - In Game, playing
        * 2 - In Game, Folded
        * 3 - Left the table (will be gone next game)
    * `m` - Move - Friendly text of the player's most recent move this round
        

#### Example state

ignore comments with #
```json
{
  "l": "", # last response
  "t": 1,  # turn number
  "a": 0,  # active player  -1 = game over
  "m": 5,  # time left to move
  "v": 0,  # =1 viewing only
  "vm": [  # valid moves
    {
      "m": "20", # move
      "n": "3,5" # name
    },
    {
      "m": "29",
      "n": "4,6"
    },
    {
      "m": "34",
      "n": "5,3"
    },
    {
      "m": "43",
      "n": "6,4"
    }
  ],
  "pl": [ # players
    {
      "n": "Hal BOT", # name
      "s": 1,         # status
      "m": "",        # Move
      "c": "B",       # Color
      "sc": 2         # Score
    },
    {
      "n": "TechCowboy",
      "s": 1,
      "m": "",
      "c": "W",
      "sc": 2
    }
  ],
  "bd": "...........................BW......WB..........................." # board 64 squares
}
```
