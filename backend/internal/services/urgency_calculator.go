package services

import (
	"log"
	"strings"

	"github.com/google/uuid"
)

// UrgencyLevel represents the urgency classification
type UrgencyLevel string

const (
	Low       UrgencyLevel = "low"
	Moderate  UrgencyLevel = "moderate"
	Critical  UrgencyLevel = "critical"
)

// UrgencyScore represents the score for a comment (0.0 - 3.0)
type UrgencyScore struct {
	Score       float64
	Level       UrgencyLevel
	Confidence  float64 // 0.0 - 1.0
}

// urgencyKeywords maps keywords to their urgency multiplier
var urgencyKeywords = map[string]float64{
	// Critical indicators (3.0x)
	"dangerous":       3.0,
	"critical":       3.0,
	"emergency":      3.0,
	"severe":         3.0,
	"urgent":         3.0,
	"fatal":          3.0,
	"death":          3.0,
	"dying":          3.0,
	"collapsed":      3.0,
	"collapse":       3.0,
	"broken":         2.5,
	"destroyed":      2.5,
	"accident":       2.5,
	"injury":         2.5,
	"injured":        2.5,
	"bleeding":       3.0,
	"fire":           3.0,
	"explod":         3.0,
	"hazard":         2.5,
	"gas":            2.5,

	// Moderate indicators (1.5x - 2.4x)
	"concern":        1.8,
	"serious":        2.0,
	"problem":        1.5,
	"issue":          1.2,
	"needs":          1.5,
	"needed":         1.5,
	"repair":         1.8,
	"damage":         2.0,
	"damaged":        2.0,
	"flood":          2.2,
	"flooding":       2.2,
	"waterlog":       2.2,
	"crack":          1.6,
	"hole":           1.5,
	"pothole":        1.8,
	"danger":         2.3,
	"risk":           2.0,
	"unsafe":         2.2,
	"sick":           2.0,
	"illness":        2.0,
	"disease":        2.0,
	"spread":         2.0,

	// Low indicators (0.5x - 1.4x)
	"minor":          0.8,
	"small":          0.7,
	"slight":         0.7,
	"bit":            0.6,
	"little":         0.6,
	"could":          0.9,
	"might":          0.9,
	"possible":       0.9,
	"maybe":          0.8,
	"suggests":       1.0,
	"seems":          0.9,
}

// CalculateCommentUrgency analyzes a comment string and returns an urgency score
func CalculateCommentUrgency(commentText string) UrgencyScore {
	if commentText == "" {
		return UrgencyScore{Score: 1.0, Level: Moderate, Confidence: 0.5}
	}

	text := strings.ToLower(commentText)
	words := strings.Fields(text)

	totalScore := 0.0
	matchCount := 0
	maxScore := 0.0

	// Calculate score from keywords
	for _, word := range words {
		// Remove punctuation for matching
		cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}") 

		// Check for direct matches
		if multiplier, found := urgencyKeywords[cleanWord]; found {
			totalScore += multiplier
			matchCount++
			if multiplier > maxScore {
				maxScore = multiplier
			}
			continue
		}

		// Check for partial matches (prefix matching)
		for keyword, multiplier := range urgencyKeywords {
			if strings.HasPrefix(cleanWord, keyword) {
				totalScore += multiplier * 0.8 // slightly lower confidence for partial matches
				matchCount++
				break
			}
		}
	}

	// Calculate average score
	var avgScore float64
	var confidence float64

	if matchCount == 0 {
		// No keywords found - default to moderate
		avgScore = 1.0
		confidence = 0.5
	} else {
		avgScore = totalScore / float64(matchCount)
		// Confidence increases with number of matches
		confidence = minFloat(1.0, float64(matchCount)*0.15)
	}

	// Clamp score to 0.0 - 3.0 range
	if avgScore > 3.0 {
		avgScore = 3.0
	}
	if avgScore < 0.5 {
		avgScore = 1.0
	}

	level := categorizeUrgency(avgScore)

	return UrgencyScore{
		Score:      avgScore,
		Level:      level,
		Confidence: confidence,
	}
}

// categorizeUrgency converts a numeric score to an urgency level
// Based on: low<=0.75, moderate<=1.5, critical>1.5
func categorizeUrgency(score float64) UrgencyLevel {
	if score <= 0.75 {
		return Low
	}
	if score <= 1.5 {
		return Moderate
	}
	return Critical
}

// CalculateAggregateUrgency combines post urgency with comment urgencies to get final score
// Algorithm: (PostUrgency * 0.5) + (Average(CommentScores) * 0.5)
func CalculateAggregateUrgency(postUrgency int, commentScores []float64) (int, UrgencyLevel) {
	// Start with post's initial urgency as baseline
	postScore := float64(postUrgency)

	// Calculate average comment urgency
	var commentAvg float64
	if len(commentScores) > 0 {
		totalComment := 0.0
		for _, score := range commentScores {
			totalComment += score
		}
		commentAvg = totalComment / float64(len(commentScores))
	} else {
		commentAvg = float64(postUrgency)
	}

	// Weighted average: 50% post, 50% comments
	// This allows comments to significantly influence the urgency
	finalScore := (postScore * 0.5) + (commentAvg * 0.5)

	// Convert to int (1-3 scale)
	finalInt := 1
	if finalScore > 2.25 {
		finalInt = 3
	} else if finalScore > 1.125 {
		finalInt = 2
	}

	level := categorizeUrgency(finalScore)

	return finalInt, level
}

// Helper function
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ParseURIFromInt converts the integer urgency (1-3) to float (0-3) for calculations
func IntToFloat(urgency int) float64 {
	switch urgency {
	case 1:
		return 1.0
	case 2:
		return 2.0
	case 3:
		return 3.0
	default:
		return 1.5
	}
}

// LogUrgencyCalculation logs the urgency calculation for debugging
func LogUrgencyCalculation(postID uuid.UUID, postUrgency int, commentScores []float64, finalUrgency int, finalLevel UrgencyLevel) {
	log.Printf(
		"[urgency_update] post_id=%s initial=%d comments_count=%d comment_avg=%.2f final=%d level=%s",
		postID.String(),
		postUrgency,
		len(commentScores),
		calculateAvgScore(commentScores),
		finalUrgency,
		finalLevel,
	)
}

func calculateAvgScore(scores []float64) float64 {
	if len(scores) == 0 {
		return 0
	}
	total := 0.0
	for _, s := range scores {
		total += s
	}
	return total / float64(len(scores))
}
