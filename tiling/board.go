package tiling

import (
	"localhost/flobrm/tilingsolver/core"
)

//Board stores the board and everything placed on it
type Board struct {
	Size       core.Coord //width and hight of the board
	Tiles      [](*Tile)  //All the placed tiles
	Candidates []gap
	// Candidates    []core.Coord //Candidate positions for next placement
	board         [][]uint8 // first x then y
	gapTable      [][][]int //lookup table for impossible gaps, in order width, height, tileIndex
	maxGapTable   [][]int   //lookup table with maximum possible area for a gap of a certain width and height, given a full tileset
	lastCollision *Tile
}

//NewBoard inits a board, including candidates
func NewBoard(boardDims core.Coord, tiles []Tile) Board {
	myTiles := make([](*Tile), len(tiles))
	// candidates := append(make([]core.Coord, 0), core.Coord{X: 0, Y: 0})
	firstGap := gap{Pos: core.Coord{}, W: boardDims.X, H: boardDims.Y, leftH: boardDims.Y, active: true, leftSideActive: true}
	candidates := append(make([]gap, len(tiles))[:0], firstGap)
	board := make([][]uint8, boardDims.X)
	gapTable, maxGapTable := buildGapTable(tiles, 15, 20) //TODO make this a variable

	for i := 0; i < len(board); i++ {
		board[i] = make([]uint8, boardDims.Y)
	}
	return Board{
		Size:        core.Coord{X: boardDims.X, Y: boardDims.Y},
		Tiles:       myTiles[:0],
		Candidates:  candidates,
		board:       board,
		gapTable:    gapTable,
		maxGapTable: maxGapTable,
	}
}

func buildGapTable(tiles []Tile, maxGapWidth int, maxGapHeight int) ([][][]int, [][]int) {
	gapTable := make([][][]int, maxGapWidth+1)
	// gapTable := make([][]int, len(tiles))
	maxGapArea := make([][]int, maxGapWidth+1)
	for width := 0; width <= maxGapWidth; width++ {
		// fmt.Print(width)
		gapTable[width] = make([][]int, maxGapHeight+1)
		maxGapArea[width] = make([]int, maxGapHeight+1)
		for height := 0; height <= maxGapHeight; height++ {
			// fmt.Print(height)
			gapTable[width][height] = make([]int, len(tiles))
			for _, tile := range tiles {
				if tile.H <= width {
					area := Min(tile.W, height) * tile.H
					if tile.W <= width {
						area = Max(area, tile.W*Min(tile.H, height))
					}
					gapTable[width][height][tile.Index] = area
					maxGapArea[width][height] += area
					// fmt.Println("w:", width, "h:", height, "tile", tile.W, tile.H, "area", area)
				}
			}
		}
	}

	return gapTable, maxGapArea
}

func (b *Board) addCandidate(newGap gap) {
	b.Candidates = append(b.Candidates, newGap)
}

//Min returns the min of two integers, why the fuck do I have to define this
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//Max returns the min of two integers, why the fuck do I have to define this
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//Place places a tile on the board if it is possible. It returns whether the tile was placed
func (b *Board) Place(tile *Tile, turned bool) bool {
	if b.fits(tile, turned) {
		b.placeTile(tile, turned)
		return true
	}
	return false
}

//HasUnfillableGaps check in different ways if there are unfillable gaps on the board
func (b *Board) HasUnfillableGaps(onlyNextCandidate bool, checkGapsFromLeft bool) bool {
	if onlyNextCandidate {
		nextGap := &b.Candidates[len(b.Candidates)-1]
		if b.gapIsUnfillable(nextGap) {
			return true
		}
	} else {
		if b.anyGapsUnfillable() {
			return true
		}
	}
	if checkGapsFromLeft {
		if b.hasUnfillableLeftSideGaps() {
			return true
		}
	}

	return false
}

