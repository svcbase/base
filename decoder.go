package base

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

func SM2_decode(privateKey, encodedData string) (out string, e error) {
	if len(privateKey) > 0 && len(encodedData) > 0 {
		priv, ee := x509.ReadPrivateKeyFromPem([]byte(privateKey), nil)
		if ee == nil {
			var b, bb, c []byte
			b, e = base64.StdEncoding.DecodeString(encodedData)
			if e == nil {
				if b[0] != 0x04 {
					bb = append(bb, 0x04)
					bb = append(bb, b...)
				} else {
					bb = b
				}
				c, e = sm2.Decrypt(priv, bb, 0)
				if e == nil {
					out = string(c)
				}
			}
		} else {
			e = ee
		}
	} else {
		e = errors.New("no privateKey.")
	}
	return
}

func DecodingQuotaConsume() (overquota bool) {
	today := Str10Unixtime(time.Now().Format("2006-01-02"))
	overquota = true
	if TodayDecodingNumber == 0 {
		ReadQuotaConsumed()
	}
	if TodayDecodingNumber < DailyDecodingQuota {
		TodayDecodingNumber++
		ss := today + "=" + strconv.Itoa(TodayDecodingNumber)
		ss = EncodeParam(ss)
		quotafile := filepath.Join(dirRun, "master")
		ioutil.WriteFile(quotafile, []byte(ss), 0600)
		overquota = false
	}
	return
}

func ReadQuotaConsumed() {
	quotafile := filepath.Join(dirRun, "master")
	if IsExists(quotafile) {
		b, e := ioutil.ReadFile(quotafile)
		if e == nil {
			str := DecodeParam(string(b))
			ss := strings.Split(str, "=")
			if len(ss) == 2 {
				TodayDecodingNumber = Str2int(ss[1])
			}
		}
	}
	today := Str10Unixtime(time.Now().Format("2006-01-02"))
	_, trace := InstanceRetrieveEx("workday", "code", today, "trace")
	if len(trace) > 0 {
		n := Str2int(DecodeParam(trace))
		if n > TodayDecodingNumber {
			TodayDecodingNumber = n
		}
	}
}

func WriteQuotaConsume2DB() {
	today := Str10Unixtime(time.Now().Format("2006-01-02"))
	trace := EncodeParam(strconv.Itoa(TodayDecodingNumber))
	InstancePropertyUpdateEx("workday", "code", today, "trace", trace)
}
