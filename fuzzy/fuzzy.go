// Package fuzzy ranks strings against a query with a subsequence matcher
// that rewards consecutive characters, word boundaries, prefixes and exact
// matches, as in fzf-like pickers.
package fuzzy

import (
	"sort"
	"strings"
	"unicode"
)

const (
	baseScore        = 10
	consecutiveScale = 12
	boundaryBonus    = 40
	gapPenalty       = 3
	exactBonus       = 1000
	prefixBonus      = 300
	substringBonus   = 180
)

// Match is the score of a value and the rune positions that matched.
type Match struct {
	Score     int
	Positions []int
}

// Field is one searchable value and its weight in ScoreFields.
type Field struct {
	Value  string
	Weight int
}

type matchState struct {
	score     int
	run       int
	positions []int
	valid     bool
}

type scoreState struct {
	score int
	run   int
	valid bool
}

// MatchValue returns the best subsequence alignment for every query token.
// Tokens must all match; their scores and character positions are combined.
func MatchValue(value, query string) (Match, bool) {
	query = strings.TrimSpace(query)
	if query == "" {
		return Match{}, true
	}
	combined := Match{}
	seenPositions := map[int]bool{}
	for _, token := range strings.Fields(query) {
		match, ok := matchToken(value, token)
		if !ok {
			return Match{}, false
		}
		combined.Score += match.Score
		for _, position := range match.Positions {
			if !seenPositions[position] {
				seenPositions[position] = true
				combined.Positions = append(combined.Positions, position)
			}
		}
	}
	sort.Ints(combined.Positions)
	return combined, true
}

// ScoreFields lets separate display/search fields carry different weights.
// Every query token may choose its best field, so a title and an alias can both
// contribute without creating synthetic matches across field boundaries.
func ScoreFields(query string, fields ...Field) (int, bool) {
	tokens := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	if len(tokens) == 0 {
		return 0, true
	}
	bestScores := make([]int, len(tokens))
	matchedTokens := make([]bool, len(tokens))
	var previous, current []scoreState
	for _, field := range fields {
		value := strings.TrimSpace(field.Value)
		if value == "" || field.Weight <= 0 {
			continue
		}
		lowerValue := strings.ToLower(value)
		haystack := []rune(lowerValue)
		if cap(previous) < len(haystack) {
			previous = make([]scoreState, len(haystack))
			current = make([]scoreState, len(haystack))
		} else {
			previous = previous[:len(haystack)]
			current = current[:len(haystack)]
		}
		for index, token := range tokens {
			score, ok := scoreTokenRunes(haystack, lowerValue, token, previous, current)
			if !ok {
				continue
			}
			score *= field.Weight
			if !matchedTokens[index] || score > bestScores[index] {
				bestScores[index], matchedTokens[index] = score, true
			}
		}
	}
	total := 0
	for index, matched := range matchedTokens {
		if !matched {
			return 0, false
		}
		total += bestScores[index]
	}
	return total, true
}

func matchToken(value, token string) (Match, bool) {
	return matchTokenInternal(value, token, true)
}

