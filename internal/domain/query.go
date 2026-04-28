package domain

// BaseQuery berisi parameter standard untuk filter query dan pagination
type BaseQuery struct {
	Search    string
	SortBy    string
	SortOrder string
	Page      int64
	Limit     int64
}
