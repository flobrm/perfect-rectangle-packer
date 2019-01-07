package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"localhost/flobrm/tilingsolver/tileio"
	"localhost/flobrm/tilingsolver/tiling"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"
)

var imgPath = "C:/Users/Florian/go/src/localhost/flobrm/tilingsolver/img/"

// var imgPath = "/home/florian/golang/src/localhost/flobrm/tilingsolver/img/"

// var inputFile = "/home/florian/golang/src/localhost/flobrm/tilingsolver/input.csv"
var inputFile = "C:/Users/Florian/go/src/localhost/flobrm/tilingsolver/input.csv"

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var inputPath = flag.String("inputpath", "", "input file with puzzles")

var numTiles = flag.Int("num_tiles", 0, "Solve only puzzles with this many tiles.")
var puzzleLimit = flag.Int("puzzle_limit", 0, "Solve at most N puzzles")
var batchSize = flag.Int("batch_size", 1, "How many puzzles should the program reserve at once")
var solverID = flag.Int("solver_id", 0, "Used to differentiate between different solvers and hardware")

func main() {
	flag.Parse()
	//profiling cpu if cpuprofile is specified
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if *solverID <= 0 {
		fmt.Println("No, or illegal, solver_id specified")
		// return //TODO reenable after debugging
	}

	//TODO remove after debugging
	// inputPath = &inputFile
	// if *inputPath == "" {
	// 	log.Fatal("No inputpath specified")
	// }

	start := time.Now()

	//solveFromDatabase(*numTiles, *puzzleLimit, *batchSize, *solverID)
	// fmt.Print(len(solveAsQas8()))
	fmt.Print(len(solveTestCase()))

	elapsed := time.Since(start)
	log.Println("time: ", elapsed)

	//Profiling memory
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
}

func solveTestCase() map[string][]tiling.Tile {
	board := tiling.Coord{X: 41, Y: 25}
	tiles := make([]tiling.Coord, 11)
	tileBytes := []byte("[{\"X\":22,\"Y\":14},{\"X\":20,\"Y\":6},{\"X\":20,\"Y\":3},{\"X\":20,\"Y\":2},{\"X\":17,\"Y\":1},{\"X\":15,\"Y\":11},{\"X\":14,\"Y\":13},{\"X\":10,\"Y\":5},{\"X\":7,\"Y\":6},{\"X\":7,\"Y\":5},{\"X\":6,\"Y\":1}]")
	json.Unmarshal(tileBytes, &tiles)
	return solveNaive(board, tiles)
}

func solveAsQas3() map[string][]tiling.Tile {
	// build asqas 3
	var tiles [3]tiling.Coord
	for i := range tiles {
		tiles[2-i] = tiling.Coord{X: i + 2, Y: i + 1}
	}
	return solveNaive(tiling.Coord{X: 5, Y: 4}, tiles[:])
	// for i := range solutions {
	// 	board := tiling.NewBoard(tiling.Coord{X: 15, Y: 16}, solutions[i])
	// 	boardTiles := make([]*tiling.Tile, len(solutions[i]))
	// 	for j := range solutions[i] {
	// 		boardTiles[j] = &solutions[i][j]
	// 	}
	// 	board.Tiles = boardTiles
	// 	tiling.SaveBoardPic(board, fmt.Sprint("img/", i, ".png"), 5)
	// 	fmt.Println(i, solutions[i])
	// }
}

func solveAsQas8() map[string][]tiling.Tile {
	// build asqas 8
	var tiles [8]tiling.Coord
	for i := range tiles {
		tiles[7-i] = tiling.Coord{X: i + 2, Y: i + 1}
	}
	return solveNaive(tiling.Coord{X: 15, Y: 16}, tiles[:])
	// for i := range solutions {
	// 	board := tiling.NewBoard(tiling.Coord{X: 15, Y: 16}, solutions[i])
	// 	boardTiles := make([]*tiling.Tile, len(solutions[i]))
	// 	for j := range solutions[i] {
	// 		boardTiles[j] = &solutions[i][j]
	// 	}
	// 	board.Tiles = boardTiles
	// 	tiling.SaveBoardPic(board, fmt.Sprint("img/", i, ".png"), 5)
	// 	fmt.Println(i, solutions[i])
	// }
}

func solveAsQas20() {
	// build asqas 20
	var tiles [20]tiling.Coord
	for i := range tiles {
		tiles[19-i] = tiling.Coord{X: i + 2, Y: i + 1}
	}
	solveNaive(tiling.Coord{X: 55, Y: 56}, tiles[:])
}

