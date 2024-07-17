package main

import (
	"log"
	"strconv"
	//"go/constant"
)

const SIZE = 8

/***********************************************
 * Calculates which squares are valid moves    *
 * for player. Valid moves are recorded in the *
 * moves array - true indicates a valid move,  *
 * false indicates an invalid move.            *
 * First parameter is the board array          *
 * Second parameter is the moves array         *
 * Third parameter identifies the player       *
 * to make the move.                           *
 * Returns valid move count.                   *
 **********************************************/

func valid_moves(state *GameState) []validMove {
	var rowdelta int = 0 /* Row increment around a square    */
	var coldelta int = 0 /* Column increment around a square */
	var x int = 0        /* Row index when searching         */
	var y int = 0        /* Column index when searching      */
	var player_color Cell = CELL_BLACK
	var opponent_color Cell = CELL_WHITE
	var moves []validMove
	var move validMove

	if state.ActivePlayer == int(STATUS_INACTIVE) {
		return nil
	}

	if state.ActivePlayer == int(CELL_BLACK) {
		log.Printf("Active Player is Black")
		player_color = CELL_BLACK
		opponent_color = CELL_WHITE
	} else {
		log.Printf("Active Player is White")
		player_color = CELL_WHITE
		opponent_color = CELL_BLACK
	}

	/* Find squares for valid moves.                           */
	/* A valid move must be on a blank square and must enclose */
	/* at least one opponent square between two player squares */
	for row := 0; row < SIZE; row++ {
		for col := 0; col < SIZE; col++ {

			if state.board[row*SIZE+col].cell != CELL_EMPTY { /* Is it a blank square?  */
				continue /* No - so on to the next */
			}

			/* Check all the squares around the blank square  */
			/* for the opponents counter                      */
			for rowdelta = -1; rowdelta <= 1; rowdelta++ {
				for coldelta = -1; coldelta <= 1; coldelta++ {
					/* Don't check outside the array, or the current square */
					if row+rowdelta < 0 || row+rowdelta >= SIZE ||
						col+coldelta < 0 || col+coldelta >= SIZE ||
						(rowdelta == 0 && coldelta == 0) {
						continue
					}

					/* Now check the square */
					if state.board[(row+rowdelta)*SIZE+col+coldelta].cell == opponent_color {
						/* If we find the opponent, move in the delta direction  */
						/* over opponent counters searching for a player counter */
						x = row + rowdelta /* Move to          */
						y = col + coldelta /* opponent square  */

						/* Look for a player square in the delta direction */
						for {
							x += rowdelta /* Go to next square */
							y += coldelta /* in delta direction*/

							/* If we move outside the array, give up */
							if x < 0 || x >= SIZE || y < 0 || y >= SIZE {
								break
							}
							/* If we find a blank square, give up */
							if state.board[x+y*SIZE].cell == CELL_EMPTY {
								break
							}

							/*  If the square has a player counter */
							/*  then we have a valid move          */
							if state.board[x+y*SIZE].cell == player_color {
								log.Printf("valid move: [%d,%d]", row+1, col+1)

								move.Move = strconv.Itoa(row*8 + col)
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

	return moves

}
