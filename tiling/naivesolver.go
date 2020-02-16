package tiling

import (
	"localhost/flobrm/tilingsolver/core"
	"time"
)

//These constants show which optimizations are available for the current solver
const (
	FullSSNCheck        = iota
	OneLevelSSN         = iota
	DoGapdetection      = iota
	OneGapDetection     = iota
	AllDownGapDetection = iota
	LeftGapDetection    = iota
	TotalGapAreaCheck   = iota
	ForceFrameUpright   = iota
)

//Debug locations
var imgPath = "C:/Users/Florian/go/src/localhost/flobrm/tilingsolver/img/"

// var imgPath = "/home/florian/golang/src/localhost/flobrm/tilingsolver/img/"

// SolveNaive is a depth first solver without many clever optimizations
// returns a map with solutions, the reason for stopping, the number of steps taken,
// and the tiles as placed on the board at the last step
func SolveNaive(boardDims core.Coord, tileDims []core.Coord, start []core.TilePlacement,
	stop []core.TilePlacement, endTime time.Time, stopOnSolution bool, optimizations map[int]bool, placementOrder int) (
	map[string]int, string, uint, []core.TilePlacement) {

	checkGaps := optimizations[DoGapdetection]
	checkFullSSN := optimizations[FullSSNCheck]
	checkLeftSideGaps := optimizations[LeftGapDetection]
	checkOnlyNextCandidate := !optimizations[AllDownGapDetection]
	checkTotalGapArea := optimizations[TotalGapAreaCheck]
	setUprightBoard := optimizations[ForceFrameUpright]
	boardFlipped := false

	tiles := make([]Tile, len(tileDims))
	if setUprightBoard && boardDims.X > boardDims.Y {
		boardFlipped = true
		tempX := boardDims.X
		boardDims.X = boardDims.Y
		boardDims.Y = tempX
		for i := range start {
			start[i].Rot = !start[i].Rot
		}
		for i := range stop {
			stop[i].Rot = !stop[i].Rot
		}
	}
	for i := range tileDims {
		tiles[i] = NewTile(tileDims[i].X, tileDims[i].Y)
		tiles[i].Index = i
	}
	board := NewBoard(boardDims, tiles, placementOrder)
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
		if doSkipLastStartTiles && start[0].Idx > len(tiles)-4 { //check if early exit is possible for this job
			return solutions, "solved", totalTilesPlaced, nil
		}
		for _, placement := range start {
			if board.Place(&tiles[placement.Idx], placement.Rot, checkFullSSN) {
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
		// if step >= 0 { //&& step%1000 == 0 { //&& step < 8500 {
		// 	// fmt.Println("step: ", step)
		// 	SaveBoardPic(board, fmt.Sprintf("%sdebugPic%010d.png", imgPath, step), 5)
		// }
		// if step >= 232 {
		// 	fmt.Println("start debugging here")
		// }
		// if step == 6000 {
		// 	return solutions, "interrupted", totalTilesPlaced, getCurrentPlacements(placedTileIndex, tiles, boardFlipped)
		// }
		//check for stop conditions
		if stop != nil {
			if len(placedTileIndex) == len(stop) {
				for i, placement := range stop {
					if placedTileIndex[i] < placement.Idx || placedTileIndex[i] == placement.Idx && placement.Rot && !tiles[placedTileIndex[i]].Turned {
						break
					}
					if placedTileIndex[i] > placement.Idx ||
						placedTileIndex[i] == placement.Idx && !placement.Rot && tiles[placement.Idx].Turned ||
						i == len(stop)-1 && placedTileIndex[i] == placement.Idx && placement.Rot == tiles[placement.Idx].Turned {
						// fmt.Println(step)
						// fmt.Println(tiles)
						// fmt.Println(stop)
						// fmt.Println(placedTileIndex)
						// SaveBoardPic(board, fmt.Sprintf("%sdebugPic%010d.png", imgPath, step), 5)
						// fmt.Println("past stopper")
						return solutions, "solved", totalTilesPlaced, getCurrentPlacements(placedTileIndex, tiles, boardFlipped)
					}
				}
			}
		}
		if time.Now().After(endTime) {
			return solutions, "interrupted", totalTilesPlaced, getCurrentPlacements(placedTileIndex, tiles, boardFlipped)
		}

		if tilesPlaced == numTiles {
			// if step == 1867505 {
			// 	fmt.Println("stop to check stuff")
			// }
			// SaveBoardPic(board, fmt.Sprintf("%s%010dFirstSolution.png", imgPath, step), 5)
			newSolution := make([]Tile, numTiles)
			copy(newSolution, tiles)
			board.GetCanonicalSolution(&newSolution)
			if boardFlipped {
				rotateTiles(&newSolution)
			}
			preLength := len(solutions)
			solutions[TileSliceToJSON(newSolution)] = 1
			if stopOnSolution {
				return solutions, "solved1", totalTilesPlaced, getCurrentPlacements(placedTileIndex, tiles, boardFlipped)
			}

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
				// handle double tiles
				if i > 0 && !tiles[i-1].Placed && tiles[i-1].X == tiles[i].X && tiles[i-1].Y == tiles[i].Y {
					startRotation = false
					continue
				}
				// fmt.Println("trying to fit tile", tiles[i])
				if startRotation == false && board.Place(&tiles[i], false, checkFullSSN) { //place normal
					// fmt.Println("fitting tile normal", tiles[i])
					// fmt.Println("placed tile normal", board)
					if checkGaps && board.HasUnfillableGaps(checkOnlyNextCandidate, checkLeftSideGaps, checkTotalGapArea) {
						// SaveBoardPic(board, fmt.Sprintf("%sdebugPic%010d_s%2d.png", imgPath, step, i), 5)
						board.RemoveLastTile()
						tiles[i].Remove()
						totalTilesPlaced++
					} else {
						startIndex = 0
						startRotation = false
						placedThisRound = true
						placedTileIndex = append(placedTileIndex, i)
						tilesPlaced++
						totalTilesPlaced++
						break
					}
				}
				// fmt.Println("trying to fit tile turned", tiles[i])
				if tiles[i].X != tiles[i].Y && board.Place(&tiles[i], true, checkFullSSN) { // place turned, if tile is not square
					// fmt.Println("fitting tile turned", tiles[i])
					// fmt.Println("placed tile turned", board)
					if checkGaps && board.HasUnfillableGaps(checkOnlyNextCandidate, checkLeftSideGaps, checkTotalGapArea) {
						// SaveBoardPic(board, fmt.Sprintf("%sdebugPic%010d_t%2d.png", imgPath, step, i), 5)
						board.RemoveLastTile()
						tiles[i].Remove()
						totalTilesPlaced++
					} else {
						startIndex = 0
						startRotation = false
						placedThisRound = true
						placedTileIndex = append(placedTileIndex, i)
						tilesPlaced++
						totalTilesPlaced++
						break
					}
				}
				startRotation = false
			}
		}
		if !placedThisRound {
			if tilesPlaced == 0 { //No tiles on board and impossible to place new tiles, so exit
				// fmt.Println("rotated solutions:", rotatedSolutions)
				// fmt.Println("total solutions:", totalSolutions)
				return solutions, "solved", totalTilesPlaced, nil
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
				if doSkipLastStartTiles && startIndex > len(tiles)-4 {
					return solutions, "solved", totalTilesPlaced, nil
				}
			}
		}
	}
}

func getCurrentPlacements(tileIndexes []int, tiles []Tile, boardFlipped bool) []core.TilePlacement {
	placements := make([]core.TilePlacement, len(tileIndexes))[:0]
	for _, idx := range tileIndexes {
		if boardFlipped {
			placements = append(placements, core.TilePlacement{Idx: idx, Rot: !tiles[idx].Turned})
		} else {
			placements = append(placements, core.TilePlacement{Idx: idx, Rot: tiles[idx].Turned})
		}
	}
	return placements
}

func rotateTiles(tiles *[]Tile) {
	for i := range *tiles {
		tempX := (*tiles)[i].X
		(*tiles)[i].X = (*tiles)[i].Y
		(*tiles)[i].Y = tempX
		(*tiles)[i].Turned = !(*tiles)[i].Turned

		tempCurW := (*tiles)[i].X
		(*tiles)[i].CurW = (*tiles)[i].CurH
		(*tiles)[i].CurH = tempCurW
	}
}
