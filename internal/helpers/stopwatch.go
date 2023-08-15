package helpers

import "time"

type Stopwatch struct {
	startTime time.Time
}

func NewStopwatch() *Stopwatch {
	return &Stopwatch{startTime: time.Now()}
}

func (stopwatch *Stopwatch) Reset() {
	stopwatch.startTime = time.Now()
}

func (stopwatch *Stopwatch) Elapsed() time.Duration {
	return time.Since(stopwatch.startTime)
}

func (stopwatch *Stopwatch) Finish(time time.Duration) bool {
	return stopwatch.Elapsed() >= time
}
