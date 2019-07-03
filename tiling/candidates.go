package tiling

import "localhost/flobrm/tilingsolver/core"

type candidateList struct {
	candidates     []gap
	nextCandidate  int
	candidateOrder int
}

const (
	lastGapFirst     = iota
	smallestGapFirst = iota
)

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

	if cl.candidateOrder == smallestGapFirst {
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
	//TODO fix this to check all candidates instead of only last two, alternatively make a better way to link tiles to candidates
	if len(cl.candidates) == 0 {
		return
	}
	cand := cl.candidates[len(cl.candidates)-1].Pos
	if isRightCandidate(cand, tile) || isTopCandidate(cand, tile) {
		cl.candidates = cl.candidates[:len(cl.candidates)-1]
		if len(cl.candidates) == 0 {
			return
		}
		cand = cl.candidates[len(cl.candidates)-1].Pos
		if isRightCandidate(cand, tile) || isTopCandidate(cand, tile) {
			cl.candidates = cl.candidates[:len(cl.candidates)-1]
		}
	}

	// removedCandidates := 0
	// for i := len(cl), cand := range cl.candidates { //TODO go through this in reverse order, to avoid skipping after delete
	// 	//if isRightCandidate(cand, )
	// }
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
