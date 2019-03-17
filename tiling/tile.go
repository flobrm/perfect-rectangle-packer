package tiling

import (
	"encoding/json"
	"fmt"
	"localhost/flobrm/tilingsolver/core"
)

/*Tile is a puzzle piece
 */
type Tile struct {
	W, H, X, Y int
	CurW       int  `json:"-"`
	CurH       int  `json:"-"`
	Placed     bool `json:"-"`
	Turned     bool `json:"T"`
	Index      int  `json:"-"`
}

// TilePair is a way to store two pairs, and keep track of them
type TilePair struct {
	a, b *Tile
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
	return fmt.Sprintf("Tile: %d, %d, Pos: %d, %d Rot: %t, Used: %t, Idx: %d", t.W, t.H, t.X, t.Y, t.Turned, t.Placed, t.Index)
}

//Place sets the X, Y and rotation of a Tile and sets Placed to true
func (t *Tile) Place(spot core.Coord, turned bool) {
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
func (t Tile) GetNeighborSpots() []core.Coord {
	spots := [2]core.Coord{
		core.Coord{X: t.X + t.CurW, Y: t.Y},
		core.Coord{X: t.X, Y: t.Y + t.CurH}}
	return spots[:]
}

func (t *Tile) collides(b *Tile) bool {
	if t.X >= b.X+b.CurW || b.X >= t.X+t.CurW {
		return false
	}
	if t.Y >= b.Y+b.CurH || b.Y >= t.Y+t.CurH {
		return false
	}
	return true
}

func (t *Tile) posCollides(pos *core.Coord) bool {
	if pos.X < t.X || pos.X >= t.X+t.CurW {
		return false
	}
	if pos.Y < t.Y || pos.Y >= t.Y+t.CurH {
		return false
	}
	return true
}

// func (t *Tile) Equals(o *Tile) bool {
// 	return t.W == o.W && t.H == o.H && t.X == o.X && t.Y == o.Y && t.Turned == o.Turned
// }

// type TileSliceSet struct {
// 	slice [][]Tile
// }

// func (slice *[]Tile) Equals(other *[]Tile)

// func (set *TileSliceSet) Add(tileSlice []Tile) {

// }

// func (set *TileSliceSet) Contains(tileSlice []Tile) bool {

// }

//TileSliceToJSON returns a JSON string encoding the width, height, and positioning data of a tile
func TileSliceToJSON(tiles []Tile) string {
	result, _ := json.Marshal(tiles) //TODO don't ignore error
	return string(result)
}
