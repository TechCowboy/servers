# 5 Card Stud Server
This is an incomplete 5 Card Stud server written in GO for the purpose of writing/testing 5 Card Stud clients. This is my first project in GO, so do not expect expert use of the language.  All game logic is written by scratch, with the exception of using github.com/gin-gonic/gin to rank the final 5 card hands.

As this is focused on assisting in designing a client, it currently provides:
* Multiple concurrent games (tables) via the `?table=[Alphanumeric value]` url parameter
* Bots that simulate a game (Not highly intelligent, but they will check, bet, raise, and fold based on combiniation of random/simple decision logic)
* Emulating new players (BOTS) joining a table during an existing game with `?count=[new player count]`
* End of game detection, determining the winner and awarding pot to winning players, and starting a new game
* Giving valid moves to each player, and only allows these moves (client can send different moves with no change to state)
* Opening Anti, posting bring-in, low/high bets in 4th/5th street. Still missing exceptions like allowing high bet with visible pairs
* **Does not** support multiple clients

## How to use

The server is **not** real time. Each time you call ``/state`` it will step forward in time, either playing a BOT's move, or giving end of round details (with activePlayer set to -1 indicating no play is left). Calling ``/state`` will then begin the next round.

The game is over when **round 5** is sent. The next call to ``/state`` will begin a new game.

You can view the state as-is by calling `/view` . This could be useful when comparing what the client is seeing without stepping forward in time.

## Concurrent games support

The server supports unlimited concurrent games (tables) going at once. Simply pass `?table=[AlphaNumeric Name]` to all calls to test in an isolated table, otherwise it will assume a table name of `default`.

## Public endpoint

As an alternative to running locally, use the latest api is running here:

https://mock-server-7udvkexssq-uc.a.run.app/

**TIP:** If using the public endpoint, append each call with your own table name, e.g. `?table=Eric123` 

## Api paths

* GET `/state` - Advance forward (AI/Game Logic) and return updated state as compact json
* GET ``/move/[code]`` - Apply your player's move and return updated state as compact json. e.g. ``/move/CH`` to "Check", ``/move/BL`` to "Bet 5 (low)".
* GET `/view` - View the current state as-is without advancing, as formatted json. Useful for debugging in a browser alongside the client. **NOTE:** If you call this for an uninitated game, a different randomly initiated game will be returned every time.

Both `state` and `move` accept GET or POST.

## Query parameters
* `table=[Alphanumeric]` - Use to play in an isolated game
* `count=[2-8]` - Include on the `/state` call to set the number of players in a game. 
    * If the number is larger than the current player count, new players will join, waiting until the next game.
    * If the number is smaller, a new game will start.

## State structure
This is highly subject to change, but focused on a low nested structure and speed of parsing for 8-bit clients.

A client centric state is returned. This means that your client will only see the values of cards it is meant to see, and the player array will always start with your client's player first.

#### Json Properties

* `lastResult` - Will be filled with text when round=5 to signal the current game is over. e.g. "So and so won with 2 pairs"
* `round` - The current round (1-5). Round 5 means the game has ended and pot awarded to winning player(s).
* `pot` - The current value of the pot for the current game
* `activePlayer` - The currently active player. Your client is always player 0. This will be `-1` at the end of a round (or end of game) to allow the client to show the last move before starting the next round.
* `validMoves` - An array of legal moves
    * `move` - The code to send to `/move`
    * `name` - The text to show onscreen in the client
* `player` - An array of player objects
    * `name` - The name of the player, or `You` for the client
    * `status` - The player's current in-game status
        * 0 - Just joined, waiting to play the next game
        * 1 - In Game, playing
        * 2 - In Game, Folded
    * `bet` - The total of the player's bet for the current round
    * `move` - Friendly text of the player's most recent move this round
    * `purse` - The player's remaining amount available to bet
    * `hand` - A string of multiple 2 character representation of cards in the player's hand:
        * First char - Value : 2 to 9, T=10, J=Jack, Q=Queen, K=King, A=Ace
        * Second char - Suit : C,S,D,H stand for Clubs, Spades, Diamonds, and Hearts
        * `??` - A hidden card. Also represents a folded hand when `hand` is just `??` and followed by no other cards
    
    

#### Example state

```json
{
    "lastResult": "Thom won with Full House, Eights full of Sixes",
    "round": 1,
    "pot": 0,
    "activePlayer": 0,
    "validMoves": [
        {
            "move": "FO",
            "name": "Fold"
        },
        {
            "move": "CA",
            "name": "Call"
        },
        {
            "move": "RL",
            "name": "Raise 5"
        }
    ],
    "players": [
        {
            "name": "You",
            "status": 1,
            "bet": 0,
            "move": "",
            "purse": 199,
            "hand": "KSKH"
        },
        {
            "name": "Thom",
            "status": 1,
            "bet": 5,
            "move": "BET",
            "purse": 194,
            "hand": "??6H"
        },
        {
            "name": "Mozzwald",
            "status": 0,
            "bet": 0,
            "move": "",
            "purse": 200,
            "hand": ""
        },
    ]
}
```