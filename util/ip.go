package util

import (
	"github.com/patrickmn/go-cache"
	"time"
)

type IpRejectorStruct struct {
	IpRejectorCache *cache.Cache
	errorMaxTimes   int
}

var IpRejector *IpRejectorStruct

func init() {
	IpRejector = &IpRejectorStruct{
		IpRejectorCache: cache.New(86400*time.Second, 86400*time.Second),
		errorMaxTimes:   5,
	}
}

func (rejector IpRejectorStruct) Pass(ip string) bool {
	times, ok := rejector.IpRejectorCache.Get(ip)
	if !ok {
		return true
	}
	if times.(int) > rejector.errorMaxTimes {
		return false
	}
	return true
}

func (rejector IpRejectorStruct) AddErrorTimes(ip string) {
	if _, ok := rejector.IpRejectorCache.Get(ip); ok {
		rejector.IpRejectorCache.Increment(ip, 1)
	} else {
		now := time.Now()
		tomorrow := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, 0, 0, 0, 0, time.Local)
		rejector.IpRejectorCache.Set(ip, 1, tomorrow.Sub(now))
	}
}
