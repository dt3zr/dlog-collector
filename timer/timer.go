package timer

import "github.com/robfig/cron"

func NewForEveryMinute(done <-chan struct{}) <-chan struct{} {
	timeChan := make(chan struct{})
	go func() {
		defer close(timeChan)
		c := cron.New()
		c.AddFunc("0 * * * * *", func() {
			timeChan <- struct{}{}
		})
		c.Start()
		<-done
	}()
	return timeChan
}
