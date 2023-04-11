package rwregister

import (
	"math"
	"time"

	"github.com/arangodb/go-driver"
)

func Profile(db driver.Database, dbConsts DBConsts, metadata map[string]int,
	f func(driver.Database, DBConsts, map[string]int, bool) bool, output bool) int64 {
	repeatingTimes := 10
	minTime := int64(math.MaxInt64)
	maxTime := int64(math.MinInt64)
	var totalTime int64
	for i := 0; i < repeatingTimes; i++ {
		start := time.Now()
		f(db, dbConsts, metadata, output)
		end := time.Now()
		temp := end.Sub(start).Nanoseconds() / 1e6
		totalTime += temp
		if temp < minTime {
			minTime = temp
		}
		if temp > maxTime {
			maxTime = temp
		}
	}
	return (totalTime - minTime - maxTime) / (int64(repeatingTimes) - 2)
}
