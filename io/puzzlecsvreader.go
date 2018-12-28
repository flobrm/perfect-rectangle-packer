package io

import (
	"encoding/csv"
	"fmt"
	"localhost/flobrm/tilingsolver/tiling"
	"log"
	"os"
	"strconv"
)

//PuzzleReader is an interface supplying all functions for getting puzzles input
type PuzzleReader interface {
	NextPuzzle() (PuzzleDescription, error)
}

//A PuzzleCSVReader accepts a csv filename and implements the PuzzleReader interface
type PuzzleCSVReader struct {
	reader *csv.Reader
	header map[string]int
}

//PuzzleDescription describes a tiling puzzle
type PuzzleDescription struct {
	Id    int
	Batch int
	Board tiling.Coord
	Tiles *[]tiling.Coord
}

//TileDescription describes what a tile is like
type TileDescription struct {
}

//NewPuzzleCSVReader opens a csv file and return an object that will reader puzzles 1 by 1
func NewPuzzleCSVReader(path string) PuzzleReader {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Couldn't open file ", path, err)
	}
	csvReader := csv.NewReader(file)
	if err != nil {
		log.Fatal("Couldn't make csvReader ", err)
	}

	header := make(map[string]int, 5)
	headerNames, err := csvReader.Read()
	if err != nil {
		log.Fatal("Couldn't read header ", err)
	}
	for i, name := range headerNames {
		header[name] = i
	}
	puzzleReader := PuzzleCSVReader{reader: csvReader}

	return puzzleReader

}

//NextPuzzle returns a PuzzleDescription, or err if there are no puzzles left in the file
func (r PuzzleCSVReader) NextPuzzle() (PuzzleDescription, error) {
	record, err := r.reader.Read()
	if err != nil {
		return PuzzleDescription{}, err
	}
	fmt.Print(record)

	tiles := make([]tiling.Coord, parseInt(record[r.header["num_tiles"]]))

	puzzle := PuzzleDescription{
		Id:    parseInt(record[r.header["id"]]),
		Batch: parseInt(record[r.header["batch"]]),
		Board: tiling.Coord{
			X: parseInt(record[r.header["board_width"]]),
			Y: parseInt(record[r.header["board_height"]])},
		Tiles: &tiles}

	return puzzle, nil
}

//parseInt converts string to int or dies
func parseInt(s string) int {
	myInt, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal("tried converting", s, "to int")
	}
	return myInt
}
