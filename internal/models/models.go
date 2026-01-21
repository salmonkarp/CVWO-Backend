package models

type User struct {
	ID             int     `json:"id"`
	Username       string  `json:"username"`
	ImageURL       *string `json:"imageUrl"`
	ImageUpdatedAt int64   `json:"imageUpdatedAt,omitempty"`
}

type Topic struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	ImageURL       *string `json:"imageUrl"`
	ImageUpdatedAt int64   `json:"imageUpdatedAt,omitempty"`
}

type Post struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Topic     string `json:"topic"`
	Creator   int    `json:"creator"`
	CreatedAt string `json:"created_at"`
	IsEdited  bool   `json:"is_edited"`
}

type Comment struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	Post      int    `json:"post"`
	Creator   int    `json:"creator"`
	CreatedAt string `json:"created_at"`
	IsEdited  bool   `json:"is_edited"`
	Parent    *int   `json:"parent,omitempty"`
}
