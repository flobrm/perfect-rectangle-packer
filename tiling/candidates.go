package tiling

import "localhost/flobrm/tilingsolver/core"

type candidateList struct {
	candidates     []gap
	nextCandidate  int
	candidateOrder int
}

//This const identifies the different ways to place the next tile.
//LastGapFirst always picks the latest added gap that is active
//SmallestGapFirst picks the smallest active gap, using largest height to win a tie
//BottomLeft picks the gap lowest in the frame, picking the leftmost gap in case of a tie.
const (
	LastGapFirst     = iota
	SmallestGapFirst = iota
	BottomLeft       = iota
)

//PlacementOrder is a const map mapping command strings to a placement order algorithm.
//Currently supported options are lastGapAdded (default), smallestGap, bottomLeft
var PlacementOrder = map[string]int{
	"lastGapAdded": LastGapFirst,
	"smalestGap":   SmallestGapFirst,
	"bottomLeft":   BottomLeft,
}

//newCandidateList is an easy way to get a candidatelist
func newCandidateList(maxCandidates int, candidateOrder int) candidateList {
	candidates := make([]gap, maxCandidates)[:0]

	return candidateList{
		candidates:     candidates,
		nextCandidate:  0,
		candidateOrder: candidateOrder,
	}
}

func (cl *candidateList) isEmpty() bool {
	return len(cl.candidates) == 0
}

func (cl *candidateList) addCandidate(candidate gap) {
	cl.candidates = append(cl.candidates, candidate)
}

//recalcNextCandidate picks a new next candidate according to a specific ruleset
//The only available ruleset right now is last candidate added.
func (cl *candidateList) recalcNextCandidate() {

	if cl.candidateOrder == SmallestGapFirst {
		smallestGapWidth := 999999999999
		smallestGapHeight := 999999999999
		for i, gap := range cl.candidates {
			if !gap.active {
				continue
			}
			if gap.W < smallestGapWidth || (gap.W == smallestGapWidth && gap.H > smallestGapHeight) {
				cl.nextCandidate = i
				smallestGapWidth = gap.W
				smallestGapHeight = gap.H
			}
		}
	} else if cl.candidateOrder == BottomLeft {
		lowestGapY := 999999999999
		lowestGapX := 999999999999
		for i, gap := range cl.candidates {
			if !gap.active {
				continue
			}
			if gap.Pos.Y < lowestGapY || gap.Pos.Y == lowestGapY && gap.Pos.X < lowestGapX {
				cl.nextCandidate = i
				lowestGapX = gap.Pos.X
				lowestGapY = gap.Pos.Y
			}
		}
	} else {
		cl.nextCandidate = len(cl.candidates) - 1
	}
}

//TODO add argument to specify what rules should be used to search (BL, stack, smallest)
func (cl *candidateList) nextGap() *gap {
	return &cl.candidates[cl.nextCandidate]
}

func (cl *candidateList) removeNextGap() {
	// cl.candidates = cl.candidates[:len(cl.candidates)-1]
	cl.candidates = append(cl.candidates[:cl.nextCandidate], cl.candidates[cl.nextCandidate+1:]...)
}

//removeLatestCandidates removes the 1 or 2 candidates that were added last if they share a border with tile.
func (cl *candidateList) removeLatestCandidates(tile *Tile) {
	if len(cl.candidates) == 0 {
		return
	}

	removedCandidates := 0
	for i := len(cl.candidates) - 1; i >= 0; i-- {
		cand := cl.candidates[i]
		if isRightCandidate(cand.Pos, tile) || isTopCandidate(cand.Pos, tile) {
			cl.candidates = append(cl.candidates[:i], cl.candidates[i+1:]...)
			removedCandidates++
			if removedCandidates == 2 {
				return
			}
		}
	}
}

func isRightCandidate(cand core.Coord, tile *Tile) bool {
	if cand.X == tile.X+tile.CurW {
		if cand.Y >= tile.Y && cand.Y < tile.Y+tile.CurH {
			return true
		}
	}
	return false
}

func isTopCandidate(cand core.Coord, tile *Tile) bool {
	if cand.Y == tile.Y+tile.CurH {
		if cand.X >= tile.X && cand.X < tile.X+tile.CurW {
			return true
		}
	}
	return false
}
