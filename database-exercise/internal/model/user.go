package model

import "time"

type User struct {
	ID        int `db:"id" gorm:"primaryKey"`
	Name      string
	Email     string
	CreatedAt time.Time
}
