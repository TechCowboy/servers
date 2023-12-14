package main

import (
	"fmt"
	"sync"
)

// Concurrent slice implementation

type ConcurrentGameSlice struct {
	sync.RWMutex
	items []*Game
}

func NewConcurrentGameSlice() ConcurrentGameSlice {
	return ConcurrentGameSlice{}
}

func (cs *ConcurrentGameSlice) Len() int {
	cs.RLock()
	defer cs.RUnlock()

	return len(cs.items)
}

// append and return a pointer to the game because the Append updates ServerUrl
func (cs *ConcurrentGameSlice) Append(game *Game) (updatedGame *Game) {
	cs.Lock()
	defer cs.Unlock()

	pos := len(cs.items)

	game.ServerUrl += fmt.Sprintf("/games/%d/", pos)

	cs.items = append(cs.items, game)

	return game
}

func (cs *ConcurrentGameSlice) GetAtPos(index int) (game *Game, exists bool) {
	cs.RLock()
	defer cs.RUnlock()

	if index < 0 {
		return game, false
	}

	if index < len(cs.items) {
		return cs.items[index], true
	}

	return game, false
}

func (cs *ConcurrentGameSlice) AllAsMap() (output MapSlice) {
	cs.RLock()
	defer cs.RUnlock()

	for _, game := range cs.items {
		output = append(output, game.M())
	}

	return output
}
