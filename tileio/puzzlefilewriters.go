package tileio

import (
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"localhost/flobrm/tilingsolver/core"
	"log"
	"os"
	"strconv"
	"time"
)

// PuzzleResolutionWriter is an interface with a number of functions to write the status of your puzzles.
type PuzzleResolutionWriter interface {
	Close()
	SaveSolutions(puzzleID int, jobID int, solutions *map[string]int) error
	SaveStatus(puzzle *PuzzleDescription, status string, tilesPlaced uint, solveTime time.Duration,
		solverID int, placements *[]core.TilePlacement) error
}

// PuzzleCSVWriter keeps track of outputfiles, and implements PuzzleResolutionWriter
type PuzzleCSVWriter struct {
	statusFile    *os.File
	solutionsFile *os.File
}

// NewPuzzleCSVWriter opens two files for writing and return a PuzzleCSVWriter with them.
func NewPuzzleCSVWriter(statusFilename string, SolutionsFilename string) (*PuzzleCSVWriter, error) {
	//TODO add header if statusFile doesn't have one already
	statusFile, err := os.OpenFile(statusFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.FileMode(0666))
	if err != nil {
		log.Println("Can't open statusFile ", err.Error())
		return nil, err
	}
	//TODO add header if solutionsfile doesn't have one already
	solutionsFile, err := os.OpenFile(SolutionsFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.FileMode(0666))
	if err != nil {
		log.Println("Can't open statusFile ", err.Error())
		return nil, err
	}
	statusFile.WriteString("job_id,puzzle_id,status,tiles_placed,duration,solver_id,current_state\n")
	solutionsFile.WriteString("puzzle_id,job_id,tiles,tiles_hash\n")
	return &PuzzleCSVWriter{statusFile: statusFile, solutionsFile: solutionsFile}, nil
}

//Close closes all filedescriptors
func (w *PuzzleCSVWriter) Close() {
	w.statusFile.Close()
	w.solutionsFile.Close()
}

//SaveSolutions appends solutions to the file w.solutionsFile
func (w *PuzzleCSVWriter) SaveSolutions(puzzleID int, jobID int, solutions *map[string]int) error {

	writer := csv.NewWriter(w.solutionsFile)
	for tiles := range *solutions {
		//write puzzleID, jobID, tiles, hash
		hasher := sha1.New()
		hasher.Write([]byte(tiles))
		hashString := hex.EncodeToString(hasher.Sum(nil))

		err := writer.Write([]string{strconv.Itoa(puzzleID), strconv.Itoa(jobID), tiles, hashString})
		if err != nil {
			return err
		}
	}
	writer.Flush()
	err := w.solutionsFile.Sync() //just to be sure this gets written to disk
	return err
}

//SaveStatus writes the results of a job to a file
func (w *PuzzleCSVWriter) SaveStatus(puzzle *PuzzleDescription, status string, tilesPlaced uint, solveTime time.Duration,
	solverID int, placements *[]core.TilePlacement) error {

	writer := csv.NewWriter(w.statusFile)

	placementString := ""
	if placements != nil && len(*placements) != 0 {
		placementBytes, err := json.Marshal(placements)
		if err != nil {
			log.Fatal("Error marshalling placement: ", placements, err)
		}
		placementString = string(placementBytes)
	}

	writer.Write([]string{
		strconv.Itoa(puzzle.JobID),
		strconv.Itoa(puzzle.PuzzleID),
		status,
		strconv.FormatUint(uint64(tilesPlaced), 10),
		strconv.FormatInt(solveTime.Nanoseconds(), 10),
		strconv.Itoa(solverID),
		placementString})

	writer.Flush()
	err := w.statusFile.Sync()
	return err
}