func solveFromFile(filePath *string) {
	pathParts := strings.Split(*filePath, ".")
	extension := pathParts[len(pathParts)-1]
	var reader tileio.PuzzleReader
	if extension == "json" {
		reader = tileio.NewPuzzleJSONReader(*filePath)
	} else if extension == "csv" {
		reader = tileio.NewPuzzleCSVReader(*filePath)
	}

	for puzzle, err := reader.NextPuzzle(); err == nil; puzzle, err = reader.NextPuzzle() {
		solutions := solveNaive(puzzle.Board, *puzzle.Tiles)
		log.Println("solved a puzzle")
		for _, solution := range solutions {
			//TODO write solutions
			fmt.Println(solution)
		}
	}
	log.Println("finished")
}

func solveFromDatabase(numTiles int, puzzleLimit int, batchSize int, solverID int) {
	db := tileio.Open()
	defer tileio.Close(db)

	puzzlesSolved := 0

	for puzzleLimit == 0 || puzzlesSolved < puzzleLimit {
		puzzles, err := tileio.GetNewPuzzles(db, batchSize, numTiles)
		if err != nil {
			log.Println(err)
			log.Fatal("whatever")
		}

		for _, puzzle := range puzzles {
			log.Println("start solving puzzle", puzzle.ID)
			solveStart := time.Now()
			solutions := solveNaive(puzzle.BoardDims, *puzzle.Tiles)
			solveTime := time.Since(solveStart)

			log.Println("finished solving puzzle ", puzzle.ID, " in ", solveTime)
			log.Println(len(solutions), "solutions found for puzzle ", puzzle.ID)

			//insert solutions into db
			err = tileio.InsertSolutions(db, puzzle.ID, solverID, solveTime, &solutions)
			if err != nil {
				log.Fatal(err)
			}
			puzzlesSolved++
		}
		if len(puzzles) < batchSize {
			break
		}
	}
	log.Println("finished, solved ", puzzlesSolved, " puzzles")
}

func solveNaive(boardDims tiling.Coord, tileDims []tiling.Coord) map[string][]tiling.Tile {
	tiles := make([]tiling.Tile, len(tileDims))
	for i := range tileDims {
		tiles[i] = tiling.NewTile(tileDims[i].X, tileDims[i].Y)
	}
	board := tiling.NewBoard(boardDims, tiles)
	// solutions := make([][]tiling.Tile, 0) //random starting value
	solutions := make(map[string][]tiling.Tile, 0)

	placedTileIndex := make([]int, len(tileDims))[:0] //keeps track of which tiles are currently placed in which order
	tilesPlaced := 0
	numTiles := len(tiles)
	startIndex := 0
	startRotation := false
	step := 0

	rotatedSolutions := 0
	totalSolutions := 0

	for {
		// if step > 0 {
		// 	fmt.Println("step: ", step)
		// 	tiling.SaveBoardPic(board, fmt.Sprintf("%sdebugPic%06d.png", imgPath, step), 5)
		// }
		// if step >= 664 {
		// 	fmt.Println("start debugging here")
		// }
		// if step == 500 {
		// 	return solutions
		// }

		if tilesPlaced == numTiles {
			// if step == 45432021 {
			// 	fmt.Println("stop to check stuff")
			// }
			//TODO return if only 1 solution requested
			// tiling.SaveBoardPic(board, fmt.Sprintf("%sSolution%06d.png", imgPath, step), 5)
			totalSolutions++
			newSolution := make([]tiling.Tile, numTiles)
			copy(newSolution, tiles)
			rotatedSolutions += board.GetCanonicalSolution(&newSolution)
			preLength := len(solutions)
			solutions[tiling.TileSliceToJSON(newSolution)] = newSolution
			if len(solutions) != preLength {
				tiling.SaveBoardPic(board, fmt.Sprintf("%sSolution%010d.png", imgPath, step), 5)
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
				fmt.Println("rotated solutions:", rotatedSolutions)
				fmt.Println("total solutions:", totalSolutions)
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

			//This only works if all tiles are smaller than both board sides
			// if tilesPlaced == 0 { //Skip the last 3 startingtiles, solutions with those already exist
			// 	if startIndex == len(tiles)-4 {
			// 		return solutions
			// 	}
			// }
		}
	}
}
