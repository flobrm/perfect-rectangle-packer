package tileio

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"localhost/flobrm/tilingsolver/core"
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
	reader     *csv.Reader
	header     map[string]int
	lineNumber int
}

//PuzzleDescription describes a tiling puzzle
type PuzzleDescription struct {
	ID    int
	Batch int
	Board core.Coord
	Tiles *[]core.Coord
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
	csvReader.TrimLeadingSpace = true

	header := make(map[string]int, 5)
	headerNames, err := csvReader.Read()
	if err != nil {
		log.Fatal("Couldn't read header ", err)
	}
	for i, name := range headerNames {
		header[name] = i
	}
	puzzleReader := PuzzleCSVReader{reader: csvReader, header: header}

	return puzzleReader
}

//NextPuzzle reads lines until it encounters a correct puzzle and returns a PuzzleDescription,
//or err if there are no puzzles left in the file
func (r PuzzleCSVReader) NextPuzzle() (PuzzleDescription, error) {
	for true {
		record, err := r.reader.Read()
		r.lineNumber++
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error reading puzzle at line:", r.lineNumber, "error:", err)
			continue
		}
		fmt.Print(record)

		tiles := make([]core.Coord, parseInt(record[r.header["num_tiles"]]))
		err = json.Unmarshal([]byte(record[r.header["tiles"]]), &tiles)
		if err != nil {
			fmt.Println("error reading tiles at line:", r.lineNumber, "error:", err)
			continue
		}

		puzzle := PuzzleDescription{
			ID:    parseInt(record[r.header["id"]]),
			Batch: parseInt(record[r.header["batch"]]),
			Board: core.Coord{
				X: parseInt(record[r.header["board_width"]]),
				Y: parseInt(record[r.header["board_height"]])},
			Tiles: &tiles}
		return puzzle, nil
	}
	return PuzzleDescription{}, io.EOF
}

//parseInt converts string to int or dies
func parseInt(s string) int {
	myInt, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal("tried converting", s, "to int")
	}
	return myInt
}

//Start of puzzleJSONReader stuff

//PuzzleJSONReader reads files where each line describes a puzzle in JSON. The full file doesn't have to be in JSON
type PuzzleJSONReader struct {
	reader     *bufio.Scanner
	lineNumber int
}

//NewPuzzleJSONReader opens a file and return an object that will read puzzles 1 by 1
func NewPuzzleJSONReader(path string) PuzzleReader {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Couldn't open file ", path, err)
	}

	reader := bufio.NewScanner(file)
	if err != nil {
		log.Fatal("Couldn't make scanner ", err)
	}

	puzzleReader := PuzzleJSONReader{reader: reader, lineNumber: 0}

	return puzzleReader
}

//NextPuzzle reads lines until it encounters a valid
func (r PuzzleJSONReader) NextPuzzle() (PuzzleDescription, error) {
	for r.reader.Scan() {
		line := r.reader.Bytes()
		r.lineNumber++

		puzzle := PuzzleDescription{}
		err := json.Unmarshal(line, &puzzle)
		if err == nil {
			return puzzle, nil
		}
		fmt.Println("lineNumber:", r.lineNumber, "error:", err)
	}
	return PuzzleDescription{}, io.EOF
}
