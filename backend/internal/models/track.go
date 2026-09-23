package models

import "time"

type Track struct {
	ID         int       `json:"id" db:"id"`
	User_id    string    `json:"-" db:"user_id"`
	Title      string    `json:"title" db:"title"`
	Cover_key  *string   `json:"-" db:"cover_key"`
	Cover_url  string    `json:"cover_url,omitempty" db:"-"`
	Object_key string    `json:"-" db:"object_key"`
	Created_at time.Time `json:"created_at" db:"created_at"`
}

type TrackCreateInput struct {
	User_id    string  `json:"user_id"`
	Title      string  `json:"title"`
	Cover_key  *string `json:"-"`
	Object_key string  `json:"object_key"`
}
