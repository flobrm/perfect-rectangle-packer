package main

import (
	"fmt"
	"localhost/flobrm/tilingsolver/tiling"
	"time"
)

var imgPath = "C:/Users/Florian/go/src/localhost/flobrm/tilingsolver/img/"

func main() {

	start := time.Now()

	// build asqas 8
	var tiles [8]tiling.Coord
	for i := range tiles {
		tiles[7-i] = tiling.Coord{X: i + 2, Y: i + 1}
	}
	solveNaive(tiling.Coord{X: 15, Y: 16}, tiles[:])

	elapsed := time.Since(start)
	fmt.Println("time: ", elapsed)

	fmt.Println()
}

func solveNaive(boardDims tiling.Coord, tileDims []tiling.Coord) [][]tiling.Tile {
	tiles := make([]tiling.Tile, len(tileDims))
	for i := range tileDims {
		tiles[i] = tiling.NewTile(tileDims[i].X, tileDims[i].Y)
	}
	board := tiling.NewBoard(boardDims, tiles)
	solutions := make([][]tiling.Tile, 0) //random starting value

	placedTileIndex := make([]int, len(tileDims))[:0] //keeps track of which tiles are currently placed in which order
	tilesPlaced := 0
	numTiles := len(tiles)
	startIndex := 0
	startRotation := false
	step := 0

	for {
		// if step > 0 {
		// 	// fmt.Println("step: ", step)
		// 	tiling.SaveBoardPic(board, fmt.Sprintf("%sdebugPic%06d.png", imgPath, step), 5)
		// }
		// if step >= 268 {
		// 	fmt.Println("start debugging here")
		// }

		if tilesPlaced == numTiles {
			//TODO record solution
			//TODO return if only 1 solution requested
			tiling.SaveBoardPic(board, fmt.Sprintf("%sSolution%06d.png", imgPath, step), 5)
			fmt.Println("solution found")
		}
		step++

		// fmt.Print("starting new round\n\n")
		// fmt.Println(board)

		placedThisRound := false
		for i := startIndex; i < len(tiles); i++ {
			if !tiles[i].Placed {
				// fmt.Println("trying to fit tile", tiles[i])
				if board.Fits(tiles[i], false) && startRotation == false { //place normal
					// fmt.Println("fitting tile normal", tiles[i])
					board.PlaceTile(&tiles[i], false)
					// fmt.Println("placed tile normal", board)
					startIndex = 0
					startRotation = false
					placedThisRound = true
					placedTileIndex = append(placedTileIndex, i)
					tilesPlaced++
					break
				}
				// fmt.Println("trying to fit tile turned", tiles[i])
				if board.Fits(tiles[i], true) { // place turned
					// fmt.Println("fitting tile turned", tiles[i])
					board.PlaceTile(&tiles[i], true)
					// fmt.Println("placed tile turned", board)
					startIndex = 0
					startRotation = false
					placedThisRound = true
					placedTileIndex = append(placedTileIndex, i)
					tilesPlaced++
					break
				}
			}
		}
		if !placedThisRound {
			if tilesPlaced == 0 { //No tiles on board and impossible to place new tiles, so exit
				return solutions
			}
			//Remove the last tile and keep track of which tile to try next
			// fmt.Println("REMOVING tile", tiles[placedTileIndex[len(placedTileIndex)-1]])
			tilesPlaced--
			board.RemoveLastTile()
			tiles[placedTileIndex[len(placedTileIndex)-1]].Remove()
			// fmt.Println(board)
			if !tiles[placedTileIndex[len(placedTileIndex)-1]].Turned {
				startIndex = placedTileIndex[len(placedTileIndex)-1]
				startRotation = true
			} else {
				startIndex = placedTileIndex[len(placedTileIndex)-1] + 1
				startRotation = false
			}
			placedTileIndex = placedTileIndex[:len(placedTileIndex)-1]
		}
	}
}