//Fits checks if a tile fits the board at the next position to fill
//TODO merge with Place
func (b *Board) fits(tile *Tile, turned bool) bool {
	gap := b.Candidates[len(b.Candidates)-1]

	tile.Place(gap.Pos, turned)
	if !b.tileFitsBoard(tile) {
		tile.Remove()
		return false
	}

	if !gap.couldFit(tile) {
		tile.Remove()
		return false
	}

	//Check if the tile is a corner piece smaller than the lower left corner tile
	if len(b.Tiles) > 0 && tile.Index < b.Tiles[0].Index {
		corner := b.isCornerTile(tile)
		if corner != noCorner && corner != bottomLeftCorner {
			tile.Remove()
			return false
		}
	}
	//TODO check if part is sticking out above gap and check that part for collisions

	//TODO check if gap is plausible

	// notIllegalPair := b.updateNeighborsTree(tile)
	// if !notIllegalPair {
	// 	b.removeTileFromPairTree(tile)
	// 	tile.Remove()
	// 	return false
	// }

	// return true
	return true
}

type gap struct {
	Pos                    core.Coord
	W, H, leftH            int
	active, leftSideActive bool
}

//couldFit returns if a tile could fit in a gap. This assumes the tile is placed in the lower left corner
// It doesn't check if the part of the tile above the gap collides with anything.
func (g *gap) couldFit(tile *Tile) bool {
	if tile.CurW <= g.W {
		return true
	}
	return false
}

func (b *Board) anyGapsUnfillable() bool {
	//TODO check if all gaps combined could be fillable
	for i := range b.Candidates {
		if b.gapIsUnfillable(&b.Candidates[i]) {
			return true
		}
	}
	return false
}

func (b *Board) hasUnfillableLeftSideGaps() bool {
	for i := range b.Candidates {
		if b.Candidates[i].leftSideActive {
			if b.leftSideGapIsUnfillable(&b.Candidates[i]) {
				return true
			}
		}
	}
	return false
}

func (b *Board) leftGapHeight(pos *core.Coord) int {
	if pos.X == 0 {
		return b.Size.Y - pos.Y
	}
	leftY := pos.Y
	for leftY < b.Size.Y && b.board[pos.X-1][leftY] != 0 && b.board[pos.X][leftY] == 0 {
		index := b.board[pos.X-1][leftY] - 1
		leftY = b.Tiles[index].Y + b.Tiles[index].CurH
	}
	return leftY - pos.Y
}

func (b *Board) rightGapHeight(pos *core.Coord, width int) int {
	rightX := pos.X + width
	if rightX >= b.Size.X {
		return b.Size.Y - pos.Y
	}
	curY := pos.Y
	for curY < b.Size.Y && b.board[rightX][curY] != 0 && b.board[rightX-1][curY] == 0 {
		index := b.board[rightX][curY] - 1
		curY = b.Tiles[index].Y + b.Tiles[index].CurH
	}

	return curY - pos.Y
}

func (b *Board) gapWidth(pos *core.Coord) int {
	if pos.Y == 0 {
		return b.Size.X - pos.X //TODO actually check board, instead of assuming
	}
	xPos := pos.X
	for xPos < b.Size.X && b.board[xPos][pos.Y-1] != 0 && b.board[xPos][pos.Y] == 0 {
		index := b.board[xPos][pos.Y-1] - 1
		xPos = b.Tiles[index].X + b.Tiles[index].CurW
	}
	return xPos - pos.X
}

func (b *Board) makeNewGap(pos *core.Coord) gap {

	width := b.gapWidth(pos)
	leftHeight := b.leftGapHeight(pos)
	rightHeight := b.rightGapHeight(pos, width)
	//TODO build gap
	return gap{
		Pos:            *pos,
		W:              width,                        //TODO
		H:              Min(leftHeight, rightHeight), //TODO
		leftH:          leftHeight,
		active:         rightHeight > 0,
		leftSideActive: pos.Y+leftHeight >= b.Size.Y && width > 0,
	}
}

func (b *Board) updateGap(g *gap) {
	g.W = b.gapWidth(&g.Pos)
	g.leftH = b.leftGapHeight(&g.Pos)
	g.H = Min(g.leftH, b.rightGapHeight(&g.Pos, g.W))
	g.active = g.H > 0
	g.leftSideActive = g.Pos.Y+g.leftH >= b.Size.Y && g.W > 0
}

