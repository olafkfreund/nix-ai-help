package community

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// DiscourseClient handles integration with NixOS Discourse forum
type DiscourseClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string // Optional, for authenticated requests
	username   string // Optional, for authenticated requests
	logger     *logger.Logger
}

// DiscoursePost represents a post from Discourse
type DiscoursePost struct {
	ID                 int        `json:"id"`
	Name               string     `json:"name"`
	Username           string     `json:"username"`
	AvatarTemplate     string     `json:"avatar_template"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	Cooked             string     `json:"cooked"`
	PostNumber         int        `json:"post_number"`
	PostType           int        `json:"post_type"`
	TopicID            int        `json:"topic_id"`
	TopicSlug          string     `json:"topic_slug"`
	DisplayUsername    string     `json:"display_username"`
	PrimaryGroupName   string     `json:"primary_group_name"`
	Version            int        `json:"version"`
	CanEdit            bool       `json:"can_edit"`
	CanDelete          bool       `json:"can_delete"`
	CanRecover         bool       `json:"can_recover"`
	UserTitle          string     `json:"user_title"`
	Raw                string     `json:"raw"`
	ActionsSummary     []Action   `json:"actions_summary"`
	Moderator          bool       `json:"moderator"`
	Admin              bool       `json:"admin"`
	Staff              bool       `json:"staff"`
	UserID             int        `json:"user_id"`
	Hidden             bool       `json:"hidden"`
	TrustLevel         int        `json:"trust_level"`
	DeletedAt          *time.Time `json:"deleted_at"`
	UserDeleted        bool       `json:"user_deleted"`
	EditReason         string     `json:"edit_reason"`
	CanViewEditHistory bool       `json:"can_view_edit_history"`
	Wiki               bool       `json:"wiki"`
}

// DiscourseTopic represents a topic from Discourse
type DiscourseTopic struct {
	ID                 int               `json:"id"`
	Title              string            `json:"title"`
	FancyTitle         string            `json:"fancy_title"`
	Slug               string            `json:"slug"`
	PostsCount         int               `json:"posts_count"`
	ReplyCount         int               `json:"reply_count"`
	HighestPostNumber  int               `json:"highest_post_number"`
	ImageURL           string            `json:"image_url"`
	CreatedAt          time.Time         `json:"created_at"`
	LastPostedAt       time.Time         `json:"last_posted_at"`
	Bumped             bool              `json:"bumped"`
	BumpedAt           time.Time         `json:"bumped_at"`
	Archetype          string            `json:"archetype"`
	Unseen             bool              `json:"unseen"`
	LastReadPostNumber int               `json:"last_read_post_number"`
	UnreadPosts        int               `json:"unread_posts"`
	NewPosts           int               `json:"new_posts"`
	Pinned             bool              `json:"pinned"`
	Unpinned           *time.Time        `json:"unpinned"`
	Excerpt            string            `json:"excerpt"`
	Visible            bool              `json:"visible"`
	Closed             bool              `json:"closed"`
	Archived           bool              `json:"archived"`
	NotificationLevel  int               `json:"notification_level"`
	Bookmarked         bool              `json:"bookmarked"`
	Liked              bool              `json:"liked"`
	Tags               []string          `json:"tags"`
	TagsDescriptions   map[string]string `json:"tags_descriptions"`
	CategoryID         int               `json:"category_id"`
	FeaturedLink       string            `json:"featured_link"`
	HasSummary         bool              `json:"has_summary"`
	Views              int               `json:"views"`
	LikeCount          int               `json:"like_count"`
	ParticipantCount   int               `json:"participant_count"`
	WordCount          int               `json:"word_count"`
	Posts              []DiscoursePost   `json:"posts_stream,omitempty"`
}

// DiscourseCategory represents a category from Discourse
type DiscourseCategory struct {
	ID                       int                    `json:"id"`
	Name                     string                 `json:"name"`
	Color                    string                 `json:"color"`
	TextColor                string                 `json:"text_color"`
	Slug                     string                 `json:"slug"`
	TopicCount               int                    `json:"topic_count"`
	PostCount                int                    `json:"post_count"`
	Position                 int                    `json:"position"`
	Description              string                 `json:"description"`
	DescriptionText          string                 `json:"description_text"`
	DescriptionExcerpt       string                 `json:"description_excerpt"`
	TopicURL                 string                 `json:"topic_url"`
	ReadRestricted           bool                   `json:"read_restricted"`
	Permission               int                    `json:"permission"`
	ParentCategoryID         *int                   `json:"parent_category_id"`
	TopicsDay                int                    `json:"topics_day"`
	TopicsWeek               int                    `json:"topics_week"`
	TopicsMonth              int                    `json:"topics_month"`
	TopicsYear               int                    `json:"topics_year"`
	TopicsAllTime            int                    `json:"topics_all_time"`
	SubcategoryIDs           []int                  `json:"subcategory_ids"`
	CanEdit                  bool                   `json:"can_edit"`
	NotificationLevel        int                    `json:"notification_level"`
	CanMovePosts             bool                   `json:"can_move_posts"`
	HasChildren              bool                   `json:"has_children"`
	SortOrder                string                 `json:"sort_order"`
	SortAscending            bool                   `json:"sort_ascending"`
	ShowSubcategoryList      bool                   `json:"show_subcategory_list"`
	NumFeaturedTopics        int                    `json:"num_featured_topics"`
	DefaultView              string                 `json:"default_view"`
	SubcategoryListStyle     string                 `json:"subcategory_list_style"`
	DefaultTopPeriod         string                 `json:"default_top_period"`
	MinimumRequiredTags      int                    `json:"minimum_required_tags"`
	NavigateToFirstPost      bool                   `json:"navigate_to_first_post_after_read"`
	TopicsAllowedBadges      []Badge                `json:"topics_allowed_badges"`
	CustomFields             map[string]interface{} `json:"custom_fields"`
	AllowedTags              []string               `json:"allowed_tags"`
	AllowedTagGroups         []string               `json:"allowed_tag_groups"`
	RequiredTagGroups        []TagGroup             `json:"required_tag_groups"`
	MinTagsFromRequiredGroup int                    `json:"min_tags_from_required_group"`
}

// DiscourseUser represents a user from Discourse
type DiscourseUser struct {
	ID               int                    `json:"id"`
	Username         string                 `json:"username"`
	Name             string                 `json:"name"`
	AvatarTemplate   string                 `json:"avatar_template"`
	Email            string                 `json:"email"`
	LastPostedAt     time.Time              `json:"last_posted_at"`
	LastSeenAt       time.Time              `json:"last_seen_at"`
	CreatedAt        time.Time              `json:"created_at"`
	IgnoredUsers     []string               `json:"ignored_users"`
	MutedUsers       []string               `json:"muted_users"`
	CanEdit          bool                   `json:"can_edit"`
	CanEditUsername  bool                   `json:"can_edit_username"`
	CanEditEmail     bool                   `json:"can_edit_email"`
	CanEditName      bool                   `json:"can_edit_name"`
	UploadedAvatarID int                    `json:"uploaded_avatar_id"`
	HasTitleBadges   bool                   `json:"has_title_badges"`
	Moderator        bool                   `json:"moderator"`
	Admin            bool                   `json:"admin"`
	TrustLevel       int                    `json:"trust_level"`
	UserFields       map[string]string      `json:"user_fields"`
	CustomFields     map[string]interface{} `json:"custom_fields"`
	PendingCount     int                    `json:"pending_count"`
	ProfileHidden    bool                   `json:"profile_hidden"`
}

// Action represents an action that can be performed on a post
type Action struct {
	ID     int  `json:"id"`
	Count  int  `json:"count"`
	Hidden bool `json:"hidden"`
	CanAct bool `json:"can_act"`
}

// Badge represents a badge in Discourse
type Badge struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	GrantCount    int           `json:"grant_count"`
	AllowTitle    bool          `json:"allow_title"`
	Multiple      bool          `json:"multiple"`
	Icon          string        `json:"icon"`
	Image         string        `json:"image"`
	Listable      bool          `json:"listable"`
	Enabled       bool          `json:"enabled"`
	BadgeGrouping BadgeGrouping `json:"badge_grouping"`
	TriggerType   int           `json:"trigger_type"`
	Target        bool          `json:"target"`
	Query         string        `json:"query"`
	AutoRevoke    bool          `json:"auto_revoke"`
	ShowPosts     bool          `json:"show_posts"`
	System        bool          `json:"system"`
}

// BadgeGrouping represents a badge grouping
type BadgeGrouping struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

// TagGroup represents a tag group
type TagGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DiscourseSearchResult represents search results from Discourse API
type DiscourseSearchResult struct {
	Posts               []DiscoursePost     `json:"posts"`
	Topics              []DiscourseTopic    `json:"topics"`
	Users               []DiscourseUser     `json:"users"`
	Categories          []DiscourseCategory `json:"categories"`
	Tags                []string            `json:"tags"`
	GroupedSearchResult GroupedSearchResult `json:"grouped_search_result"`
}

// GroupedSearchResult represents grouped search results
type GroupedSearchResult struct {
	PostIDs        []int         `json:"post_ids"`
	UserIDs        []int         `json:"user_ids"`
	CategoryIDs    []int         `json:"category_ids"`
	TagIDs         []int         `json:"tag_ids"`
	MorePosts      bool          `json:"more_posts"`
	MoreUsers      bool          `json:"more_users"`
	MoreCategories bool          `json:"more_categories"`
	Term           string        `json:"term"`
	SearchLogID    int           `json:"search_log_id"`
	SearchContext  SearchContext `json:"search_context"`
}

// SearchContext represents the context of a search
type SearchContext struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DiscourseTopicsResponse represents the response from topics API
type DiscourseTopicsResponse struct {
	Users         []DiscourseUser `json:"users"`
	PrimaryGroups []interface{}   `json:"primary_groups"`
	TopicList     TopicList       `json:"topic_list"`
}

// TopicList represents a list of topics
type TopicList struct {
	CanCreateTopic bool             `json:"can_create_topic"`
	MoreTopicsURL  string           `json:"more_topics_url"`
	DraftKey       string           `json:"draft_key"`
	DraftSequence  int              `json:"draft_sequence"`
	PerPage        int              `json:"per_page"`
	Topics         []DiscourseTopic `json:"topics"`
}

// NewDiscourseClient creates a new Discourse API client
func NewDiscourseClient(baseURL string, apiKey string, username string) *DiscourseClient {
	if baseURL == "" {
		baseURL = "https://discourse.nixos.org"
	}

	return &DiscourseClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  baseURL,
		apiKey:   apiKey,
		username: username,
		logger:   logger.NewLoggerWithLevel("info"),
	}
}

// SearchPosts searches for posts in Discourse
func (d *DiscourseClient) SearchPosts(ctx context.Context, query string, limit int) (*DiscourseSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	params := url.Values{
		"q":         []string{query},
		"max_posts": []string{strconv.Itoa(limit)},
	}

	endpoint := fmt.Sprintf("%s/search.json?%s", d.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if available
	if d.apiKey != "" && d.username != "" {
		req.Header.Set("Api-Key", d.apiKey)
		req.Header.Set("Api-Username", d.username)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "nixai-community-client/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result DiscourseSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetTopicsByCategory retrieves topics from a specific category
func (d *DiscourseClient) GetTopicsByCategory(ctx context.Context, categorySlug string, limit int) (*DiscourseTopicsResponse, error) {
	if limit <= 0 {
		limit = 30
	}

	endpoint := fmt.Sprintf("%s/c/%s.json", d.baseURL, categorySlug)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if available
	if d.apiKey != "" && d.username != "" {
		req.Header.Set("Api-Key", d.apiKey)
		req.Header.Set("Api-Username", d.username)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "nixai-community-client/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result DiscourseTopicsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Apply limit to topics
	if len(result.TopicList.Topics) > limit {
		result.TopicList.Topics = result.TopicList.Topics[:limit]
	}

	return &result, nil
}

// GetTopic retrieves a specific topic with its posts
func (d *DiscourseClient) GetTopic(ctx context.Context, topicID int) (*DiscourseTopic, error) {
	endpoint := fmt.Sprintf("%s/t/%d.json", d.baseURL, topicID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if available
	if d.apiKey != "" && d.username != "" {
		req.Header.Set("Api-Key", d.apiKey)
		req.Header.Set("Api-Username", d.username)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "nixai-community-client/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result DiscourseTopic
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetCategories retrieves all categories
func (d *DiscourseClient) GetCategories(ctx context.Context) ([]DiscourseCategory, error) {
	endpoint := fmt.Sprintf("%s/categories.json", d.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if available
	if d.apiKey != "" && d.username != "" {
		req.Header.Set("Api-Key", d.apiKey)
		req.Header.Set("Api-Username", d.username)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "nixai-community-client/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		CategoryList struct {
			Categories []DiscourseCategory `json:"categories"`
		} `json:"category_list"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.CategoryList.Categories, nil
}

