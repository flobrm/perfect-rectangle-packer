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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

// var inputPath = flag.String("inputpath", "", "input file with puzzles")

var numTiles = flag.Int("num_tiles", 0, "Solve only puzzles with this many tiles.")
var puzzleLimit = flag.Int("puzzle_limit", 0, "Solve at most N puzzles")
var batchSize = flag.Int("batch_size", 1, "How many puzzles should the program reserve at once")
var solverID = flag.Int("solver_id", 0, "Used to differentiate between different solvers and hardware")
var useJobs = flag.Bool("jobs", false, "If set solve jobs instead of full puzzles")
var dbstring = flag.String("dbstring", "tiler:tiler@(localhost:3306)/tiling", "Database connection string")
var processTimeout = flag.Int("process_timeout", 0, "Max time in seconds that the solver is allowed")
var puzzleTimeout = flag.Int("puzzle_timeout", 0, "Max time before a puzzle/job is interrupted")
var stopOnSolution = flag.Bool("stop_on_solution", false, "Stop the solver after finding the first solution")

// Optimization flags
var allSameSideNeighborCheck = flag.Bool("full_ssn_check", true, "set hierarchical same side neighbor check")
var oneLevelSSNCheck = flag.Bool("1level_ssn_check", false, "set one level same side neighbor check") //TODO Not implemented yet
var gapDetectionCheck = flag.Bool("gap_detection_check", true, "Enable gap detection, overrules more specific options")
var nextGapDetectionCheck = flag.Bool("next_gap_check", true, "check the next gap where a tile will be placed")
var allDownDetectionCheck = flag.Bool("all_down_gap_check", true, "check all normal gaps")
var leftSideGapCheck = flag.Bool("left_side_gaps_check", true, "check gaps from the left side to the frame top")
var totalGapAreaCheck = flag.Bool("total_gap_area_check", false, "check if the total gap area can be filled") //Turned of because it doesn't work or is never triggered
var forceFrameUpright = flag.Bool("force_frame_upright", true, "Rotate the frame, start, and stop so the shortest frame side is used as the width.")
var placementChoice = flag.String("placement_choice", "smallestGap", "The algorithm determining the position of the next tile. [lastGapAdded (default), smallestGap, bottomLeft]")

// File based multithreaded options
var numSolvers = flag.Int("workers", 1, "number of worker threads")
var processID = flag.String("processID", "1", "An identifier to be able to recognize output from multiple processes")
var jobsFile = flag.String("input_file", "", "File with puzzles/jobs")
var outputDir = flag.String("output_dir", "", "Directory where output should go")

func main() {
	flag.Parse()
	//validate placementChoice:
	if _, ok := tiling.PlacementOrderOptions[*placementChoice]; !ok {
		log.Fatal("Couldn't recognize placement_choice.")
	}

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
	}
	if *processTimeout == 0 {
		*processTimeout = 3600 * 24 * 365 // a year in seconds, could be any big number
	}
	if *puzzleTimeout == 0 {
		*puzzleTimeout = 3600 * 24 * 365 // a year in seconds, could be any big number
	}
	optimizationFlags := make(map[int]bool)
	optimizationFlags[tiling.FullSSNCheck] = *allSameSideNeighborCheck
	optimizationFlags[tiling.OneLevelSSN] = *oneLevelSSNCheck
	optimizationFlags[tiling.DoGapdetection] = *gapDetectionCheck
	optimizationFlags[tiling.OneGapDetection] = *nextGapDetectionCheck
	optimizationFlags[tiling.AllDownGapDetection] = *allDownDetectionCheck
	optimizationFlags[tiling.LeftGapDetection] = *leftSideGapCheck
	optimizationFlags[tiling.TotalGapAreaCheck] = *totalGapAreaCheck
	optimizationFlags[tiling.ForceFrameUpright] = *forceFrameUpright

	start := time.Now()

	if *jobsFile != "" {
		taskReader := tileio.NewPuzzleCSVReader(*jobsFile)
		//TODO setup output stuff, for now print to output
		//outputer
		// solveTasks(taskReader, *solverID, *processTimeout, *puzzleTimeout, *stopOnSolution, *processID, *outputDir,
		//*numSolvers, optimizationFlags, tiling.PlacementOrder[*placementChoice])
		solveConcurrentTasks(taskReader, *solverID, *processTimeout, *puzzleTimeout, *stopOnSolution, *processID,
			*outputDir, *numSolvers, optimizationFlags, tiling.PlacementOrderOptions[*placementChoice])
	}
	// fmt.Print(len(solveAsQas8()))
	// fmt.Println(len(solveTestCase()))

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

