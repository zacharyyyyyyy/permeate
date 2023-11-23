package util

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type session struct {
	sid string
}

const (
	SidLength   = 32
	sessionName = "permeate_id"
	LifeTime    = 86400 * time.Second
)

var (
	sidReg     = regexp.MustCompile(fmt.Sprintf("[a-z0-9]{%d}", SidLength))
	localCache = cache.New(86400*time.Second, 86400*time.Second)
)

func NewSession(w http.ResponseWriter, r *http.Request) *session {
	var sid string
	cookie, _ := r.Cookie(sessionName)
	if cookie == nil || !sidReg.MatchString(cookie.Value) {
		sid = genSid()
	} else {
		sid = cookie.Value
	}
	http.SetCookie(w, &http.Cookie{Name: sessionName, Value: sid, Path: "/", HttpOnly: false, Secure: true, Expires: time.Now().Add(LifeTime)})
	return &session{
		sid: sid,
	}
}

func (s session) Login() {
	localCache.Set(s.sid, 1, LifeTime)
}

func (s session) HasAuth() bool {
	_, ok := localCache.Get(s.sid)
	return ok
}

func genSid() string {
	strBuilder := strings.Builder{}
	strBuilder.Grow(SidLength)
	var str = "0123456789abcdefghijklmnopqrstuvwxyz"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < SidLength; i++ {
		strBuilder.WriteByte(str[r.Intn(len(str))])
	}
	return strBuilder.String()
}
