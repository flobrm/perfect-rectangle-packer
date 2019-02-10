package tiling

import (
	"localhost/flobrm/tilingsolver/core"
	"time"
)

// SolveNaive is a depth first solver without many clever optimizations
func SolveNaive(boardDims core.Coord, tileDims []core.Coord, start []core.TilePlacement,
	stop []core.TilePlacement, endTime time.Time) (map[string]int, string, uint) {
	tiles := make([]Tile, len(tileDims))
	for i := range tileDims {
		tiles[i] = NewTile(tileDims[i].X, tileDims[i].Y)
		tiles[i].Index = i
	}
	board := NewBoard(boardDims, tiles)
	// solutions := make([][]Tile, 0) //random starting value
	solutions := make(map[string]int, 0)

	// Only skip the last 3 start tiles if we have to use a separate tile for each corner
	// aka only if the largest side of the largest tile is smaller than the smallest side of the board.
	doSkipLastStartTiles := boardDims.Y > tileDims[0].X

	placedTileIndex := make([]int, len(tileDims))[:0] //keeps track of which tiles are currently placed in which order
	tilesPlaced := 0
	numTiles := len(tiles)
	startIndex := 0
	startRotation := false
	step := 0
	totalTilesPlaced := uint(0)

	// rotatedSolutions := 0
	// totalSolutions := 0

	//place startTiles
	if start != nil { //TODO test what if nil, what if no fit, what if index out of bounds?
		if doSkipLastStartTiles && start[0].Idx >= len(tiles)-4 { //check if early exit is possible for this job
			return solutions, "solved", totalTilesPlaced
		}
		for _, placement := range start {
			if board.Fits(tiles[placement.Idx], placement.Rot) {
				board.PlaceTile(&tiles[placement.Idx], placement.Rot)
				tilesPlaced++
				placedTileIndex = append(placedTileIndex, placement.Idx)
			} else {
				if !placement.Rot {
					startIndex = placement.Idx
					startRotation = true
				} else {
					startIndex = placement.Idx + 1
					startRotation = false
				}
				break
			}
		}
	}
	//check if stopTiles is legit
	if stop != nil {
		for _, placement := range stop {
			if placement.Idx >= len(tiles) {
				stop = nil
				break
			}
		}
	}

	for {
		// if step >= 0 { //&& step < 8500 {
		// fmt.Println("step: ", step)
		// 	SaveBoardPic(board, fmt.Sprintf("%sdebugPic%010d.png", imgPath, step), 5)
		// }
		// if step >= 1867505 {
		// 	fmt.Println("start debugging here")
		// }
		// if step == 500 {
		// 	return solutions
		// }

		//check for stop conditions
		if stop != nil {
			if len(placedTileIndex) == len(stop) {
				for i, placement := range stop {
					if placedTileIndex[i] < placement.Idx {
						break
					}
					if placedTileIndex[i] > placement.Idx ||
						placedTileIndex[i] == placement.Idx && !placement.Rot && tiles[placement.Idx].Turned {
						// fmt.Println(step)
						// fmt.Println(tiles)
						// fmt.Println(stop)
						// fmt.Println(placedTileIndex)
						//SaveBoardPic(board, fmt.Sprintf("%sdebugPic%010d.png", imgPath, step), 5)
						return solutions, "solved", totalTilesPlaced
					}
				}
			}
		}
		if time.Now().After(endTime) {
			return solutions, "interrupted", totalTilesPlaced
		}

		if tilesPlaced == numTiles {
			// if step == 1867505 {
			// 	fmt.Println("stop to check stuff")
			// }
			// SaveBoardPic(board, fmt.Sprintf("%s%010dFirstSolution.png", imgPath, step), 5)
			newSolution := make([]Tile, numTiles)
			copy(newSolution, tiles)
			board.GetCanonicalSolution(&newSolution)
			preLength := len(solutions)
			solutions[TileSliceToJSON(newSolution)] = 1
			if len(solutions) != preLength {
				// fmt.Println("solution found:")
				// fmt.Println(placedTileIndex)
				// fmt.Println(tiles)
				// SaveBoardPic(board, fmt.Sprintf("%s%010d_Solution.png", imgPath, step), 5)
				// SavePicFromPuzzle(board.Size, newSolution, fmt.Sprintf("%s%010d_RotatedSolution.png", imgPath, step), 5)
			}
			// solutions = append(solutions, newSolution)
			// fmt.Println("solution found")
		}
		step++

		// fmt.Print("starting new round\n\n")
		// fmt.Println(board)

		placedThisRound := false //Is this still necessary? we break after placing a tile
		for i := startIndex; i < len(tiles); i++ {
			if !tiles[i].Placed {
				// fmt.Println("trying to fit tile", tiles[i])
				if startRotation == false && board.Fits(tiles[i], false) { //place normal
					// fmt.Println("fitting tile normal", tiles[i])
					board.PlaceTile(&tiles[i], false)
					// fmt.Println("placed tile normal", board)
					startIndex = 0
					startRotation = false
					placedThisRound = true
					placedTileIndex = append(placedTileIndex, i)
					tilesPlaced++
					totalTilesPlaced++
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
					totalTilesPlaced++
					break
				}
				startRotation = false
			}
		}
		if !placedThisRound {
			if tilesPlaced == 0 { //No tiles on board and impossible to place new tiles, so exit
				// fmt.Println("rotated solutions:", rotatedSolutions)
				// fmt.Println("total solutions:", totalSolutions)
				return solutions, "solved", totalTilesPlaced
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

			//This only works if all tiles are smaller than both board sides
			if tilesPlaced == 0 { //Skip the last 3 startingtiles, solutions with those already exist
				if doSkipLastStartTiles && startIndex == len(tiles)-4 {
					return solutions, "solved", totalTilesPlaced
				}
			}
		}
	}
}
