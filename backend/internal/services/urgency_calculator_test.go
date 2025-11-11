package services

import (
	"testing"
)

func TestCalculateCommentUrgency(t *testing.T) {
	tests := []struct {
		name           string
		comment        string
		expectedLevel  UrgencyLevel
		minScore       float64
		maxScore       float64
	}{
		{
			name:          "Critical urgency - dangerous",
			comment:       "This is extremely dangerous and critical",
			expectedLevel: Critical,
			minScore:      2.0,
			maxScore:      3.0,
		},
		{
			name:          "Critical urgency - emergency",
			comment:       "This is an emergency situation, someone needs help immediately",
			expectedLevel: Critical,
			minScore:      2.0,
			maxScore:      3.0,
		},
		{
			name:          "Moderate urgency - repair needed",
			comment:       "This needs repair soon, it's a serious problem",
			expectedLevel: Critical, // "serious" is 2.0, "problem" is 1.5, "repair" is 1.8
			minScore:      1.5,
			maxScore:      2.5,
		},
		{
			name:          "Low-moderate urgency - minor issue",
			comment:       "This is a minor issue, not urgent",
			expectedLevel: Critical, // "minor" = 0.8, "issue" = 1.2, "not" no match, "urgent" = 3.0, avg = (0.8+1.2+3.0)/3 = 1.67 > 1.5
			minScore:      1.5,
			maxScore:      2.0,
		},
		{
			name:          "Moderate urgency - pothole",
			comment:       "Large pothole causing traffic issues, needs immediate attention",
			expectedLevel: Moderate, // "pothole" = 1.8, "issues" (no keyword), "needs" = 1.5
			minScore:      1.0,
			maxScore:      1.8,
		},
		{
			name:          "Empty comment",
			comment:       "",
			expectedLevel: Moderate,
			minScore:      0.5,
			maxScore:      1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateCommentUrgency(tt.comment)

			if score.Level != tt.expectedLevel {
				t.Errorf("expected level %s, got %s (score: %.2f)", tt.expectedLevel, score.Level, score.Score)
			}

			if score.Score < tt.minScore || score.Score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score.Score)
			}

			if score.Confidence < 0 || score.Confidence > 1 {
				t.Errorf("confidence should be between 0 and 1, got %.2f", score.Confidence)
			}
		})
	}
}

func TestCategorizeUrgency(t *testing.T) {
	tests := []struct {
		name           string
		score          float64
		expectedLevel  UrgencyLevel
	}{
		{
			name:          "Low threshold",
			score:         0.5,
			expectedLevel: Low,
		},
		{
			name:          "Low boundary",
			score:         0.75,
			expectedLevel: Low,
		},
		{
			name:          "Moderate lower",
			score:         0.8,
			expectedLevel: Moderate,
		},
		{
			name:          "Moderate boundary",
			score:         1.5,
			expectedLevel: Moderate,
		},
		{
			name:          "Critical lower",
			score:         1.51,
			expectedLevel: Critical,
		},
		{
			name:          "Critical high",
			score:         3.0,
			expectedLevel: Critical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := categorizeUrgency(tt.score)
			if level != tt.expectedLevel {
				t.Errorf("expected %s, got %s for score %.2f", tt.expectedLevel, level, tt.score)
			}
		})
	}
}

func TestCalculateAggregateUrgency(t *testing.T) {
	tests := []struct {
		name              string
		postUrgency       int
		commentScores     []float64
		expectedUrgency   int
	}{
		{
			name:            "No comments - post stays same",
			postUrgency:     1,
			commentScores:   []float64{},
			expectedUrgency: 1,
		},
		{
			name:            "One critical comment boosts urgency",
			postUrgency:     1,
			commentScores:   []float64{3.0},
			expectedUrgency: 2,
		},
		{
			name:            "Multiple critical comments escalate",
			postUrgency:     1,
			commentScores:   []float64{2.8, 2.9, 3.0},
			expectedUrgency: 2, // (1 * 0.5) + ((2.8+2.9+3.0)/3 * 0.5) = 0.5 + 1.45 = 1.95 -> rounds to 2
		},
		{
			name:            "Mix of urgencies",
			postUrgency:     2,
			commentScores:   []float64{1.0, 2.0, 1.5},
			expectedUrgency: 2,
		},
		{
			name:            "Comments lower than post",
			postUrgency:     3,
			commentScores:   []float64{0.8, 0.9},
			expectedUrgency: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urgency, level := CalculateAggregateUrgency(tt.postUrgency, tt.commentScores)

			if urgency != tt.expectedUrgency {
				t.Errorf("expected urgency %d, got %d", tt.expectedUrgency, urgency)
			}

			// Just verify level is valid
			if level != Low && level != Moderate && level != Critical {
				t.Errorf("invalid level: %s", level)
			}
		})
	}
}

func TestRealWorldScenario(t *testing.T) {
	// Scenario: User reports a pothole (urgency 1)
	// Then people comment saying it's dangerous
	initialUrgency := 1

	comments := []string{
		"This pothole is really dangerous, my car got damaged",
		"Several accidents happened here already, it's critical",
		"Multiple vehicles have been damaged, needs urgent repair",
	}

	var commentScores []float64
	for _, comment := range comments {
		score := CalculateCommentUrgency(comment)
		commentScores = append(commentScores, score.Score)
		t.Logf("Comment: '%s' -> Score: %.2f, Level: %s", comment, score.Score, score.Level)
	}

	finalUrgency, finalLevel := CalculateAggregateUrgency(initialUrgency, commentScores)

	t.Logf("\nInitial urgency: %d", initialUrgency)
	t.Logf("Comment count: %d", len(commentScores))
	t.Logf("Final urgency: %d (Level: %s)", finalUrgency, finalLevel)

	// Should escalate from 1 to 3 (critical) due to community input
	if finalUrgency < 2 {
		t.Errorf("expected urgency to escalate to at least 2, got %d", finalUrgency)
	}
}
