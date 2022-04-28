package clock

import "time"

func Real() Clock {
	return &realClock{}
}

type realClock struct{}

func (rt *realClock) Now() time.Time {
	return time.Now().UTC()
}
