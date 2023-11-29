package util

import (
	"github.com/patrickmn/go-cache"
	"time"
)

type IpRejectorStruct struct {
	ipRejectorCache *cache.Cache
	errorMaxTimes   int
}

var IpRejector *IpRejectorStruct

func init() {
	IpRejector = &IpRejectorStruct{
		ipRejectorCache: cache.New(86400*time.Second, 86400*time.Second),
		errorMaxTimes:   5,
	}
}

func (rejector IpRejectorStruct) Pass(cookie string) bool {
	times, ok := rejector.ipRejectorCache.Get(cookie)
	if !ok {
		return true
	}
	if times.(int) > rejector.errorMaxTimes {
		return false
	}
	return true
}

func (rejector IpRejectorStruct) AddErrorTimes(cookie string) {
	if _, ok := rejector.ipRejectorCache.Get(cookie); ok {
		rejector.ipRejectorCache.Increment(cookie, 1)
	} else {
		now := time.Now()
		tomorrow := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, 0, 0, 0, 0, time.Local)
		rejector.ipRejectorCache.Set(cookie, 1, tomorrow.Sub(now))
	}
}
