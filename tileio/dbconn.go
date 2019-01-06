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
func Open() *sql.DB {
	connectstring := "tiler:tiler@(localhost:3306)/tiling"

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
	NumTiles  int
	BoardDims tiling.Coord
	Tiles     *[]tiling.Coord
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

// InsertSolutions adds solutions as json to the solutions table
func InsertSolutions(db *sql.DB, puzzleID int, solverID int, duration time.Duration, solutions *map[string][]tiling.Tile) error {

	if len(*solutions) > 0 {
		puzzleIDString := strconv.Itoa(puzzleID)
		query := "INSERT IGNORE INTO tiling.solutions (puzzles_id, tiles_hash, tiles) VALUES "
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

	query := "UPDATE tiling.puzzles SET status = 'solved', solver_id = ?, duration = ? WHERE id = ? "
	_, err := db.Exec(query, solverID, duration.Nanoseconds(), puzzleID)
	if err != nil {
		log.Println("inserted all tiles, but failed updating puzzle status: ", err)
		return errors.New("failedSolutionsUpdate")
	}

	return nil
}
