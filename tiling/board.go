package tiling

//Board stores the board and everything placed on it
type Board struct {
	Size          Coord     //width and hight of the board
	Tiles         [](*Tile) //All the placed tiles
	Candidates    []Coord   //Candidate positions for next placement
	lastCollision *Tile
}

//NewBoard inits a board, including candidates
func NewBoard(boardDims Coord, tiles []Tile) Board {
	myTiles := make([](*Tile), len(tiles))
	candidates := append(make([]Coord, 0), Coord{X: 0, Y: 0})
	return Board{
		Size:       Coord{boardDims.X, boardDims.Y},
		Tiles:      myTiles[:0],
		Candidates: candidates}
	//Candidates: make([]Coord, len(tiles)+1)}
}

func (b *Board) addCandidate(newCand Coord) {
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

	if b.lastCollision != nil {
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
	//TODO store tile for easy removal
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
			collision, _ := b.posCollides(Coord{X: newCands[1].X - 1, Y: newCands[1].Y})
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

func (b *Board) posCollides(pos Coord) (collides bool, collider *Tile) {
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
	b.addCandidate(Coord{tile.X, tile.Y})
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

func isRightCandidate(cand Coord, tile Tile) bool {
	if cand.X == tile.X+tile.CurW {
		if cand.Y >= tile.Y && cand.Y < tile.Y+tile.CurH {
			return true
		}
	}
	return false
}

func isTopCandidate(cand Coord, tile Tile) bool {
	if cand.Y == tile.Y+tile.CurH {
		if cand.X >= tile.X && cand.X < tile.X+tile.CurW {
			return true
		}
	}
	return false
}
