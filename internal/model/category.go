package model

type Category struct {
	Id         int     `json:"id"`
	Nome       string  `json:"nome"`
	Quantidade float64 `json:"quantidade"`
	Limite     int64   `json:"limite"`
}
