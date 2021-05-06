package cache

import "time"

type Item struct {
	Key        string
	Value      interface{}
	Expiration time.Duration
}