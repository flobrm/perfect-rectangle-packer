package tileio

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"localhost/flobrm/tilingsolver/tiling"
	"log"
	"strconv"
	"strings"

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

		fmt.Println(puzzles[rowsRead], *puzzles[rowsRead].Tiles)
	}
	puzzles = puzzles[:rowsRead] //TODO check off by one error

	//Now start updating puzzles to 'busy'
	ids := make([]string, len(puzzles))
	for i, puzzle := range puzzles {
		ids[i] = strconv.Itoa(puzzle.ID)
	}
	idsString := strings.Join(ids, ",")
	query = fmt.Sprintf("UPDATE puzzles SET status = 'busy' WHERE id IN (%s)", idsString)
	fmt.Println(query)
	updateResult, err := transaction.Exec(query)
	if err != nil {
		log.Println("Error updating puzzle status:", err)
		err = transaction.Rollback()
		return puzzles, errors.New("rollback")
	}
	fmt.Println(updateResult)
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
