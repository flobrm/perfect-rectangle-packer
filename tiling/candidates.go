package tiling

import "localhost/flobrm/tilingsolver/core"

type candidateList struct {
	candidates    []gap
	nextCandidate int
}

//newCandidateList is an easy way to get a candidatelist
func newCandidateList(maxCandidates int) candidateList {
	candidates := make([]gap, maxCandidates)[:0]

	return candidateList{
		candidates:    candidates,
		nextCandidate: 0,
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
	cl.nextCandidate = len(cl.candidates) - 1

	//other possible next candidate
	smallestGapWidth := 999999999999
	for i, gap := range cl.candidates {
		if !gap.active {
			continue
		}
		if gap.W < smallestGapWidth {
			cl.nextCandidate = i
			smallestGapWidth = gap.W
		}
		if gap.W == smallestGapWidth {
			//TODO what should happen in case of a tiebreak, highest gap first?
			continue
		}
	}
}

//TODO add argument to specify what rules should be used to search (BL, stack, smallest)
func (cl *candidateList) nextGap() *gap {
	return &cl.candidates[len(cl.candidates)-1]
}

func (cl *candidateList) removeNextGap() {
	cl.candidates = cl.candidates[:len(cl.candidates)-1]
}

//removeLatestCandidates removes the 1 or 2 candidates that were added last if they share a border with tile.
func (cl *candidateList) removeLatestCandidates(tile *Tile) {
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
