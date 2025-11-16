package services

import (
	config "crowdsourcedurbanissuereportingwithai/backend/configs"
	"crowdsourcedurbanissuereportingwithai/backend/internal/repository"
	"crowdsourcedurbanissuereportingwithai/backend/models"
	"sort"
	"log"
	"strings"

	"github.com/google/uuid"
)

type FeedService struct {
	PostRepo *repository.PostRepository
}

func NewFeedService(postRepo *repository.PostRepository) *FeedService {
	return &FeedService{PostRepo: postRepo}
}

type ReportService struct {
	PostRepo *repository.PostRepository
}

func NewReportService(postRepo *repository.PostRepository) *ReportService {
	return &ReportService{PostRepo: postRepo}
}

func (s *ReportService) ReportIssueViaPost(userID, issueName, issueDesc, issueCat, postDesc, status string, urgency int, lat, lng float64, mediaURL string) (*models.Post, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	// Predict urgency and score from the description; use this score to initialize incremental scoring
	var initScore float64 = 0.0
	if u, sc, err := PredictUrgencyDetailed(postDesc); err == nil {
		if u != 0 { urgency = u }
		initScore = sc
	} else {
		// heuristics fallback is embedded in PredictUrgencyDetailed; call again to get a score
		_, sc, _ := PredictUrgencyDetailed(postDesc)
		initScore = sc
	}

	// If an image classification API is configured, attempt to classify the image.
	classifiedAs := ""
	if classified, err := ClassifyImage(mediaURL); err == nil && classified != "" {
		classifiedAs = classified
	} else if err != nil {
		// non-fatal: log and continue without classification
		// ClassifyImage logs details; we don't block reporting if classification fails.
	}

	post, err := s.PostRepo.ReportIssueViaPost(uid.String(), issueName, issueDesc, issueCat, postDesc, status, urgency, lat, lng, mediaURL, classifiedAs)
	if err != nil {
		return nil, err
	}
	// Initialize incremental score with the description score
	if err := s.PostRepo.UpdatePostScoreAdd(post.ID, initScore, 1); err != nil {
		log.Printf("warning: failed to initialize post score: %v", err)
	}
	return post, nil
}
func (s *FeedService) GetFeed() ([]models.Post, error) {
	posts, err := s.PostRepo.GetFeedPosts()
	if err != nil {
		return nil, err
	}

	// Determine scoring mode
	mode := config.GetFeedScoringMode() // ml | heuristic | none | incremental

	if mode == "incremental" {
		// Use persisted incremental average per post and blend with upvote presence
		for i := range posts {
			p := &posts[i]
			cnt := p.ScoreCount
			if cnt <= 0 { cnt = 1 }
			mlAvg := p.ScoreSum / float64(cnt)
			if mlAvg < 0 { mlAvg = 0 }
			if mlAvg > 1 { mlAvg = 1 }
			upvotePresence := 0.0
			if len(p.Upvotes) > 0 { upvotePresence = 1.0 }
			p.Score = 0.8*mlAvg + 0.2*upvotePresence
			p.ComputedUrgency = mapScoreToUrgency(mlAvg)
		}
		sort.SliceStable(posts, func(i, j int) bool { return posts[i].Score > posts[j].Score })
		return posts, nil
	}
	// Pre-compute max upvotes for normalization across the feed
	maxUpvotes := 0
	for i := range posts {
		if n := len(posts[i].Upvotes); n > maxUpvotes { maxUpvotes = n }
	}

	// Enrich each post with computed ml_score based on description and comments, then
	// blend with normalized upvotes: score = 0.8*ml_score + 0.2*votes_norm
	// Put a guardrail on total ML calls to keep the endpoint responsive.
	const maxMLCalls = 50
	calls := 0
	for i := range posts {
		p := &posts[i]
		// accumulate ML/heuristic scores for post description and each comment (range: 0..1)
		var scores []float64

		if p.Description != "" {
			if calls < maxMLCalls {
				var urg int
				var sc float64
				var err error
				switch mode {
				case "ml":
					urg, sc, err = PredictUrgencyDetailed(p.Description)
				case "heuristic":
					sc = heuristicScore(p.Description)
					urg = mapScoreToUrgency(sc)
				default: // none
					sc = mapNumericUrgencyToScore(p.Urgency)
					urg = p.Urgency
				}
				if err == nil {
					// use computed urgency only for the transient field; don't overwrite DB urgency
					p.ComputedUrgency = urg
					scores = append(scores, sc)
				}
				calls++
			}
		}
		// Sample only the first few comments per post for scoring to avoid long latency
		maxComments := 5
		for idx, c := range p.Comments {
			if idx >= maxComments || calls >= maxMLCalls {
				break
			}
			if c.Content == "" {
				continue
			}
			var sc float64
			var err error
			switch mode {
			case "ml":
				_, sc, err = PredictUrgencyDetailed(c.Content)
			case "heuristic":
				sc = heuristicScore(c.Content)
			default: // none
				// we don't have comment urgency persisted; skip scoring
				err = nil
				sc = -1 // sentinel to skip
			}
			if err == nil && sc >= 0 {
				scores = append(scores, sc)
			}
			calls++
		}

		// average ml score (0..1)
		var mlAvg float64
		if len(scores) > 0 {
			var sum float64
			for _, s := range scores { sum += s }
			mlAvg = sum / float64(len(scores))
		} else {
			mlAvg = 0.0
		}
		// upvote presence (0 or 1)
		var upvotePresence float64
		if len(p.Upvotes) > 0 { upvotePresence = 1.0 } else { upvotePresence = 0.0 }

		// final blended score per spec (independent of upvote count)
		blended := 0.8*mlAvg + 0.2*upvotePresence
		p.Score = blended
		// computed urgency from mlAvg only (not from votes)
		p.ComputedUrgency = mapScoreToUrgency(mlAvg)
	}

	// Sort by computed score descending (higher urgency first)
	sort.SliceStable(posts, func(i, j int) bool { return posts[i].Score > posts[j].Score })
	return posts, nil
}