func solveConcurrentTasks(tasks tileio.PuzzleReader, solverID int, processTimeout int, puzzleTimeout int, stopOnSolution bool,
	processID string, outputDir string, workers int, optimizations map[int]bool, placementOrder int) {
	// parse options, determine endtime
	puzzlesSolved := 0
	activeWorkers := 0
	processEndTime := time.Now().Add(time.Duration(1000000000 * int64(processTimeout)))

	// for i in workers create and start worker
	fileWriters := make([]tileio.PuzzleResolutionWriter, workers)
	finishedJobsChan := make(chan int, workers) //one buffered channel of len workers
	// interrupter := make(chan string, 1) //all workers will receive the same interrupt

	for worker := 0; worker < workers; worker++ {
		//open files
		var err error
		statusFile := fmt.Sprintf("%s/%s_%d.status.csv", outputDir, processID, worker) //TODO zero pad worker
		solutionsFile := fmt.Sprintf("%s/%s_%d.solutions.csv", outputDir, processID, worker)
		fmt.Println(statusFile, solutionsFile)
		fileWriters[worker], err = tileio.NewPuzzleCSVWriter(statusFile, solutionsFile)
		if err != nil {
			log.Fatal("Could not open logging files: ", err)
		}
		//get puzzle
		puzzle, err := tasks.NextPuzzle()
		if err != nil {
			//TODO handle error
			log.Println("Couldn't read puzzle, assuming EOF:", err)
			break
		}
		go runWorker(finishedJobsChan, worker, solverID, puzzle, puzzleTimeout, processEndTime,
			stopOnSolution, fileWriters[worker], optimizations, placementOrder)
		activeWorkers++
	}

	//loop and wait around
	for activeWorkers > 0 {
		//if receive signal from worker
		//get resolutionWriter back from workerFunc (or its id perhaps)
		log.Println("ActiveWorkers: ", activeWorkers)
		select {
		case worker := <-finishedJobsChan:
			puzzlesSolved++
			activeWorkers--
			log.Println("Finished puzzle on ", worker, activeWorkers)

			if time.Now().Before(processEndTime) {
				//start new worker
				puzzle, err := tasks.NextPuzzle()
				if err != nil {
					//TODO handle error
					log.Println("Couldn't read puzzle:", err)
					fileWriters[worker].Close()
					continue
				}
				go runWorker(finishedJobsChan, worker, solverID, puzzle, puzzleTimeout, processEndTime, stopOnSolution,
					fileWriters[worker], optimizations, placementOrder)
				activeWorkers++
				log.Println("Started puzzle ", puzzle.JobID, " on worker ", worker)

			}
			//case interrupt
		}
	}
	log.Println("Finished puzzleSolving, with", puzzlesSolved, "done  ")
	log.Println(processEndTime.Sub(time.Now()).String(), "before end time")
}

func runWorker(out chan int, workerID int, solverID int, puzzle tileio.PuzzleDescription, puzzleTimeout int, processEndTime time.Time,
	stopOnSolution bool, resolutionWriter tileio.PuzzleResolutionWriter, optimizations map[int]bool, placementOrder int) {
	solveStart := time.Now()
	solveEnd := solveStart.Add(time.Duration(1000000000 * int64(puzzleTimeout)))
	if processEndTime.Before(solveEnd) {
		solveEnd = processEndTime
	}
	solutions, status, tilesPlaced, currentPlacement := tiling.SolveNaive(puzzle.Board, *puzzle.Tiles, *puzzle.Start,
		*puzzle.End, solveEnd, stopOnSolution, optimizations, placementOrder)
	solveTime := time.Since(solveStart)
	resolutionWriter.SaveSolutions(puzzle.PuzzleID, puzzle.JobID, &solutions)
	resolutionWriter.SaveStatus(&puzzle, status, tilesPlaced, solveTime, solverID, &currentPlacement)

	log.Println("finished solving job ", puzzle.JobID, "on worker", workerID, " in ", solveTime)
	log.Println(len(solutions), "solutions found for puzzle ", puzzle.PuzzleID)
	out <- workerID
	return
}

