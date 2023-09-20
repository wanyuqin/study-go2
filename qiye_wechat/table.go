package main

type UserCheckInLog struct {
	Name            string `json:"name"`
	Date            string `json:"date"`
	Dept            string `json:"dept"`
	Position        string `json:"position"`
	EarliestTime    string `json:"earliest_time"`
	LatestTime      string `json:"latest_time"`
	RegularWorkTime string `json:"regular_work_time"`
	AvaWorkTime     string `json:"ava_work_time"`
}
