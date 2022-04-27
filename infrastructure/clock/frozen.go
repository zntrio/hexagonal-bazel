package clock

import "time"

func Frozen(dt time.Time) Clock {
	return &frozenClock{
		ref: dt,
	}
}

type frozenClock struct {
	ref time.Time
}

func (fc *frozenClock) Now() time.Time {
	return fc.ref
}