func (b *Board) updateExistingCandidates(tile *Tile) {
	for i, gap := range b.Candidates {
		//Only update if the tile is inside, or directly adjacent to a gap
		if tile.Y+tile.CurH <= gap.Pos.Y || tile.Y > gap.Pos.Y+gap.H {
			continue
		}
		if tile.X > gap.Pos.X+gap.W || tile.X+tile.CurW <= gap.Pos.X {
			continue
		}
		b.updateGap(&b.Candidates[i])
	}
}

// gapIsUnfillable returns true if there is no way to sum to the area of a gap with the unplaced tiles.
// It is named like this because proving that a gap is fillable with a given tileset is a lot
// harder and outside the scope of this function.
func (b *Board) gapIsUnfillable(g *gap) bool {
	if !g.active {
		return false
	}
	width := g.W
	height := g.H
	//first check if our lookup tables contain a gap of those dimensions
	if width+1 > len(b.gapTable) || height+1 > len(b.gapTable[0]) {
		return false
	}
	targetArea := width * height
	maxArea := b.maxGapTable[width][height]
	if maxArea < targetArea {
		return true
	}
	for _, tile := range b.Tiles { // remove the areas of already placed tiles
		maxArea -= b.gapTable[width][height][tile.Index]
	}
	if maxArea < targetArea {
		return true
	}
	return false
}

