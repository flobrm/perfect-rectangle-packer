package tiling

import (
	"fmt"
)

//Coord type to store an integer vector in
type Coord struct {
	X, Y int
}

/*Tile is a puzzle piece
 */
type Tile struct {
	W, H, X, Y, CurW, CurH int
	Placed, Turned         bool
}

//NewTile initializes tile
func NewTile(w, h int) (t Tile) {
	t.W = w
	t.CurW = w
	t.H = h
	t.CurH = h
	return
}

//Print is a Printer for Tile
func (t Tile) String() string {
	return fmt.Sprintf("Tile: %d, %d, Pos: %d, %d Rot: %t, Used: %t", t.W, t.H, t.X, t.Y, t.Turned, t.Placed)
}

//Place sets the X, Y and rotation of a Tile and sets Placed to true
func (t *Tile) Place(spot Coord, turned bool) {
	t.X = spot.X
	t.Y = spot.Y
	t.Turned = turned
	t.Placed = true
	if t.Turned {
		t.CurW = t.H
		t.CurH = t.W
	} else {
		t.CurW = t.W
		t.CurH = t.H
	}

}

//Remove marks a tile as not placed
func (t *Tile) Remove() {
	t.Placed = false
}

//GetNeighborSpots returns the positions bottomright and topleft of the tile (in that order)
func (t Tile) GetNeighborSpots() []Coord {
	spots := [2]Coord{
		Coord{X: t.X + t.CurW, Y: t.Y},
		Coord{X: t.X, Y: t.Y + t.CurH}}
	return spots[:]
}

func (t Tile) collides(b Tile) bool { //TODO pass by reference if necessary for speed
	if t.X >= b.X+b.CurW || b.X >= t.X+t.CurW {
		return false
	}
	if t.Y >= b.Y+b.CurH || b.Y >= t.Y+t.CurH {
		return false
	}
	return true
}

func (t Tile) posCollides(pos Coord) bool {
	if pos.X < t.X || pos.X >= t.X+t.CurW {
		return false
	}
	if pos.Y < t.Y || pos.Y >= t.Y+t.CurH {
		return false
	}
	return true
}
