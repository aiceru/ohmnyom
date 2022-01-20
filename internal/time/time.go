package time

import (
	"time"

	"github.com/aiceru/protonyom/gonyom"
)

var UTC = time.UTC

type Time struct {
	time.Time
}

func Now() Time {
	return Time{time.Now()}
}

// Date returns time.Date that rounded down nsec to microsecond, due to limitation of firestore's time data type.
func Date(year int, month time.Month, day int, hour int, min int, sec int, nsec int, loc *time.Location) Time {
	return Time{
		time.Date(year, month, day, hour, min, sec, nsec/int(time.Microsecond), loc),
	}
}

func (t *Time) Proto() *gonyom.Timestamp {
	nano := t.UnixNano()
	return &gonyom.Timestamp{
		Seconds: nano / int64(time.Second),
		Nanos:   int32(nano % int64(time.Second)),
	}
}