func (b *Board) leftSideGapIsUnfillable(g *gap) bool {
	if !g.leftSideActive {
		return false
	}
	width := g.leftH
	height := g.W
	//first check if our lookup tables contain a gap of those dimensions
	if width+1 > len(b.gapTable) || height+1 > len(b.gapTable[0]) {
		return false
	}
	targetArea := width * height
	maxArea := b.maxGapTable[width][height]
	if maxArea < targetArea {
		return true
	}
	for _, tile := range b.Tiles { // remove the areas of already placed tiles
		maxArea -= b.gapTable[width][height][tile.Index]
	}
	if maxArea < targetArea {
		return true
	}
	return false
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
func (b *Board) placeTile(tile *Tile, turned bool) {
	candIndex := len(b.Candidates) - 1
	tile.Place(b.Candidates[candIndex].Pos, turned)
	b.putTileOnBoard(tile)
	b.Tiles = append(b.Tiles, tile)
	b.Candidates = b.Candidates[:candIndex] //remove last candidate
	b.updateExistingCandidates(tile)
	b.addCandidates(*tile)
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
			b.addCandidate(b.makeNewGap(&core.Coord{X: 0, Y: candidateY}))
		} else if b.board[tile.X-1][candidateY] != 0 {
			for x := tile.X; x < tile.X+tile.CurW; x++ {
				if b.board[x][candidateY] == 0 {
					b.addCandidate(b.makeNewGap(&core.Coord{X: x, Y: candidateY}))
					break
				}
			}
		}
	}

	//find right tile
	candidateX := tile.X + tile.CurW
	if candidateX < b.Size.X {
		if tile.Y == 0 { //always add candidate if tile on bottom
			b.addCandidate(b.makeNewGap(&core.Coord{X: candidateX, Y: 0}))
		} else if b.board[candidateX][tile.Y-1] != 0 {
			for y := tile.Y; y < tile.Y+tile.CurH; y++ {
				if b.board[candidateX][y] == 0 {
					b.addCandidate(b.makeNewGap(&core.Coord{X: candidateX, Y: y}))
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
	b.removeTileFromPairTree(&tile)
	b.removeCandidates(tile)
	b.removeTileFromBoard(&tile)
	b.addCandidate(b.makeNewGap(&core.Coord{X: tile.X, Y: tile.Y}))
	if b.lastCollision != nil && b.lastCollision.Index == tile.Index {
		b.lastCollision = nil
	}
	tile.Remove()
	b.Tiles = b.Tiles[:len(b.Tiles)-1]
}

func (b *Board) removeCandidates(tile Tile) {
	if len(b.Candidates) == 0 {
		return
	}
	cand := b.Candidates[len(b.Candidates)-1].Pos
	if isRightCandidate(cand, tile) || isTopCandidate(cand, tile) {
		b.Candidates = b.Candidates[:len(b.Candidates)-1]
		if len(b.Candidates) == 0 {
			return
		}
		cand = b.Candidates[len(b.Candidates)-1].Pos
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

func (b *Board) addTilePair(t1 *Tile, t2 *Tile) {
	var X, Y, W, H int
	X = Min(t1.X, t2.X)
	Y = Min(t1.Y, t2.Y)
	if t1.CurW == t2.CurW {
		W = t1.CurW
		H = t1.CurH + t2.CurH
	} else {
		W = t1.CurW + t2.CurW
		H = t1.CurH
	}
	parent := NewTile(W, H)
	parent.Place(core.Coord{X: X, Y: Y}, false)
	parent.Index = Max(t1.Index, t2.Index) //TODO should this be min or max?
	parent.lChild = t1
	parent.rChild = t2
	t1.parent = &parent
	t2.parent = &parent
}

//Check if the tile has neighbors on the board. If all neighbors have a larger index return true and
func (b *Board) updateNeighborsTree(tile *Tile) bool {
	for tileAddition := tile; tileAddition != nil; tileAddition = tileAddition.parent {
		//check bottom
		if tileAddition.Y > 0 && b.board[tileAddition.X][tileAddition.Y-1] > 0 {
			otherIndex := b.board[tileAddition.X][tileAddition.Y-1] - 1
			for other := b.Tiles[otherIndex]; other != nil; other = other.parent {
				if other.X == tileAddition.X && other.CurW == tileAddition.CurW {
					if other.Index > tileAddition.Index {
						return false // found an illegal pair
					}
					//There is a legal pairing, see what to do about it
					if tileAddition.parent != nil || other.parent != nil {
						continue //already have a parent, can't add a new node
					} else {
						b.addTilePair(other, tileAddition)
						break
					}
				} else { //No pair, continue checking the other directions
					break
				}
			}
		}

		//check left
		if tileAddition.X > 0 && b.board[tileAddition.X-1][tileAddition.Y] > 0 {
			otherIndex := b.board[tileAddition.X-1][tileAddition.Y] - 1
			for other := b.Tiles[otherIndex]; other != nil; other = other.parent {
				if other.Y == tileAddition.Y && other.CurH == tileAddition.CurH {
					if other.Index > tileAddition.Index {
						return false // found an illegal pair
					}
					//There is a legal pairing, see what to do about it
					if tileAddition.parent != nil || other.parent != nil {
						continue //already have a parent, can't add a new node
					} else {
						b.addTilePair(other, tileAddition)
						break
					}
				} else { //No pair, continue checking the other directions
					break
				}
			}
		}

		//check right
		if tileAddition.X+tileAddition.CurW < b.Size.X-1 && b.board[tileAddition.X+tileAddition.CurW][tileAddition.Y] > 0 {
			otherIndex := b.board[tileAddition.X+tileAddition.CurW][tileAddition.Y] - 1
			for other := b.Tiles[otherIndex]; other != nil; other = other.parent {
				if other.Y == tileAddition.Y && other.CurH == tileAddition.CurH {
					if other.Index > tileAddition.Index {
						return false // found an illegal pair
					}
					//There is a legal pairing, see what to do about it
					if tileAddition.parent != nil || other.parent != nil {
						continue //already have a parent, can't add a new node
					} else {
						b.addTilePair(other, tileAddition)
						break
					}
				} else { //No pair, continue checking the other directions
					break
				}
			}
		}

		//check top
		if tileAddition.Y+tileAddition.CurH < b.Size.Y-1 && b.board[tileAddition.X][tileAddition.Y+tileAddition.CurH] > 0 {
			otherIndex := b.board[tileAddition.X][tileAddition.Y-1] - 1
			for other := b.Tiles[otherIndex]; other != nil; other = other.parent {
				if other.X == tileAddition.X && other.CurW == tileAddition.CurW {
					if other.Index > tileAddition.Index {
						return false // found an illegal pair
					}
					//There is a legal pairing, see what to do about it
					if tileAddition.parent != nil || other.parent != nil {
						continue //already have a parent, can't add a new node
					} else {
						b.addTilePair(other, tileAddition)
						break
					}
				} else { //No pair, continue checking the other directions
					break
				}
			}
		}
	}
	return true
}

func (b *Board) removeTileFromPairTree(tile *Tile) {
	for parent := tile.parent; parent != nil; parent = parent.parent {
		parent.lChild.parent = nil
		parent.rChild.parent = nil
		parent.lChild = nil
		parent.rChild = nil
	}
}
