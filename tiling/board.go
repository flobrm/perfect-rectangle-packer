package tiling

import "localhost/flobrm/tilingsolver/core"

//Board stores the board and everything placed on it
type Board struct {
	Size          core.Coord   //width and hight of the board
	Tiles         [](*Tile)    //All the placed tiles
	Candidates    []core.Coord //Candidate positions for next placement
	lastCollision *Tile
}

//NewBoard inits a board, including candidates
func NewBoard(boardDims core.Coord, tiles []Tile) Board {
	myTiles := make([](*Tile), len(tiles))
	candidates := append(make([]core.Coord, 0), core.Coord{X: 0, Y: 0})
	return Board{
		Size:       core.Coord{X: boardDims.X, Y: boardDims.Y},
		Tiles:      myTiles[:0],
		Candidates: candidates}
	//Candidates: make([]core.Coord, len(tiles)+1)}
}

func (b *Board) addCandidate(newCand core.Coord) {
	b.Candidates = append(b.Candidates, newCand)
}

//Min returns the min of two integers, why the fuck do I have to define this
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//Fits checks if a tile fits the board at the next position to fill
func (b *Board) Fits(tile Tile, turned bool) bool {
	// fmt.Println(tile)
	candIndex := len(b.Candidates) - 1
	pos := b.Candidates[candIndex]
	tile.Place(pos, turned)
	if !b.tileFitsBoard(&tile) {
		tile.Remove()
		return false
	}
	//Check if the tile is a corner piece smaller than the lower left corner tile
	if len(b.Tiles) > 0 && tile.Index < b.Tiles[0].Index {
		corner := b.isCornerTile(&tile)
		if corner != noCorner && corner != bottomLeftCorner {
			return false
		}
	}

	if b.lastCollision != nil && b.lastCollision.Placed {
		if tile.collides(b.lastCollision) {
			tile.Remove()
			return false
		}
	}

	for _, tile2 := range b.Tiles {
		if tile.collides(tile2) {
			tile.Remove()
			b.lastCollision = tile2
			return false
		}
	}
	return true
}

//tileFitsBoard checks if the tile with it's internal rotation and position fits inside the board
func (b *Board) tileFitsBoard(tile *Tile) bool {
	if tile.X+tile.CurW > b.Size.X || tile.Y+tile.CurH > b.Size.Y || //check top and right side
		tile.X < 0 || tile.Y < 0 { //check bottom and left side
		return false
	}
	return true
}

//PlaceTile registers a tile as placed on the current spot to fill, Fits should be called first
func (b *Board) PlaceTile(tile *Tile, turned bool) {
	candIndex := len(b.Candidates) - 1
	pos := b.Candidates[candIndex]
	tile.Place(pos, turned)
	b.Candidates = b.Candidates[:candIndex] //remove last candidate
	b.addCandidates(*tile)
	b.Tiles = append(b.Tiles, tile)
}

func (b *Board) addCandidates(tile Tile) {
	newCands := tile.GetNeighborSpots()

	//first add the new one on top
	//Only add this candidate if it's in a corner
	//Assume it is empty, because we fill from bottom to top
	if newCands[1].Y < b.Size.Y { //TODO check if position is free
		if newCands[1].X == 0 { //corner with the wall
			b.addCandidate(newCands[1])
		} else { //check for corner with a tile on the left
			collision, _ := b.posCollides(core.Coord{X: newCands[1].X - 1, Y: newCands[1].Y})
			if collision {
				b.addCandidate(newCands[1])
			}
		}
	}

	//then add new one on the right
	if newCands[0].X < b.Size.X {
		var lastCollision *Tile
		// check for positions
		maxY := newCands[0].Y + tile.CurH
		for y := newCands[0].Y; y < maxY; y++ {
			newCands[0].Y = y
			//if there was a collision before, check if the same tile collides again
			if lastCollision != nil {
				if lastCollision.posCollides(&newCands[0]) {
					continue
				}
			}
			// no collision with last collider, check all tiles to be sure
			collides, collider := b.posCollides(newCands[0])
			lastCollision = collider
			if collides {
				continue
			}
			//no collisions, add it to candidates
			b.addCandidate(newCands[0])
			break
			//b.Candidates = append(b.Candidates[:], newCands[0])
		}
	}
}

func (b *Board) posCollides(pos core.Coord) (collides bool, collider *Tile) {
	collides = false
	collider = nil
	for _, tile := range b.Tiles {
		if tile.posCollides(&pos) {
			collider = tile
			collides = true
			return
		}
	}
	return
}

