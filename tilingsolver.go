package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"localhost/flobrm/tilingsolver/core"
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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var inputPath = flag.String("inputpath", "", "input file with puzzles")

var numTiles = flag.Int("num_tiles", 0, "Solve only puzzles with this many tiles.")
var puzzleLimit = flag.Int("puzzle_limit", 0, "Solve at most N puzzles")
var batchSize = flag.Int("batch_size", 1, "How many puzzles should the program reserve at once")
var solverID = flag.Int("solver_id", 0, "Used to differentiate between different solvers and hardware")
var useJobs = flag.Bool("jobs", false, "If set solve jobs instead of full puzzles")
var dbstring = flag.String("dbstring", "tiler:tiler@(localhost:3306)/tiling", "Database connection string")
var processTimeout = flag.Int("process_timeout", 0, "Max time in seconds that the solver is allowed")
var puzzleTimeout = flag.Int("puzzle_timeout", 0, "Max time before a puzzle/job is interrupted")
var stopOnSolution = flag.Bool("stop_on_solution", false, "Stop the solver after finding the first solution")

var numSolvers = flag.Int("workers", 1, "number of worker threads")
var processID = flag.String("processID", "1", "An identifier to be able to recognize output from multiple processes")
var jobsFile = flag.String("inputFile", "", "File with puzzles/jobs")

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
		return
	} //TODO check more input
	if *processTimeout == 0 {
		*processTimeout = 3600 * 24 * 365 // a year in seconds, could be any big number
	}
	if *puzzleTimeout == 0 {
		*puzzleTimeout = 3600 * 24 * 365 // a year in seconds, could be any big number
	}

	start := time.Now()

	if *jobsFile != "" {
		taskReader := tileio.NewPuzzleCSVReader(*jobsFile)
		//TODO setup output stuff, for now print to output
		//outputer
		solveTasks(taskReader, *solverID, *processTimeout, *puzzleTimeout, *stopOnSolution)
	} else if *useJobs {
		solveJobsFromDatabase(*dbstring, *numTiles, *puzzleLimit, *batchSize, *solverID, *processTimeout, *puzzleTimeout, *stopOnSolution)
	} else {
		solveFromDatabase(*dbstring, *numTiles, *puzzleLimit, *batchSize, *solverID, *processTimeout, *puzzleTimeout, *stopOnSolution)
	}
	// fmt.Print(len(solveAsQas8()))
	// fmt.Print(len(solveTestCase()))

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

func solveTasks(tasks tileio.PuzzleReader, solverID int, processTimeout int, puzzleTimeout int, stopOnSolution bool) {
	//TODO timekeeping
	puzzlesSolved := 0
	processEndTime := time.Now().Add(time.Duration(1000000000 * int64(processTimeout)))

	var resolutionWriter tileio.PuzzleResolutionWriter
	var err error
	resolutionWriter, err = tileio.NewPuzzleCSVWriter("status.csv", "puzzles.csv") //TODO check for error, close at the end
	if err != nil {
		log.Println("Could not open logging files: ", err)
	}
	defer resolutionWriter.Close()

	for puzzle, err := tasks.NextPuzzle(); err != io.EOF; puzzle, err = tasks.NextPuzzle() {
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("start solving job", puzzle.JobID)
		solveStart := time.Now()
		solveEnd := solveStart.Add(time.Duration(1000000000 * int64(puzzleTimeout)))
		if processEndTime.Before(solveEnd) {
			solveEnd = processEndTime
		}
		solutions, status, tilesPlaced := tiling.SolveNaive(puzzle.Board, *puzzle.Tiles, *puzzle.Start, *puzzle.End, solveEnd, stopOnSolution)
		solveTime := time.Since(solveStart)

		//TODO write everything to file
		resolutionWriter.SaveSolutions(puzzle.PuzzleID, puzzle.JobID, &solutions)
		resolutionWriter.SaveStatus(&puzzle, status, tilesPlaced, solveTime, nil)

		log.Println("finished solving job ", puzzle.JobID, " in ", solveTime)
		log.Println(len(solutions), "solutions found for puzzle ", puzzle.PuzzleID)
		puzzlesSolved++

	}
	log.Println("finished all puzzles")
	log.Println("finished, solved ", puzzlesSolved, " puzzles")
}

func solveTestCase() map[string]int {
	board := core.Coord{X: 41, Y: 25}
	tiles := make([]core.Coord, 11)
	tileBytes := []byte("[{\"X\":22,\"Y\":14},{\"X\":20,\"Y\":6},{\"X\":20,\"Y\":3},{\"X\":20,\"Y\":2},{\"X\":17,\"Y\":1},{\"X\":15,\"Y\":11},{\"X\":14,\"Y\":13},{\"X\":10,\"Y\":5},{\"X\":7,\"Y\":6},{\"X\":7,\"Y\":5},{\"X\":6,\"Y\":1}]")
	json.Unmarshal(tileBytes, &tiles)
	result, _, _ := tiling.SolveNaive(board, tiles, nil, nil, time.Now().Add(time.Duration(1000000000*3600)), false)
	return result
}

