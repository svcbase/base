package base

import (
	"strconv"
	"strings"
)

type CurrencyT struct {
	Currency_id               int
	Code, Title               string
	Exchange_rate             float64
	Symbol_left, Symbol_right string
}

var MapCurrency map[string]CurrencyT

func init() {
	MapCurrency = make(map[string]CurrencyT)
}

func ExchangeRate(currency string) (rate float64) {
	rate = 1
	if c, ok := MapCurrency[currency]; ok {
		rate = c.Exchange_rate
	}
	return
}

func Currency_id(currency string) string {
	id := ""
	if v, ok := MapCurrency[currency]; ok {
		id = strconv.Itoa(v.Currency_id)
	}
	return id
}

func Currency_code(id string) (code string) {
	if len(id) > 0 {
		i_d := Str2int(id)
		for k, v := range MapCurrency {
			if v.Currency_id == i_d {
				code = k
				break
			}
		}
	}
	return
}

func BaseCurrency_id() (id string) {
	curr := GetConfigurationSimple("SYS_BASE_CURRENCY")
	if strings.Contains(curr, ":") {
		id, _ = SplitKV(curr, ":")
	} else {
		i_d := RightInt(curr)
		if i_d > 0 {
			id = strconv.Itoa(i_d)
		}
	}
	return
}

func BaseCurrency_code() (code string) {
	code = Currency_code(BaseCurrency_id())
	return
}

func QuotationMethod(code string) (method string) {
	switch code {
	case "MYR", "RUB", "ZAR", "KRW", "AED", "SAR", "HUF", "PLN", "DKK", "SEK", "NOK", "TRY", "MXN", "THB":
		method = "100人民币折合"
	default:
		method = "100外币折合"
	}
	return
}