//RemoveLastTile removes a tile and resets it's candidate positions
func (b *Board) RemoveLastTile() *Tile {
	tile := *b.Tiles[len(b.Tiles)-1]
	b.removeCandidates(tile)
	b.addCandidate(core.Coord{X: tile.X, Y: tile.Y})
	tile.Remove()
	b.Tiles = b.Tiles[:len(b.Tiles)-1]

	return &tile
}

func (b *Board) removeCandidates(tile Tile) {
	if len(b.Candidates) == 0 {
		return
	}
	cand := b.Candidates[len(b.Candidates)-1]
	if isRightCandidate(cand, tile) || isTopCandidate(cand, tile) {
		b.Candidates = b.Candidates[:len(b.Candidates)-1]
		if len(b.Candidates) == 0 {
			return
		}
		cand = b.Candidates[len(b.Candidates)-1]
		if isRightCandidate(cand, tile) || isTopCandidate(cand, tile) {
			b.Candidates = b.Candidates[:len(b.Candidates)-1]
		}
	}
}

func isRightCandidate(cand core.Coord, tile Tile) bool {
	if cand.X == tile.X+tile.CurW {
		if cand.Y >= tile.Y && cand.Y < tile.Y+tile.CurH {
			return true
		}
	}
	return false
}

func isTopCandidate(cand core.Coord, tile Tile) bool {
	if cand.Y == tile.Y+tile.CurH {
		if cand.X >= tile.X && cand.X < tile.X+tile.CurW {
			return true
		}
	}
	return false
}

//Functions to flip new tiles horizontally or vertically

//Ways to consistently refer to corners of a board or tile
const (
	noCorner          = iota
	bottomLeftCorner  = iota
	bottomRightCorner = iota
	topLeftCorner     = iota
	topRightCorner    = iota
)

//GetCanonicalSolution returns a slice of tiles as placed on the current board
//but flips the board so the corner with the largest tile is bottom right.
func (b *Board) GetCanonicalSolution(tiles *[]Tile) int {
	var largestCornerTile *Tile
	var largestCorner = bottomLeftCorner
	//first determine what the largest corner is
	for i, tile := range *tiles {
		corner := b.isCornerTile(&tile)
		if corner != noCorner {
			if largestCornerTile == nil {
				largestCornerTile = &(*tiles)[i]
				largestCorner = corner
			} else if tile.W > largestCornerTile.W ||
				tile.W == largestCornerTile.W && tile.Y > largestCornerTile.Y {
				largestCornerTile = &(*tiles)[i]
				largestCorner = corner
			}
		}
	}

	//Second flip the tiles depending on the largest corner
	if largestCorner == bottomLeftCorner {
		//Do nothing
	} else if largestCorner == bottomRightCorner {
		b.flipTilesHorizontally(tiles)
		return 1
	} else if largestCorner == topLeftCorner {
		b.flipTilesVertically(tiles)
		return 1
	} else { //TOP_RIGHT_CORNER
		b.flipTilesHorizontally(tiles)
		b.flipTilesVertically(tiles)
		return 1
	}
	return 0
}

//isCornerTile checks if t is a corner on the current board.
//It returns one of the constants ending in CORNER.
func (b *Board) isCornerTile(t *Tile) int {
	if t.X == 0 && t.Y == 0 {
		return bottomLeftCorner
	} else if t.Y == 0 && t.X+t.CurW == b.Size.X {
		return bottomRightCorner
	} else if t.X == 0 && t.Y+t.CurH == b.Size.Y {
		return topLeftCorner
	} else if t.X+t.CurW == b.Size.X && t.Y+t.CurH == b.Size.Y {
		return topRightCorner
	} else {
		return noCorner
	}
}

func (b *Board) flipTilesHorizontally(tiles *[]Tile) *[]Tile {
	for i := range *tiles {
		// tile2.X = b.Size.X - tile2.X - tile2.CurW
		(*tiles)[i].X = b.Size.X - (*tiles)[i].X - (*tiles)[i].CurW
	}

	return tiles
}

func (b *Board) flipTilesVertically(tiles *[]Tile) *[]Tile {
	for i := range *tiles {
		(*tiles)[i].Y = b.Size.Y - (*tiles)[i].Y - (*tiles)[i].CurH
	}
	return tiles
}
