package audit

type Criteria struct {
	ForbiddenWords     []string `json:"forbidden_words"`
	ProfessionalCheck  bool     `json:"professional_check"`
	Tone               string   `json:"tone"`
	ExcludePolitics    bool     `json:"exclude_politics"`
	CustomRules        []string `json:"custom_rules"`
	IncludeReplies     bool     `json:"include_replies"`
	IncludeRetweets    bool     `json:"include_retweets"`
	MinFavoritesToKeep int      `json:"min_favorites_to_keep"`
	MinRetweetsToKeep  int      `json:"min_retweets_to_keep"`
}

type Verdict struct {
	TweetID   string
	Username  string
	URL       string
	Text      string
	Flag      bool
	Reason    string
	CreatedAt string
}
