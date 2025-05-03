package model

import "time"

type Person struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Patronymic  string    `json:"patronymic,omitempty"`
	Age         int       `json:"age,omitempty"`
	Gender      string    `json:"gender,omitempty"`
	Nationality string    `json:"nationality,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PersonInput struct {
	Name       string `json:"name" binding:"required"`
	Surname    string `json:"surname" binding:"required"`
	Patronymic string `json:"patronymic,omitempty"`
}

type PersonFilter struct {
	Gender      string
	Nationality string
	AgeFrom     int
	AgeTo       int
	Page        int
	Limit       int
}