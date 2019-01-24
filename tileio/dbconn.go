package tileio

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"localhost/flobrm/tilingsolver/tiling"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" //import mysql driver, blank because the interface is "database/sql"
)

//Open gets a database object
func Open(connectstring string) *sql.DB {
	db, err := sql.Open("mysql", connectstring)
	if err != nil {
		log.Println(err)
		log.Fatal("Failed to get db object")
	}
	err = db.Ping()
	if err != nil {
		log.Println(err)
		log.Fatal("Failed to connect to database")
	}
	return db
}

//Close closes a db
func Close(db *sql.DB) {
	err := db.Close()
	log.Println("error closing db: ", err)
}

type puzzleRow struct {
	id                      int
	numTiles                int
	boardWidth, boardHeight int
	tiles                   *string
}

//Puzzle store important puzzle attributes as they are fetched from the database
type Puzzle struct {
	ID        int
	JobID     int
	NumTiles  int
	BoardDims tiling.Coord
	Tiles     *[]tiling.Coord
	Start     *[]TilePlacement
	Stop      *[]TilePlacement
}

//TilePlacement store an index and rotation of a tile in a puzzle
type TilePlacement struct {
	Idx int  //Tile index
	Rot bool //false is flat, true is upright
}

//GetNewPuzzles returns
func GetNewPuzzles(db *sql.DB, numPuzzles int, numTiles int) ([]Puzzle, error) {
	//TODO random expo backoff

	context := context.Background()
	// txOptions := sql.TxOptions{}}
	transaction, err := db.BeginTx(context, nil)
	if err != nil {
		log.Fatal("Couldn't start transaction:", err)
	}

	numTilesQuery := ""
	if numTiles != 0 {
		numTilesQuery = fmt.Sprintf("and num_tiles =  %d ", numTiles)
	}

	query := fmt.Sprint("SELECT id, num_tiles, board_width, board_height, tiles FROM puzzles WHERE status = 'new' ", numTilesQuery, "LIMIT ?  FOR UPDATE")
	statement, err := transaction.Prepare(query)

	// query := "SELECT id, num_tiles, board_width, board_height, tiles FROM puzzles WHERE id < 10 FOR UPDATE"
	result, err := statement.Query(numPuzzles)
	defer result.Close()
	if err != nil {
		log.Println("Error fetching puzzles: ", err)
		transaction.Rollback()
		//TODO return error
	}
	puzzles := make([]Puzzle, numPuzzles)
	rowsRead := 0
	for ; result.Next(); rowsRead++ {
		var tileJSON []byte
		err = result.Scan(&puzzles[rowsRead].ID, &puzzles[rowsRead].NumTiles, &puzzles[rowsRead].BoardDims.X,
			&puzzles[rowsRead].BoardDims.Y, &tileJSON)
		if err != nil {
			log.Println("Error reading puzzle row:", err)
			log.Println("skipping unknown puzzle")
			rowsRead--
			continue
		}
		tiles := make([]tiling.Coord, puzzles[rowsRead].NumTiles)
		err = json.Unmarshal(tileJSON, &tiles)
		if err != nil {
			log.Println("Error reading tiles json:", err)
			log.Println("skipping puzzle id", puzzles[rowsRead].ID)
			rowsRead--
		}
		puzzles[rowsRead].Tiles = &tiles
	}
	puzzles = puzzles[:rowsRead] //TODO check off by one error

	//Now start updating puzzles to 'busy'
	ids := make([]string, len(puzzles))
	for i, puzzle := range puzzles {
		ids[i] = strconv.Itoa(puzzle.ID)
	}
	idsString := strings.Join(ids, ",")
	query = fmt.Sprintf("UPDATE puzzles SET status = 'busy' WHERE id IN (%s)", idsString)
	log.Println(query)
	_, err = transaction.Exec(query)
	if err != nil {
		log.Println("Error updating puzzle status:", err)
		err = transaction.Rollback()
		return puzzles, errors.New("rollback")
	}
	err = transaction.Commit()
	if err != nil {
		log.Println("error commiting busy puzzles", err)
		err = transaction.Rollback()
		if err != nil {
			log.Fatal("Also error rolling back ", err)
		}
		return puzzles, errors.New("rollback")
	}

	return puzzles, nil
}

