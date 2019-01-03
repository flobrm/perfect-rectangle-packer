package tileio

import (
	"context"
	"database/sql"
	"fmt"
	"localhost/flobrm/tilingsolver/tiling"
	"log"

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
func GetNewPuzzles(db *sql.DB, numPuzzles int, numTiles int) []Puzzle {
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
	}
	resultRows := make([]puzzleRow, numPuzzles)
	rowsRead := 0
	for ; result.Next(); rowsRead++ {
		err = result.Scan(&resultRows[rowsRead].id, &resultRows[rowsRead].numTiles, &resultRows[rowsRead].boardWidth, &resultRows[rowsRead].boardHeight, &resultRows[rowsRead].tiles)
		if err != nil {
			log.Println("Error reading puzzle row:", err)
		}

		fmt.Println(resultRows[rowsRead], *resultRows[rowsRead].tiles)
	}
	resultRows = resultRows[:rowsRead] //TODO check off by one error

	err = transaction.Commit() //TODO

	return puzzles
}