// GetLatestTopics retrieves the latest topics across all categories
func (d *DiscourseClient) GetLatestTopics(ctx context.Context, limit int) (*DiscourseTopicsResponse, error) {
	if limit <= 0 {
		limit = 30
	}

	endpoint := fmt.Sprintf("%s/latest.json", d.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if available
	if d.apiKey != "" && d.username != "" {
		req.Header.Set("Api-Key", d.apiKey)
		req.Header.Set("Api-Username", d.username)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "nixai-community-client/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result DiscourseTopicsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Apply limit to topics
	if len(result.TopicList.Topics) > limit {
		result.TopicList.Topics = result.TopicList.Topics[:limit]
	}

	return &result, nil
}

// GetTopTopics retrieves the top topics for a given period
func (d *DiscourseClient) GetTopTopics(ctx context.Context, period string, limit int) (*DiscourseTopicsResponse, error) {
	if limit <= 0 {
		limit = 30
	}

	// Valid periods: daily, weekly, monthly, yearly, all
	validPeriods := map[string]bool{
		"daily": true, "weekly": true, "monthly": true, "yearly": true, "all": true,
	}

	if !validPeriods[period] {
		period = "weekly" // Default to weekly
	}

	endpoint := fmt.Sprintf("%s/top/%s.json", d.baseURL, period)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if available
	if d.apiKey != "" && d.username != "" {
		req.Header.Set("Api-Key", d.apiKey)
		req.Header.Set("Api-Username", d.username)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "nixai-community-client/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result DiscourseTopicsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Apply limit to topics
	if len(result.TopicList.Topics) > limit {
		result.TopicList.Topics = result.TopicList.Topics[:limit]
	}

	return &result, nil
}

// ConvertToConfiguration converts a Discourse topic to a Configuration struct
func (d *DiscourseClient) ConvertToConfiguration(topic DiscourseTopic, user DiscourseUser) Configuration {
	// Extract tags and convert to lowercase for consistency
	tags := make([]string, len(topic.Tags))
	for i, tag := range topic.Tags {
		tags[i] = strings.ToLower(tag)
	}

	// Calculate a rating based on like count and views (simplified scoring)
	rating := float64(topic.LikeCount) / float64(max(topic.Views, 1)) * 10
	if rating > 5.0 {
		rating = 5.0
	}

	return Configuration{
		ID:          fmt.Sprintf("discourse_%d", topic.ID),
		Name:        topic.Title,
		Author:      user.Username,
		Description: topic.Excerpt,
		Tags:        tags,
		Rating:      rating,
		Downloads:   0, // Discourse doesn't track downloads
		Views:       topic.Views,
		URL:         fmt.Sprintf("%s/t/%s/%d", d.baseURL, topic.Slug, topic.ID),
		FilePath:    "", // Not applicable for Discourse topics
		Content:     "", // Would need to fetch topic content separately
		CreatedAt:   topic.CreatedAt,
		UpdatedAt:   topic.LastPostedAt,
		Size:        int64(topic.WordCount),
		Language:    "markdown", // Discourse posts are typically markdown
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