//GetNewJobs is a function that returns puzzles
func GetNewJobs(db *sql.DB, numPuzzles int, numTiles int) ([]Puzzle, error) {
	//TODO random expo backoff

	context := context.Background()
	// txOptions := sql.TxOptions{}}
	transaction, err := db.BeginTx(context, nil)
	if err != nil {
		log.Fatal("Couldn't start transaction:", err)
	}

	numTilesQuery := ""
	if numTiles != 0 {
		numTilesQuery = fmt.Sprintf("and num_tiles =  %d ", numTiles)
	}

	query := fmt.Sprint("SELECT p.id, j.id, p.num_tiles, p.board_width, p.board_height, p.tiles, j.start_size, j.end_size, j.start_state, j.end_state FROM jobs j "+
		" JOIN puzzles p on j.puzzle_id = p.id "+
		" WHERE j.status = 'new' ", numTilesQuery, "ORDER BY j.id LIMIT ?  FOR UPDATE")
	statement, err := transaction.Prepare(query)

	result, err := statement.Query(numPuzzles)
	defer result.Close()
	if err != nil {
		log.Println("Error fetching puzzles: ", err)
		transaction.Rollback()
		//TODO return error
	}
	puzzles := make([]Puzzle, numPuzzles)
	rowsRead := 0
	for ; result.Next(); rowsRead++ {
		var tileJSON []byte
		var startJSON []byte
		var stopJSON []byte
		var startSize int
		var endSize int
		err = result.Scan(&puzzles[rowsRead].ID, &puzzles[rowsRead].JobID, &puzzles[rowsRead].NumTiles, &puzzles[rowsRead].BoardDims.X,
			&puzzles[rowsRead].BoardDims.Y, &tileJSON, &startSize, &endSize, &startJSON, &stopJSON)
		if err != nil {
			log.Println("Error reading puzzle row:", err)
			log.Println("skipping unknown puzzle")
			rowsRead--
			continue
		}
		// fmt.Println(string(startJSON), string(stopJSON), string(tileJSON))
		tiles := make([]tiling.Coord, puzzles[rowsRead].NumTiles)
		start := make([]TilePlacement, startSize)
		stop := make([]TilePlacement, endSize) //TODO handle 0 length stop or start
		err = json.Unmarshal(tileJSON, &tiles)
		if err == nil {
			err = json.Unmarshal(startJSON, &start)
		}
		if err == nil {
			err = json.Unmarshal(stopJSON, &stop)
		}
		if err != nil {
			log.Println("Error reading tiles json:", err)
			log.Println("skipping puzzle id", puzzles[rowsRead].ID)
			rowsRead--
		}
		puzzles[rowsRead].Tiles = &tiles
		puzzles[rowsRead].Start = &start
		puzzles[rowsRead].Stop = &stop
	}
	puzzles = puzzles[:rowsRead] //TODO check off by one error

	//Now start updating puzzles to 'busy'
	ids := make([]string, len(puzzles))
	for i, puzzle := range puzzles {
		ids[i] = strconv.Itoa(puzzle.JobID)
	}
	idsString := strings.Join(ids, ",")
	query = fmt.Sprintf("UPDATE jobs SET status = 'busy' WHERE id IN (%s)", idsString)
	log.Println(query)
	_, err = transaction.Exec(query)
	if err != nil {
		log.Println("Error updating job status:", err)
		err = transaction.Rollback()
		return puzzles, errors.New("rollback")
	}
	err = transaction.Commit()
	if err != nil {
		log.Println("error commiting busy jobs", err)
		err = transaction.Rollback()
		if err != nil {
			log.Fatal("Also error rolling back ", err)
		}
		return puzzles, errors.New("rollback")
	}

	return puzzles, nil
}

// InsertSolutions adds solutions as json to the solutions table
func InsertSolutions(db *sql.DB, puzzleID int, solutions *map[string][]tiling.Tile) error {

	if len(*solutions) > 0 {
		puzzleIDString := strconv.Itoa(puzzleID)
		query := "INSERT IGNORE INTO solutions (puzzle_id, tiles_hash, tiles) VALUES "
		//TODO add key index to (puzzleId,tiles)
		var values []interface{}
		args := make([]string, len(*solutions))[:0]

		for key := range *solutions {
			hasher := sha1.New()
			hasher.Write([]byte(key))
			hashString := hex.EncodeToString(hasher.Sum(nil))
			values = append(values, puzzleIDString, hashString, key)
			args = append(args, "(?,?,?)")
		}

		query += strings.Join(args, ",")
		_, err := db.Exec(query, values...)
		if err != nil {
			log.Println("error inserting solutions: ", err)
			log.Fatal("Giving up on all solutions")
			//TODO return error
		}
	}
	return nil
}

//MarkPuzzleSolved set the status to solved and updates the solver and duration
func MarkPuzzleSolved(db *sql.DB, puzzleID int, solverID int, duration time.Duration) error {
	query := "UPDATE puzzles SET status = 'solved', solver_id = ?, duration = ? WHERE id = ? "
	_, err := db.Exec(query, solverID, duration.Nanoseconds(), puzzleID)
	if err != nil {
		log.Println("inserted all tiles, but failed updating puzzle status: ", err)
		return errors.New("failedSolutionsUpdate")
	}
	return nil
}

//MarkPuzzle set the status to solved and updates the solver and duration
func MarkPuzzle(db *sql.DB, puzzleID int, solverID int, duration time.Duration, status string, tilesPlaced uint) error {
	query := "UPDATE puzzles SET status = ?, solver_id = ?, duration = ?, tiles_placed = ? WHERE id = ? "
	_, err := db.Exec(query, status, solverID, duration.Nanoseconds(), tilesPlaced, puzzleID)
	if err != nil {
		log.Println("inserted all tiles, but failed updating puzzle status: ", err)
		return errors.New("failedSolutionsUpdate")
	}
	return nil
}

//MarkJob set the status to solved and updates the solver and duration
func MarkJob(db *sql.DB, jobID int, solverID int, duration time.Duration, status string, tilesPlaced uint) error {
	query := "UPDATE jobs SET status = ?, solver_id = ?, duration = ?, tiles_placed = ? WHERE id = ? "
	_, err := db.Exec(query, status, solverID, duration.Nanoseconds(), tilesPlaced, jobID)
	if err != nil {
		log.Println("inserted all tiles, but failed updating puzzle status: ", err)
		return errors.New("failedSolutionsUpdate")
	}
	return nil
}