func solveAsQas3() map[string]int {
	// build asqas 3
	var tiles [3]core.Coord
	for i := range tiles {
		tiles[2-i] = core.Coord{X: i + 2, Y: i + 1}
	}
	result, _, _ := tiling.SolveNaive(core.Coord{X: 5, Y: 4}, tiles[:], nil, nil, time.Now().Add(time.Duration(1000000000*3600)), false)
	return result
}

func solveAsQas8() map[string]int {
	// build asqas 8
	var tiles [8]core.Coord
	for i := range tiles {
		tiles[7-i] = core.Coord{X: i + 2, Y: i + 1}
	}
	result, _, _ := tiling.SolveNaive(core.Coord{X: 15, Y: 16}, tiles[:], nil, nil, time.Now().Add(time.Duration(1000000000*3600)), false)
	return result
}

func solveAsQas20() {
	// build asqas 20
	var tiles [20]core.Coord
	for i := range tiles {
		tiles[19-i] = core.Coord{X: i + 2, Y: i + 1}
	}
	tiling.SolveNaive(core.Coord{X: 55, Y: 56}, tiles[:], nil, nil, time.Now().Add(time.Duration(1000000000*3600)), false)
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
		solutions, _, _ := tiling.SolveNaive(puzzle.Board, *puzzle.Tiles, nil, nil, time.Now().Add(time.Duration(1000000000*3600)), false)
		log.Println("solved a puzzle")
		for _, solution := range solutions {
			//TODO write solutions
			fmt.Println(solution)
		}
	}
	log.Println("finished")
}

//printMemUsage prints the memory usage, is used as a debug function
func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func solveFromDatabase(dbstring string, numTiles int, puzzleLimit int, batchSize int, solverID int, processTimeout int, puzzleTimeout int, stopOnSolution bool) {
	db := tileio.Open(dbstring)
	defer tileio.Close(db)

	puzzlesSolved := 0
	processEndTime := time.Now().Add(time.Duration(1000000000 * int64(processTimeout)))
	fmt.Println(time.Now(), processEndTime)

	for puzzleLimit == 0 || puzzlesSolved < puzzleLimit {
		puzzles, err := tileio.GetNewPuzzles(db, batchSize, numTiles)
		if err != nil {
			log.Println(err)
			log.Fatal("whatever")
		}

		for _, puzzle := range puzzles {
			log.Println("start solving puzzle", puzzle.ID)
			solveStart := time.Now()
			if solveStart.After(processEndTime) {
				return
			}
			solveEnd := solveStart.Add(time.Duration(1000000000 * int64(puzzleTimeout)))
			if processEndTime.Before(solveEnd) {
				solveEnd = processEndTime
			}
			solutions, status, tilesPlaced := tiling.SolveNaive(puzzle.BoardDims, *puzzle.Tiles, nil, nil, solveEnd, stopOnSolution)
			solveTime := time.Since(solveStart)

			log.Println("finished solving puzzle ", puzzle.ID, " in ", solveTime)
			log.Println(len(solutions), "solutions found for puzzle ", puzzle.ID)
			// log.Println(status)

			printMemUsage()

			//insert solutions into db
			err = tileio.InsertSolutions(db, puzzle.ID, &solutions)
			if err != nil {
				log.Fatal(err)
			}
			err = tileio.MarkPuzzle(db, puzzle.ID, solverID, solveTime, status, tilesPlaced)
			if err != nil {
				log.Fatal(err)
			}
			puzzlesSolved++
			solutions = nil
			runtime.GC()
			printMemUsage()
		}
		if len(puzzles) < batchSize {
			break
		}
	}
	log.Println("finished, solved ", puzzlesSolved, " puzzles")
}

func solveJobsFromDatabase(dbstring string, numTiles int, puzzleLimit int, batchSize int, solverID int, processTimeout int, puzzleTimeout int, stopOnSolution bool) {
	db := tileio.Open(dbstring)
	defer tileio.Close(db)

	puzzlesSolved := 0
	processEndTime := time.Now().Add(time.Duration(1000000000 * int64(processTimeout)))

	for puzzleLimit == 0 || puzzlesSolved < puzzleLimit {
		puzzles, err := tileio.GetNewJobs(db, batchSize, numTiles)
		if err != nil {
			log.Println(err)
			log.Fatal("whatever")
		}

		for _, puzzle := range puzzles {
			log.Println("start solving job", puzzle.JobID)
			solveStart := time.Now()
			solveEnd := solveStart.Add(time.Duration(1000000000 * int64(puzzleTimeout)))
			if processEndTime.Before(solveEnd) {
				solveEnd = processEndTime
			}
			solutions, status, tilesPlaced := tiling.SolveNaive(puzzle.BoardDims, *puzzle.Tiles, *puzzle.Start, *puzzle.Stop, solveEnd, stopOnSolution)
			solveTime := time.Since(solveStart)

			log.Println("finished solving job ", puzzle.JobID, " in ", solveTime)
			log.Println(len(solutions), "solutions found for puzzle ", puzzle.ID)
			// log.Println(status)

			// insert solutions into db
			err = tileio.InsertSolutions(db, puzzle.ID, &solutions)
			if err != nil {
				log.Fatal(err)
			}
			err = tileio.MarkJob(db, puzzle.JobID, solverID, solveTime, status, tilesPlaced)
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