func solveTasks(tasks tileio.PuzzleReader, solverID int, processTimeout int, puzzleTimeout int, stopOnSolution bool,
	processID string, outputDir string, workers int, optimizationFlags map[int]bool, placementOrder int) {
	log.Println("starting solveTasks", solverID, workers)
	puzzlesSolved := 0
	processEndTime := time.Now().Add(time.Duration(1000000000 * int64(processTimeout)))

	var resolutionWriter tileio.PuzzleResolutionWriter
	var err error
	statusFile := fmt.Sprintf("%s/%s.status.csv", outputDir, processID)
	solutionsFile := fmt.Sprintf("%s/%s.solutions.csv", outputDir, processID)
	fmt.Println(statusFile, solutionsFile)
	resolutionWriter, err = tileio.NewPuzzleCSVWriter(statusFile, solutionsFile) //TODO check for error, close at the end
	defer resolutionWriter.Close()
	if err != nil {
		log.Println("Could not open logging files: ", err)
	}

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
		solutions, status, tilesPlaced, currentPlacement := tiling.SolveNaive(puzzle.Board, *puzzle.Tiles, *puzzle.Start,
			*puzzle.End, solveEnd, stopOnSolution, optimizationFlags, placementOrder)
		solveTime := time.Since(solveStart)

		resolutionWriter.SaveSolutions(puzzle.PuzzleID, puzzle.JobID, &solutions)
		resolutionWriter.SaveStatus(&puzzle, status, tilesPlaced, solveTime, solverID, &currentPlacement)

		log.Println("finished solving job ", puzzle.JobID, " in ", solveTime)
		log.Println(len(solutions), "solutions found for puzzle ", puzzle.PuzzleID)
		puzzlesSolved++
		// log.Fatal("quiting early")
	}
	log.Println("finished all puzzles")
	log.Println("finished, solved ", puzzlesSolved, " puzzles")
}

// func startTask(w *resolutionWriter) {

// }

func getDefaultOptimizations() map[int]bool {
	optimizations := map[int]bool{
		tiling.FullSSNCheck:        true,
		tiling.OneLevelSSN:         false,
		tiling.DoGapdetection:      true,
		tiling.OneGapDetection:     false,
		tiling.AllDownGapDetection: true,
		tiling.LeftGapDetection:    true,
		tiling.TotalGapAreaCheck:   true,
		tiling.ForceFrameUpright:   true,
	}
	return optimizations
}

func solveTestCase() map[string]int {
	board := core.Coord{X: 41, Y: 25}
	tiles := make([]core.Coord, 11)
	tileBytes := []byte("[{\"X\":22,\"Y\":14},{\"X\":20,\"Y\":6},{\"X\":20,\"Y\":3},{\"X\":20,\"Y\":2},{\"X\":17,\"Y\":1},{\"X\":15,\"Y\":11},{\"X\":14,\"Y\":13},{\"X\":10,\"Y\":5},{\"X\":7,\"Y\":6},{\"X\":7,\"Y\":5},{\"X\":6,\"Y\":1}]")
	json.Unmarshal(tileBytes, &tiles)
	result, _, _, _ := tiling.SolveNaive(board, tiles, nil, nil, time.Now().Add(time.Duration(1000000000*3600)),
		false, getDefaultOptimizations(), tiling.LastGapFirst)
	return result
}

func solveAsQas3() map[string]int {
	// build the almost square puzzle instance with 3 tiles
	var tiles [3]core.Coord
	for i := range tiles {
		tiles[2-i] = core.Coord{X: i + 2, Y: i + 1}
	}
	result, _, _, _ := tiling.SolveNaive(core.Coord{X: 5, Y: 4}, tiles[:], nil, nil,
		time.Now().Add(time.Duration(1000000000*3600)), false, getDefaultOptimizations(), tiling.LastGapFirst)
	return result
}

func solveAsQas8() map[string]int {
	// build the almost square puzzle instance with 8 tiles
	var tiles [8]core.Coord
	for i := range tiles {
		tiles[7-i] = core.Coord{X: i + 2, Y: i + 1}
	}
	result, _, steps, _ := tiling.SolveNaive(core.Coord{X: 15, Y: 16}, tiles[:], nil, nil,
		time.Now().Add(time.Duration(1000000000*3600)), false, getDefaultOptimizations(), tiling.LastGapFirst)
	fmt.Println("steps", steps)
	return result
}

func solveAsQas20() map[string]int {
	// build the almost square puzzle instance with 20 tiles
	var tiles [20]core.Coord
	for i := range tiles {
		tiles[19-i] = core.Coord{X: i + 2, Y: i + 1}
	}
	results, _, steps, _ := tiling.SolveNaive(core.Coord{X: 55, Y: 56}, tiles[:], nil, nil,
		time.Now().Add(time.Duration(1000000000*3600)), true, getDefaultOptimizations(), tiling.LastGapFirst)
	fmt.Println("steps", steps)
	return results
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
		solutions, _, _, _ := tiling.SolveNaive(puzzle.Board, *puzzle.Tiles, nil, nil,
			time.Now().Add(time.Duration(1000000000*3600)), false, getDefaultOptimizations(), tiling.LastGapFirst)
		log.Println("solved a puzzle")
		for _, solution := range solutions {
			//TODO write solutions
			fmt.Println(solution)
		}
	}
	log.Println("finished")
}

// A debug function to check memory
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
