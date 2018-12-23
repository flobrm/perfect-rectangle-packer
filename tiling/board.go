package tiling

//Board stores the board and everything placed on it
type Board struct {
	Size       Coord     //width and hight of the board
	Tiles      [](*Tile) //All the placed tiles
	Candidates []Coord   //Candidate positions for next placement
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
	if len(b.Candidates) == 0 {
		b.Candidates = append(b.Candidates, newCand)
		return
	}

	added := false
	for i := len(b.Candidates) - 1; i >= 0; i-- {
		candidate := b.Candidates[i]
		if (newCand.Y < candidate.Y) ||
			(newCand.Y == candidate.Y && newCand.X < candidate.X) {
			// temp := append(b.Candidates[:i+1], newCand)

			b.Candidates = append(b.Candidates[:Min(len(b.Candidates), i+2)], b.Candidates[Min(len(b.Candidates)-1, i+1):]...)
			b.Candidates[i+1] = newCand
			// last := b.Candidates[i+1:] //b.Candidates[Min(len(b.Candidates), i+2):]...
			// b.Candidates = append(b.Candidates[:i+1], newCand)
			// b.Candidates = append(b.Candidates[:], last...)
			added = true
			break
		}
	}
	if !added { //All the other candidates were smaller
		b.Candidates = append([]Coord{newCand}, b.Candidates[:]...)
	}
}

//Min returns the min of two integers, why the fuck do I have to define this
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//Fits checks if a tile fits the board at the next position to fill
func (b Board) Fits(tile Tile, turned bool) bool {
	// fmt.Println(tile)
	candIndex := len(b.Candidates) - 1
	pos := b.Candidates[candIndex]
	tile.Place(pos, turned)
	if !b.tileFitsBoard(tile) {
		tile.Remove()
		return false
	}
	for _, tile2 := range b.Tiles { //TODO cache last collided tile, and check that one first
		if tile.collides(*tile2) {
			tile.Remove()
			return false
		}
	}
	return true
}

//tileFitsBoard checks if the tile with it's internal rotation and position fits inside the board
func (b Board) tileFitsBoard(tile Tile) bool {
	if tile.X < 0 || tile.Y < 0 || //check bottom and left side
		tile.X+tile.CurW > b.Size.X || tile.Y+tile.CurH > b.Size.Y { //check top and right side
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
				if lastCollision.posCollides(newCands[0]) {
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

func (b Board) posCollides(pos Coord) (collides bool, collider *Tile) {
	collides = false
	collider = nil
	for _, tile := range b.Tiles {
		if tile.posCollides(pos) {
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
	tile.Remove()
	b.removeCandidates(tile)
	b.addCandidate(Coord{tile.X, tile.Y})
	b.Tiles = b.Tiles[:len(b.Tiles)-1]

	return &tile
}

func (b *Board) removeCandidates(tile Tile) {

	toRemove := tile.GetNeighborSpots()

	//remove top if it exists
	for i, cand := range b.Candidates {
		if cand.X == toRemove[1].X && cand.Y == toRemove[1].Y {
			b.Candidates = append(b.Candidates[:i], b.Candidates[Min(len(b.Candidates), i+1):]...)
			break //assume only one occurance
		}
	}

	//remove right spot if it exists
	for i, cand := range b.Candidates {
		if cand.X == toRemove[0].X { //only check same column, row is different
			if cand.Y >= tile.Y && cand.Y < tile.Y+tile.CurH {
				b.Candidates = append(b.Candidates[:i], b.Candidates[Min(len(b.Candidates), i+1):]...)
				break //assume only one occurance
			}
		}
	}
}
