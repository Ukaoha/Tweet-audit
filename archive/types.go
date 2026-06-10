package archive

type tweetWrapper struct {
	Tweet Tweet `json:"tweet"`
}

type Tweet struct {
	IDStr         string   `json:"id_str"`
	FullText      string   `json:"full_text"`
	CreatedAt     string   `json:"created_at"`
	FavoriteCount string   `json:"favorite_count"`
	RetweetCount  string   `json:"retweet_count"`
	Lang          string   `json:"lang"`
	Retweeted     bool     `json:"retweeted"`
	Entities      Entities `json:"entities"`
}

type Entities struct {
	Hashtags     []Hashtag `json:"hashtags"`
	UserMentions []any     `json:"user_mentions"`
	URLs         []URL     `json:"urls"`
}

type Hashtag struct {
	Text string `json:"text"`
}

type URL struct {
	URL         string `json:"url"`
	ExpandedURL string `json:"expanded_url"`
	DisplayURL  string `json:"display_url"`
}

type accountWrapper struct {
	Account Account `json:"account"`
}

type Account struct {
	Username  string `json:"username"`
	AccountID string `json:"accountId"`
}
