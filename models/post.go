package models

type Post struct {
	Id           int       `json:"id"`
	Body         string    `json:"body"`
	TableName    string    `json:"table_name"`
	UserId       int       `json:"user_id"`
	SecretUserId int       `json:"-"`
	UserName     string    `json:"user_name"`
	Nimg         string    `json:"nimg"`
	Comments     int       `json:"comments"`
	Likes        int       `json:"likes"`
	Liked        bool      `json:"liked"`
	Rank         int       `json:"rank"`
	Created      time.Time `json:"created"`
}
