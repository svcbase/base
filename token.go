package base

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/svcbase/ipaddr"
)

type TokenSet struct {
	sm     sync.Mutex
	tokens map[string]int
}

func (ts *TokenSet) Init() {
	ts.tokens = make(map[string]int)
}

func (ts *TokenSet) NewToken(r *http.Request, seconds int) (token string) {
	ipaddress := int64(ipaddr.ClientIP(r))
	unixnano := time.Now().UnixNano()
	token = EncodeByKey(strconv.FormatInt(ipaddress, 10)+":"+strconv.FormatInt(unixnano, 10), DeepdataKeys())
	ts.sm.Lock()
	ts.tokens[token] = seconds
	ts.sm.Unlock()
	return
}

func (ts *TokenSet) RemoveToken(token string) {
	ts.sm.Lock()
	if _, ok := ts.tokens[token]; ok {
		delete(ts.tokens, token)
	}
	ts.sm.Unlock()
}

func (ts *TokenSet) WeedExpires() (n int) {
	n = 0
	ts.sm.Lock()
	for token, seconds := range ts.tokens {
		ss := strings.Split(DecodeByKey(token, DeepdataKeys()), ":")
		if len(ss) == 2 {
			unixnano := Str2int64(ss[1])
			tm := time.Unix(0, unixnano).Add(time.Duration(seconds) * time.Second)
			if time.Now().After(tm) {
				delete(ts.tokens, token)
				n++
			}
		}
	}
	ts.sm.Unlock()
	return
}

func (ts *TokenSet) ValidToken(r *http.Request, token string) (flag bool) {
	flag = false
	if len(token) > 0 {
		ts.sm.Lock()
		if seconds, ok := ts.tokens[token]; ok {
			ss := strings.Split(DecodeByKey(token, DeepdataKeys()), ":")
			if len(ss) == 2 {
				unixnano := Str2int64(ss[1])
				ipaddress := int64(ipaddr.ClientIP(r))
				if strconv.FormatInt(ipaddress, 10) == ss[0] {
					tm := time.Unix(0, unixnano).Add(time.Duration(seconds) * time.Second)
					flag = time.Now().Before(tm)
				}
			}
		}
		ts.sm.Unlock()
	}
	return
}