// AddComment wraps repository call to add a comment
func (s *ReportService) AddComment(userID, postID, content string) (*models.Comment, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	pid, err := uuid.Parse(postID)
	if err != nil {
		return nil, err
	}
	comment, err := s.PostRepo.AddComment(uid, pid, content)
	if err != nil {
		return nil, err
	}

	// Incremental scoring: add this comment's score to the post average
	if strings.TrimSpace(content) != "" {
		_, sc, _ := PredictUrgencyDetailed(content)
		if err := s.PostRepo.UpdatePostScoreAdd(pid, sc, 1); err != nil {
			log.Printf("warning: failed to update post score after comment: %v", err)
		}
	}

	// IMPORTANT: Recalculate post urgency based on this new comment
	if err := s.UpdatePostUrgencyFromComments(pid); err != nil {
		// Non-fatal: log but don't fail the comment creation
		log.Printf("warning: failed to update post urgency after comment: %v", err)
	}

	return comment, nil
}

// ToggleUpvote wraps repository call to toggle an upvote
func (s *ReportService) ToggleUpvote(userID, postID string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	pid, err := uuid.Parse(postID)
	if err != nil {
		return false, err
	}
	return s.PostRepo.ToggleUpvote(uid, pid)
}

// UpdatePostUrgencyFromComments recalculates the post's urgency based on all its comments
// This is called after a new comment is added to dynamically update the post's priority
func (s *ReportService) UpdatePostUrgencyFromComments(postID uuid.UUID) error {
	// Fetch the post
	post, err := s.PostRepo.GetPost(postID)
	if err != nil {
		return err
	}

	// Fetch all comments for this post
	comments, err := s.PostRepo.GetPostComments(postID)
	if err != nil {
		return err
	}

	// Calculate urgency scores for each comment
	var commentScores []float64
	mode := config.GetFeedScoringMode()
	for _, comment := range comments {
		if strings.TrimSpace(comment.Content) == "" { continue }
		switch mode {
		case "ml":
			// Use ML score (0..1) scaled to 0..3 to match existing aggregator
			_, sc, _ := PredictUrgencyDetailed(comment.Content)
			commentScores = append(commentScores, sc*3.0)
		case "heuristic":
			// Use local heuristic (0..1) scaled to 0..3
			sc := heuristicScore(comment.Content)
			commentScores = append(commentScores, sc*3.0)
		default:
			// none: fallback to simple keyword-based calculator (already ~0..3)
			s := CalculateCommentUrgency(comment.Content)
			commentScores = append(commentScores, s.Score)
		}
	}

	// Calculate the new urgency level
	newUrgency, newLevel := CalculateAggregateUrgency(post.Urgency, commentScores)

	// Update the post's urgency in the database
	if err := s.PostRepo.UpdatePostUrgency(postID, newUrgency); err != nil {
		return err
	}

	// Log the urgency update for debugging
	LogUrgencyCalculation(postID, post.Urgency, commentScores, newUrgency, newLevel)

	return nil
}
