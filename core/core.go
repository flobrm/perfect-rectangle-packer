package core

// This package should contain all datatypes that are used to exchange data between different tiling packages

//Coord type to store an integer vector in
type Coord struct {
	X, Y int
}

//TilePlacement store an index and rotation of a tile in a puzzle
type TilePlacement struct {
	Idx int  //Tile index
	Rot bool //false is flat, true is upright
}