func matchTokenInternal(value, token string, collectPositions bool) (Match, bool) {
	haystack := []rune(strings.ToLower(value))
	needle := []rune(strings.ToLower(strings.TrimSpace(token)))
	if len(needle) == 0 {
		return Match{}, true
	}
	if len(haystack) == 0 || len(needle) > len(haystack) {
		return Match{}, false
	}

	previous := make([]matchState, len(haystack))
	for queryIndex, wanted := range needle {
		current := make([]matchState, len(haystack))
		var bestGap matchState
		bestGapAdjusted := 0
		bestGapValid := false
		for valueIndex, candidate := range haystack {
			// A gapped transition into valueIndex can only come from positions
			// strictly before valueIndex-1. Keep the best adjusted prefix so the
			// dynamic program remains O(query × value).
			if queryIndex > 0 && valueIndex >= 2 {
				state := previous[valueIndex-2]
				adjusted := state.score + gapPenalty*(valueIndex-2)
				if state.valid && (!bestGapValid || adjusted > bestGapAdjusted) {
					bestGap, bestGapAdjusted, bestGapValid = state, adjusted, true
				}
			}
			if candidate != wanted {
				continue
			}
			boundary := 0
			if isBoundary(haystack, valueIndex) {
				boundary = boundaryBonus
			}
			store := func(state matchState) {
				if !current[valueIndex].valid || state.score > current[valueIndex].score {
					current[valueIndex] = state
				}
			}

			if queryIndex == 0 {
				state := matchState{valid: true, run: 1, score: baseScore + consecutiveScale + boundary - valueIndex}
				if collectPositions {
					state.positions = []int{valueIndex}
				}
				store(state)
				continue
			}
			if valueIndex > 0 && previous[valueIndex-1].valid {
				state := previous[valueIndex-1]
				run := state.run + 1
				increment := baseScore + boundary + consecutiveScale*(run*run-state.run*state.run)
				next := matchState{valid: true, run: run, score: state.score + increment}
				if collectPositions {
					next.positions = appendPosition(state.positions, valueIndex)
				}
				store(next)
			}
			if bestGapValid {
				// bestGap is selected using score+penalty*position, so recover
				// its last position only when positions are requested. In the hot
				// score-only path the adjusted score gives the result directly.
				gapScore := bestGapAdjusted + baseScore + consecutiveScale + boundary - gapPenalty*(valueIndex-1)
				next := matchState{valid: true, run: 1, score: gapScore}
				if collectPositions {
					next.positions = appendPosition(bestGap.positions, valueIndex)
				}
				store(next)
			}
		}
		previous = current
	}

	bestState := matchState{}
	found := false
	for _, state := range previous {
		if state.valid && (!found || state.score > bestState.score) {
			bestState, found = state, true
		}
	}
	if !found {
		return Match{}, false
	}

	best := Match{Score: bestState.score, Positions: bestState.positions}
	lowerValue := strings.ToLower(strings.TrimSpace(value))
	lowerToken := strings.ToLower(strings.TrimSpace(token))
	switch {
	case lowerValue == lowerToken:
		best.Score += exactBonus
	case strings.HasPrefix(lowerValue, lowerToken):
		best.Score += prefixBonus
	case strings.Contains(lowerValue, lowerToken):
		best.Score += substringBonus
	}
	return best, best.Score > 0
}

func scoreTokenRunes(haystack []rune, lowerValue, lowerToken string, previous, current []scoreState) (int, bool) {
	needle := []rune(lowerToken)
	if len(needle) == 0 {
		return 0, true
	}
	if len(haystack) == 0 || len(needle) > len(haystack) {
		return 0, false
	}
	clear(previous)
	clear(current)
	for queryIndex, wanted := range needle {
		clear(current)
		bestGapAdjusted := 0
		bestGapValid := false
		for valueIndex, candidate := range haystack {
			if queryIndex > 0 && valueIndex >= 2 {
				state := previous[valueIndex-2]
				adjusted := state.score + gapPenalty*(valueIndex-2)
				if state.valid && (!bestGapValid || adjusted > bestGapAdjusted) {
					bestGapAdjusted, bestGapValid = adjusted, true
				}
			}
			if candidate != wanted {
				continue
			}
			boundary := 0
			if isBoundary(haystack, valueIndex) {
				boundary = boundaryBonus
			}
			store := func(state scoreState) {
				if !current[valueIndex].valid || state.score > current[valueIndex].score {
					current[valueIndex] = state
				}
			}
			if queryIndex == 0 {
				store(scoreState{valid: true, run: 1, score: baseScore + consecutiveScale + boundary - valueIndex})
				continue
			}
			if valueIndex > 0 && previous[valueIndex-1].valid {
				state := previous[valueIndex-1]
				run := state.run + 1
				increment := baseScore + boundary + consecutiveScale*(run*run-state.run*state.run)
				store(scoreState{valid: true, run: run, score: state.score + increment})
			}
			if bestGapValid {
				gapScore := bestGapAdjusted + baseScore + consecutiveScale + boundary - gapPenalty*(valueIndex-1)
				store(scoreState{valid: true, run: 1, score: gapScore})
			}
		}
		previous, current = current, previous
	}

	best, found := 0, false
	for _, state := range previous {
		if state.valid && (!found || state.score > best) {
			best, found = state.score, true
		}
	}
	if !found {
		return 0, false
	}
	switch {
	case lowerValue == lowerToken:
		best += exactBonus
	case strings.HasPrefix(lowerValue, lowerToken):
		best += prefixBonus
	case strings.Contains(lowerValue, lowerToken):
		best += substringBonus
	}
	return best, best > 0
}

func appendPosition(positions []int, position int) []int {
	out := make([]int, len(positions)+1)
	copy(out, positions)
	out[len(positions)] = position
	return out
}

func isBoundary(value []rune, index int) bool {
	if index == 0 {
		return true
	}
	previous := value[index-1]
	return !unicode.IsLetter(previous) && !unicode.IsNumber(previous)
}

// Contains reports whether every token of needle matches haystack.
func Contains(haystack, needle string) bool {
	_, ok := ScoreFields(needle, Field{Value: haystack, Weight: 1})
	return ok
}
