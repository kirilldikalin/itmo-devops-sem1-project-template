package storage

import "time"

type Product struct {
	ID         int
	Name       string
	Category   string
	Price      float64
	CreateDate time.Time
}