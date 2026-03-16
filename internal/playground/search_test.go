package playground

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newTestStateWithTweets creates a minimal State with users and tweets for search testing.
func newTestStateWithTweets() *State {
	s := &State{
		users:  make(map[string]*User),
		tweets: make(map[string]*Tweet),
		media:  make(map[string]*Media),
		lists:  make(map[string]*List),
		spaces: make(map[string]*Space),
		polls:  make(map[string]*Poll),
		places: make(map[string]*Place),
		topics: make(map[string]*Topic),
	}
	now := time.Now()

	// Create users
	s.users["u1"] = &User{ID: "u1", Name: "Clever Fox", Username: "CleverFox"}
	s.users["u2"] = &User{ID: "u2", Name: "Dolphin Tech", Username: "DolphinTech"}
	s.users["u3"] = &User{ID: "u3", Name: "Silent Owl", Username: "SilentOwl"}

	// Create tweets
	s.tweets["t1"] = &Tweet{
		ID:        "t1",
		Text:      "Just shipped a new open source API library for developers #coding",
		AuthorID:  "u1",
		CreatedAt: now.Add(-1 * time.Hour),
		Lang:      "en",
		Entities: &TweetEntities{
			Hashtags: []EntityHashtag{{Tag: "coding"}},
		},
	}
	s.tweets["t2"] = &Tweet{
		ID:        "t2",
		Text:      "Exploring the deep ocean with new underwater drones",
		AuthorID:  "u2",
		CreatedAt: now.Add(-2 * time.Hour),
		Lang:      "en",
	}
	s.tweets["t3"] = &Tweet{
		ID:        "t3",
		Text:      "La programación es el futuro de la tecnología",
		AuthorID:  "u3",
		CreatedAt: now.Add(-3 * time.Hour),
		Lang:      "es",
	}
	s.tweets["t4"] = &Tweet{
		ID:        "t4",
		Text:      "Retweeting the best coding tips from @CleverFox",
		AuthorID:  "u2",
		CreatedAt: now.Add(-30 * time.Minute),
		Lang:      "en",
		ReferencedTweets: []ReferencedTweet{
			{Type: "retweeted", ID: "t1"},
		},
		Entities: &TweetEntities{
			Mentions: []EntityMention{{Username: "CleverFox"}},
		},
	}

	return s
}

func TestSearchTweets_EmptyQuery(t *testing.T) {
	s := newTestStateWithTweets()
	results := s.SearchTweets(context.Background(), "", 100, "", "", nil, nil)
	assert.Len(t, results, 4, "empty query should return all tweets")
}

func TestSearchTweets_BareKeyword(t *testing.T) {
	s := newTestStateWithTweets()
	results := s.SearchTweets(context.Background(), "coding", 100, "", "", nil, nil)
	assert.NotEmpty(t, results, "bare keyword 'coding' should match tweets containing that word")
	for _, tw := range results {
		assert.Contains(t, tw.Text+tw.ID, "coding", "each result should relate to 'coding'")
	}
}

func TestSearchTweets_FromOperator(t *testing.T) {
	s := newTestStateWithTweets()
	results := s.SearchTweets(context.Background(), "from:CleverFox", 100, "", "", nil, nil)
	assert.NotEmpty(t, results)
	for _, tw := range results {
		assert.Equal(t, "u1", tw.AuthorID, "from:CleverFox should only return tweets by user u1")
	}
}

func TestSearchTweets_FromOperatorOR(t *testing.T) {
	s := newTestStateWithTweets()
	results := s.SearchTweets(context.Background(), "(from:CleverFox OR from:DolphinTech)", 100, "", "", nil, nil)
	assert.NotEmpty(t, results)
	for _, tw := range results {
		assert.Contains(t, []string{"u1", "u2"}, tw.AuthorID,
			"OR query should return tweets from either user")
	}
}

func TestSearchTweets_LangOperator(t *testing.T) {
	s := newTestStateWithTweets()
	results := s.SearchTweets(context.Background(), "lang:es", 100, "", "", nil, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, "t3", results[0].ID, "lang:es should match only the Spanish tweet")
}

func TestSearchTweets_NegationOperator(t *testing.T) {
	s := newTestStateWithTweets()
	// All English tweets except retweets
	results := s.SearchTweets(context.Background(), "lang:en -is:retweet", 100, "", "", nil, nil)
	for _, tw := range results {
		assert.Equal(t, "en", tw.Lang)
		for _, ref := range tw.ReferencedTweets {
			assert.NotEqual(t, "retweeted", ref.Type, "should exclude retweets")
		}
	}
}

func TestSearchTweets_HasHashtags(t *testing.T) {
	s := newTestStateWithTweets()
	results := s.SearchTweets(context.Background(), "has:hashtags", 100, "", "", nil, nil)
	assert.NotEmpty(t, results)
	for _, tw := range results {
		assert.NotNil(t, tw.Entities, "matched tweets should have entities")
		assert.NotEmpty(t, tw.Entities.Hashtags, "matched tweets should have hashtags")
	}
}

func TestSearchTweets_ComplexQuery(t *testing.T) {
	s := newTestStateWithTweets()
	// Test keyword + lang + negation combined with implicit AND
	results := s.SearchTweets(context.Background(), "from:CleverFox lang:en", 100, "", "", nil, nil)
	assert.NotEmpty(t, results)
	for _, tw := range results {
		assert.Equal(t, "u1", tw.AuthorID)
		assert.Equal(t, "en", tw.Lang)
	}
}

func TestSearchTweets_NoResults(t *testing.T) {
	s := newTestStateWithTweets()
	results := s.SearchTweets(context.Background(), "from:nonexistent_user_xyz", 100, "", "", nil, nil)
	assert.Empty(t, results, "non-existent user should return no results")
}

func TestSearchTweets_QuotedPhrase(t *testing.T) {
	s := newTestStateWithTweets()
	results := s.SearchTweets(context.Background(), "\"open source\"", 100, "", "", nil, nil)
	assert.NotEmpty(t, results)
	for _, tw := range results {
		assert.Contains(t, tw.Text, "open source")
	}
}

func TestSearchTweets_TimeFiltersStillWork(t *testing.T) {
	s := newTestStateWithTweets()
	now := time.Now()
	twoHoursAgo := now.Add(-2*time.Hour - 30*time.Minute)
	oneHourAgo := now.Add(-30 * time.Minute)

	results := s.SearchTweets(context.Background(), "", 100, "", "", &twoHoursAgo, &oneHourAgo)
	for _, tw := range results {
		assert.True(t, tw.CreatedAt.After(twoHoursAgo) || tw.CreatedAt.Equal(twoHoursAgo))
		assert.True(t, tw.CreatedAt.Before(oneHourAgo) || tw.CreatedAt.Equal(oneHourAgo))
	}
}
