package main

import "time"

// message는 단일 메시지를 나타냄
type message struct {
	Name    string
	Message string
	When    time.Time
}
