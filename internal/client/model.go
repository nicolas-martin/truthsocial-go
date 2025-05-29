package client

type Status struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Account   struct {
		Username string `json:"username"`
	} `json:"account"`
	URL             string `json:"url"`
	InReplyToID     string `json:"in_reply_to_id"`
	ReblogsCount    int    `json:"reblogs_count"`
	FavouritesCount int    `json:"favourites_count"`
	RepliesCount    int    `json:"replies_count"`
}

type Account struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	DisplayName    string `json:"display_name"`
	FollowersCount int    `json:"followers_count"`
	StatusesCount  int    `json:"statuses_count"`
	Verified       bool   `json:"verified"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}
