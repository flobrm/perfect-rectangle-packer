package tiling

import "localhost/flobrm/tilingsolver/core"

//Board stores the board and everything placed on it
type Board struct {
	Size          core.Coord   //width and hight of the board
	Tiles         [](*Tile)    //All the placed tiles
	Candidates    []core.Coord //Candidate positions for next placement
	board         [][]uint8    // first x then y
	pairs         [](*TilePair)
	lastCollision *Tile
}

//NewBoard inits a board, including candidates
func NewBoard(boardDims core.Coord, tiles []Tile) Board {
	myTiles := make([](*Tile), len(tiles))
	candidates := append(make([]core.Coord, 0), core.Coord{X: 0, Y: 0})
	board := make([][]uint8, boardDims.X)
	for i := 0; i < len(board); i++ {
		board[i] = make([]uint8, boardDims.Y)
	}
	return Board{
		Size:       core.Coord{X: boardDims.X, Y: boardDims.Y},
		Tiles:      myTiles[:0],
		Candidates: candidates,
		board:      board,
		pairs:      make([](*TilePair), len(tiles))[:0]}
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
//TODO merge with Place
func (b *Board) Fits(tile *Tile, turned bool) bool {
	// fmt.Println(tile)
	candIndex := len(b.Candidates) - 1
	pos := b.Candidates[candIndex]
	tile.Place(pos, turned)
	if !b.tileFitsBoard(tile) {
		tile.Remove()
		return false
	}
	//Check if the tile is a corner piece smaller than the lower left corner tile
	if len(b.Tiles) > 0 && tile.Index < b.Tiles[0].Index {
		corner := b.isCornerTile(tile)
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

	pair, isCanonical := b.getPair(tile)
	if pair == nil {
		return true
	} else if isCanonical {
		b.pairs = append(b.pairs, pair) //TODO perhaps in placeTile?
		return true
	} else {
		return false
	}

	// return true
}

//tileFitsBoard checks if the tile with it's internal rotation and position fits inside the board
func (b *Board) tileFitsBoard(tile *Tile) bool {
	if tile.X+tile.CurW > b.Size.X || tile.Y+tile.CurH > b.Size.Y || //check top and right side
		tile.X < 0 || tile.Y < 0 { //check bottom and left side
		return false
	}
	return true
}

// getPair checks if tile forms a larger rectangle with the tiles already placed on the board.
// It only returns the first found pair where the counter tile is not already part of a pair.
// isCanonical is true if the tile with the lower index is lower or more to the left
func (b *Board) getPair(tile *Tile) (*TilePair, bool) {
	//check only last tile, assume last tile should be below or to the left
	if len(b.Tiles) > 1 {
		lastTile := b.Tiles[len(b.Tiles)-1]
		if tile.Y == lastTile.Y && tile.CurH == lastTile.CurH && //same row and same height
			lastTile.X+lastTile.CurW == tile.X { //lastTile is to the left of tile
			return &TilePair{a: lastTile, b: tile}, lastTile.Index > tile.Index
		}
	}
	return nil, false
}

//PlaceTile registers a tile as placed on the current spot to fill, Fits should be called first
func (b *Board) PlaceTile(tile *Tile, turned bool) {
	candIndex := len(b.Candidates) - 1
	pos := b.Candidates[candIndex]
	tile.Place(pos, turned)
	b.putTileOnBoard(tile)
	b.Candidates = b.Candidates[:candIndex] //remove last candidate
	b.addCandidates(*tile)
	b.Tiles = append(b.Tiles, tile)
}

func (b *Board) putTileOnBoard(tile *Tile) {
	index := uint8(len(b.Tiles) + 1)
	tileTop := tile.Y + tile.CurH - 1
	for x := tile.X; x < tile.X+tile.CurW; x++ {
		b.board[x][tile.Y] = index
		b.board[x][tileTop] = index
	}
	tileRightEdge := tile.X + tile.CurW - 1
	for y := tile.Y + 1; y < tile.Y+tile.CurH-1; y++ {
		b.board[tile.X][y] = index
		b.board[tileRightEdge][y] = index
	}
}

func (b *Board) removeTileFromBoard(tile *Tile) {
	index := uint8(0)
	tileTop := tile.Y + tile.CurH - 1
	for x := tile.X; x < tile.X+tile.CurW; x++ {
		b.board[x][tile.Y] = index
		b.board[x][tileTop] = index
	}
	tileRightEdge := tile.X + tile.CurW - 1
	for y := tile.Y + 1; y < tile.Y+tile.CurH-1; y++ {
		b.board[tile.X][y] = index
		b.board[tileRightEdge][y] = index
	}
}

func (b *Board) addCandidates(tile Tile) {

	//find top tile
	candidateY := tile.Y + tile.CurH
	if candidateY < b.Size.Y {
		if tile.X == 0 { //left border counts as corner
			b.addCandidate(core.Coord{X: 0, Y: candidateY})
		} else if b.board[tile.X-1][candidateY] != 0 {
			for x := tile.X; x < tile.X+tile.CurW; x++ {
				if b.board[x][candidateY] == 0 {
					b.addCandidate(core.Coord{X: x, Y: candidateY})
					break
				}
			}
		}
	}

	//find right tile
	candidateX := tile.X + tile.CurW
	if candidateX < b.Size.X {
		if tile.Y == 0 { //always add candidate if tile on bottom
			b.addCandidate(core.Coord{X: candidateX, Y: 0})
		} else if b.board[candidateX][tile.Y-1] != 0 {
			for y := tile.Y; y < tile.Y+tile.CurH; y++ {
				if b.board[candidateX][y] == 0 {
					b.addCandidate(core.Coord{X: candidateX, Y: y})
					break
				}
			}
		}
	}
}

func (b *Board) posCollides(pos core.Coord) (collides bool, collider *Tile) {
	collides = b.board[pos.X][pos.Y] != 0
	if collides {
		collider = b.Tiles[b.board[pos.X][pos.Y]-1]
		return collides, collider
	}
	return collides, collider
}

//RemoveLastTile removes a tile and resets it's candidate positions
func (b *Board) RemoveLastTile() {
	tile := *b.Tiles[len(b.Tiles)-1]
	b.removeLastPair(&tile)
	b.removeCandidates(tile)
	b.removeTileFromBoard(&tile)
	b.addCandidate(core.Coord{X: tile.X, Y: tile.Y})
	tile.Remove()
	b.Tiles = b.Tiles[:len(b.Tiles)-1]
}

func (b *Board) removeLastPair(tile *Tile) {
	index := len(b.pairs) - 1
	if index >= 0 {
		if b.pairs[index].a.Index == tile.Index || b.pairs[index].b.Index == tile.Index {
			b.pairs = b.pairs[:index]
			// fmt.Println("removing")
		}
	}
	// fmt.Println(len(b.pairs))
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
func (b *Board) GetCanonicalSolution(tiles *[]Tile) int { //TODO fix to use grid
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
