package internal

import "time"

func timeZero() time.Time {
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
}

func timeFuture() time.Time {
	return time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
}

func timeNowDateStr() string {
	return time.Now().Format("2006-01-02")
}
