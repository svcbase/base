package base

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"

	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Strval(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func RightInt(ss string) int {
	v := ""
	n := len(ss)
	for i := 0; i < n; i++ {
		vv := ss[i]
		if IsDigit(vv) {
			v += string(vv)
		} else {
			break
		}
	}
	return Str2int(v)
}

func Int2str(i int) (ss string) {
	ss = strconv.Itoa(i)
	return
}

func Int64str(i int64) (ss string) {
	ss = strconv.FormatInt(i, 10)
	return
}

func Str2int(ss string) int {
	tt := TrimBLANK(ss)
	var r int = 0
	if IsDigital(tt) {
		i, err := strconv.Atoi(tt)
		if err == nil {
			r = i
		}
	}
	return r
}

func Str2int64(ss string) int64 {
	tt := TrimBLANK(ss)
	var i64 int64 = 0
	if IsDigital(tt) {
		i, err := strconv.ParseInt(tt, 10, 64)
		if err == nil {
			i64 = i
		}
	}
	return i64
}

func Str2float64(ss string) (ff float64) {
	tt := TrimBLANK(ss)
	ff = 0.0
	if IsNumber(tt) {
		f64, e := strconv.ParseFloat(tt, 64)
		if e == nil {
			ff = f64
		}
	}
	return
}

func QuoteSegment(str, separator, quote string) string {
	ss := strings.Split(str, separator)
	for i, v := range ss {
		ss[i] = quote + v + quote
	}
	return strings.Join(ss, separator)
}

// ToXmlString convert the map[string]string to xml string
func ToXmlString(param map[string]string) string {
	xml := "<xml>"
	for k, v := range param {
		xml = xml + fmt.Sprintf("<%s>%s</%s>", k, v, k)
	}
	xml = xml + "</xml>"

	return xml
}

func ToJsonString(param map[string]string) (txt string) {
	ss := []string{}
	for k, v := range param {
		ss = append(ss, `"`+k+`":"`+StringJSONEscape(v)+`"`)
	}
	txt = "{" + strings.Join(ss, ",") + "}"
	return
}
func IntMapToJson(param map[int]int) (txt string) {
	ss := []string{}
	for k, v := range param {
		ss = append(ss, `"`+strconv.Itoa(k)+`":`+strconv.Itoa(v))
	}
	txt = "{" + strings.Join(ss, ",") + "}"
	return
}

// ToMap convert the xml struct to map[string]string
func ToMap(in interface{}) (map[string]string, error) {
	out := make(map[string]string)

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts structs; got %T", v)
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv := fi.Tag.Get("xml"); tagv != "" && tagv != "xml" {
			// set key of map to value in struct field
			out[tagv] = v.Field(i).String()
		}
	}
	return out, nil
}

func String2Map(str string) (mapResult map[string]string) { //"0:halt/1:normal/2:hang-up"
	return String2MapOper(str, "/", ":")
}

func String2MapOper(str, oper1, oper2 string) (mapResult map[string]string) { //"0:halt/1:normal/2:hang-up"
	mapResult = make(map[string]string)
	ss := strings.Split(str, oper1)
	for _, v := range ss {
		vv := strings.Split(v, oper2)
		if len(vv) == 2 {
			mapResult[vv[0]] = vv[1]
		}
	}
	return
}

func ReverseStr(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

var C2A0 []byte = []byte{'\xC2', '\xA0'}

func ReplaceNBSP(ss string) string { //&nbsp; -> c2a0 -> space
	return strings.ReplaceAll(ss, string(C2A0), " ")
}

var BOM []byte = []byte{'\xEF', '\xBB', '\xBF'}

//var EmSpace []byte = []byte{'\xE2','\x80','\x83'}

var NEWPAGE byte = '\f'

func ContainsNEWPAGE(ss string) (flag bool) {
	flag = strings.Contains(ss, string(NEWPAGE))
	return
}
func ReplaceNEWPAGE(ss string) string {
	return strings.ReplaceAll(ss, string(NEWPAGE), "")
}

var NEWLINE byte = '\n'

func ContainsNEWLINE(ss string) (flag bool) {
	flag = strings.Contains(ss, string(NEWLINE))
	return
}
func ReplaceNEWLINE(ss string) string {
	return strings.ReplaceAll(ss, string(NEWLINE), "")
}

var RETURN byte = '\r'

func ContainsRETURN(ss string) (flag bool) {
	flag = strings.Contains(ss, string(RETURN))
	return
}
func ReplaceRETURN(ss string) string {
	return strings.ReplaceAll(ss, string(RETURN), "")
}

// \t 制表符
var TAB rune = '\t'

func ContainsTAB(ss string) (flag bool) {
	flag = strings.Contains(ss, string(TAB))
	return
}
func ReplaceTAB(ss string) string {
	return strings.ReplaceAll(ss, string(TAB), "")
}

const aOnlyTable = "\u0000\u0001\u0002\u0003\u0004" + // NUL, SOH, STX, ETX, EOT
	"\u0005\u0006\u0007\u0008\u0009" + // ENQ, ACK, BEL, BS,  HT
	"\u000A\u000B\u000C\u000D\u000E" + // LF,  VT,  FF,  CR,  SO
	"\u000F\u0010\u0011\u0012\u0013" + // SI,  DLE, DC1, DC2, DC3
	"\u0014\u0015\u0016\u0017\u0018" + // DC4, NAK, SYN, ETB, CAN
	"\u0019\u001A\u001B\u001C\u001D" + // EM,  SUB, ESC, FS,  GS
	"\u001E\u001F" // RS,  US

var U0002 rune = '\u0002'

func ContainsU0002(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U0002))
	return
}
func ReplaceU0002(ss string) string {
	return strings.ReplaceAll(ss, string(U0002), "")
}

var U0008 rune = '\u0008' //Backspace
func ContainsU0008(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U0008))
	return
}
func ReplaceU0008(ss string) string {
	return strings.ReplaceAll(ss, string(U0008), "")
}

var U000a rune = '\u000a' //换行符

func ContainsU000a(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U000a))
	return
}
func ReplaceU000a(ss string) string {
	return strings.ReplaceAll(ss, string(U000a), "")
}

var VTAB rune = '\u000b'

func ContainsVTAB(ss string) (flag bool) {
	flag = strings.Contains(ss, string(VTAB))
	return
}
func ReplaceVTAB(ss string) string {
	return strings.ReplaceAll(ss, string(VTAB), "")
}

var U000c rune = '\u000c' //换页符

func ContainsU000c(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U000c))
	return
}
func ReplaceU000c(ss string) string {
	return strings.ReplaceAll(ss, string(U000c), "")
}

var U000d rune = '\u000d' //回车符

func ContainsU000d(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U000d))
	return
}
func ReplaceU000d(ss string) string {
	return strings.ReplaceAll(ss, string(U000d), "")
}

var U00a0 rune = '\u00a0' //不换行符

func ContainsU00a0(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U00a0))
	return
}
func ReplaceU00a0(ss string) string {
	return strings.ReplaceAll(ss, string(U00a0), "")
}

var BLANK001C = '\u001c'

func ContainsBLANK001C(ss string) (flag bool) {
	flag = strings.Contains(ss, string(BLANK001C))
	return
}
func ReplaceBLANK001C(ss string) string {
	return strings.ReplaceAll(ss, string(BLANK001C), "")
}

var BLANK001D = '\u001d'

func ContainsBLANK001D(ss string) (flag bool) {
	flag = strings.Contains(ss, string(BLANK001D))
	return
}
func ReplaceBLANK001D(ss string) string {
	return strings.ReplaceAll(ss, string(BLANK001D), "")
}

var BLANK001E = '\u001e'

func ContainsBLANK001E(ss string) (flag bool) {
	flag = strings.Contains(ss, string(BLANK001E))
	return
}
func ReplaceBLANK001E(ss string) string {
	return strings.ReplaceAll(ss, string(BLANK001E), "")
}

var BLANK001F = '\u001f'

func ContainsBLANK001F(ss string) (flag bool) {
	flag = strings.Contains(ss, string(BLANK001F))
	return
}
func ReplaceBLANK001F(ss string) string {
	return strings.ReplaceAll(ss, string(BLANK001F), "")
}

var U2006 rune = '\u2006' //空格符

func ContainsU2006(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U2006))
	return
}
func ReplaceU2006(ss string) string {
	return strings.ReplaceAll(ss, string(U2006), "")
}

var U200d rune = '\u200d' //零宽连接符

func ContainsU200d(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U200d))
	return
}
func ReplaceU200d(ss string) string {
	return strings.ReplaceAll(ss, string(U200d), "")
}

var U200e rune = '\u200e' //从左至右书写标记

func ContainsU200e(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U200e))
	return
}
func ReplaceU200e(ss string) string {
	return strings.ReplaceAll(ss, string(U200e), "")
}

var U200f rune = '\u200f' //从右至左书写标记

func ContainsU200f(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U200f))
	return
}
func ReplaceU200f(ss string) string {
	return strings.ReplaceAll(ss, string(U200f), "")
}

var U2028 rune = '\u2028' //行分隔符

func ContainsU2028(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U2028))
	return
}
func ReplaceU2028(ss string) string {
	return strings.ReplaceAll(ss, string(U2028), "")
}

var U2029 rune = '\u2029' //行分隔符

func ContainsU2029(ss string) (flag bool) {
	flag = strings.Contains(ss, string(U2029))
	return
}
func ReplaceU2029(ss string) string {
	return strings.ReplaceAll(ss, string(U2029), "")
}

var Ufeff rune = '\ufeff' //字节顺序标记

func ContainsUfeff(ss string) (flag bool) {
	flag = strings.Contains(ss, string(Ufeff))
	return
}
func ReplaceUfeff(ss string) string {
	return strings.ReplaceAll(ss, string(Ufeff), "")
}

func TrimBLANK(ss string) string {
	return strings.Trim(ss, BLANK_CUTSET)
}

func TrimBOM(ss string) string {
	return strings.Replace(ss, string(BOM), "", -1)
}

func TrimPAD(code, pad_formats string) (tcode string) {
	tcode = code
	if len(pad_formats) > 0 {
		ff := strings.Split(pad_formats, ",")
		for _, f := range ff {
			tcode = strings.TrimSuffix(tcode, f)
		}
	}
	return
}

func TrimSquareBrackets(ss string) string {
	b := []byte(ss)
	n := len(b)
	if n > 1 && strings.HasPrefix(ss, "[") && strings.HasSuffix(ss, "]") {
		b = b[1 : n-1]
	}
	return string(b)
}

func TrimDoubleQuote(ss string) string {
	b := []byte(ss)
	n := len(b)
	if n > 1 && strings.HasPrefix(ss, `"`) && strings.HasSuffix(ss, `"`) {
		b = b[1 : n-1]
	}
	return string(b)
}

var ParticularSpaceRUNE []rune = []rune{'\u3000', '\u0002', '\u0007', '\u0008' /*Backspace*/, '\u000b',
	'\u000c' /*换页符*/, '\u00a0' /*不换行符*/, '\u001b', '\u001c', '\u001d', '\u001e', '\u001f', '\u2006', /*空格符*/
	'\u200d' /*零宽连接符*/, '\u200e' /*从左至右书写标记*/, '\u200f' /*从右至左书写标记*/, '\u2028' /*行分隔符*/, '\u2029', /*行分隔符*/
	'\uc2a0' /*空白字符*/, '\ufeff' /*字节顺序标记*/}

func runeEscape(r rune) string {
	return "\\u" + fmt.Sprintf("%04x", int(r))
}

//\t	制表符 ('\u0009')
//\n	新行（换行）符 ('\u000a')
//\r	回车符 ('\u000d')
//\f	换页符 ('\u000c')
//\a	报警 (bell) 符 ('\u0007')
//\e	转义符 ('\u001b')

func IsParticularSpace(r rune) (flag bool) {
	flag = false
	m := len(ParticularSpaceRUNE)
	for j := 0; j < m; j++ {
		if ParticularSpaceRUNE[j] == r {
			flag = true
			break
		}
	}
	return flag
}

func ReplaceParticularSpace(ss string) (rr string) {
	rr = ""
	r := []rune(ss)
	n := len(r)
	for i := 0; i < n; i++ {
		switch r[i] {
		case '\u0009':
			rr += "\\t"
		case '\u000a':
			rr += "\\n"
		case '\u000d':
			rr += "\\r"
		default:
			if IsParticularSpace(r[i]) {
				rr += " "
			} else {
				rr += string(r[i])
			}
		}
	}
	return
}

func TrimParticularSpace(ss string) (rr string) {
	rr = ""
	r := []rune(ss)
	n := len(r)
	for i := 0; i < n; i++ {
		switch r[i] {
		case '\u0009', '\u000a', '\u000d':
		default:
			if !IsParticularSpace(r[i]) {
				rr += string(r[i])
			}
		}
	}
	/*for _, s := range ParticularSpaceRUNE {
		tt := runeEscape(s)
		rr = strings.ReplaceAll(rr, tt, "")
	}
	rr = strings.ReplaceAll(rr, "\\f", "")
	rr = strings.ReplaceAll(rr, "\\a", "")
	rr = strings.ReplaceAll(rr, "\\e", "")*/
	return
}

var IdeographicSpace []byte = []byte{'\xE3', '\x80', '\x80'} //227,128,128  \u3000
var IdeographicSpaceRUNE rune = '\u3000'
var IdeographicSpaceES string = "&#12288;" //Escape Sequence

func TrimIdeographicSpace(ss string) (rr string) {
	rr = ss
	r := []rune(ss)
	n := len(r)
	if n > 0 {
		i, j := 0, n
		for i = 0; i < n; i++ {
			if r[i] != IdeographicSpaceRUNE {
				break
			}
		}
		for j = n; j > i; j-- {
			if r[j-1] != IdeographicSpaceRUNE {
				break
			}
		}
		rr = string(r[i:j])
	}
	return
}

func ReplaceIdeographicSpace(ss string) string {
	return strings.Replace(ss, string(IdeographicSpace), "　", -1)
}

/*func ReplaceEmSpace( ss string ) string {
	return strings.Replace( ss,string(EmSpace),"　",-1 )
}*/

func TrimCell(str string) (ss string) {
	ss = strings.Trim(str, " \r\n\t")
	bd := []byte(ss)
	n := len(bd)
	if n > 1 {
		var sq byte = '\''
		var dq byte = '"'
		if bd[0] == sq && bd[n-1] == sq {
			ss = string(bd[1 : n-1])
		} else if bd[0] == dq && bd[n-1] == dq {
			ss = string(bd[1 : n-1])
		}
	}
	return
}

func Backquote(str string) (ss string) {
	if strings.HasPrefix(str, "`") && strings.HasSuffix(str, "`") {
		ss = str
	} else {
		ss = "`" + str + "`"
	}
	return
}

func AddSlashes(str string) string {
	size := len(str)

	var bb *bytes.Buffer = bytes.NewBuffer(make([]byte, 0, size*2))
	for i := 0; i < size; i++ {
		ch := str[i]
		switch ch {
		case '\'':
			bb.WriteString("\\'")
		case '"':
			bb.WriteString("\\\"")
		case '\\':
			bb.WriteString("\\\\")
		case '\r':
			bb.WriteString("\\r")
		case '\n':
			bb.WriteString("\\n")
		default:
			bb.WriteByte(ch)
		}
	}
	return bb.String()
}

func StringJSONEscape(ss string) (rr string) { //www.json.org escape " \ / b f n r t u0000
	pChar := []rune(ss)
	n := len(pChar)
	rr = ""
	for i := 0; i < n; i++ {
		ch := pChar[i]
		switch ch {
		/*		case '/':
					rr += "\\/"
				case '\'':
					rr += "\\'"*/
		case '"':
			rr += `\"`
		case '\\':
			rr += "\\\\"
		case '\r':
			rr += "\\r"
		case '\n':
			rr += "\\n"
		case '\t':
			rr += "\\t"
		default:
			rr += string(ch)
		}
	}
	return
}

func RectifyCurrency(num string) (ss string) {
	if strings.Contains(num, ".") {
		b := []rune(num)
		n := len(b)
		k := n - 1
		for i := k; i >= 0; i-- {
			ch := b[i]
			if (ch == '1') || (ch == '2') || (ch == '3') || (ch == '4') || (ch == '5') ||
				(ch == '6') || (ch == '7') || (ch == '8') || (ch == '9') || (ch == '.') {
				k = i
				if ch == '.' {
					k--
				}
				break
			} else {
				sc := string(ch)
				if sc == "万" || sc == "元" {
					ss = sc + ss
				}
			}
		}
		for i := k; i >= 0; i-- {
			ss = string(b[i]) + ss
		}
	}
	return
}

/*"1,256,850.00 元 *\n* 1.3245000  万元 (\n)  456.00 （万元）"*/
func RectifyCurrencyInText(ss string) (tt string) {
	tt = ss
	r, _ := regexp.Compile("[0-9]+\\.[0-9]+\\s*[（|\\(]*万*元[\\)|）]*")
	rr := r.FindAllString(ss, -1)
	for _, v := range rr {
		vv := RectifyCurrency(v)
		if vv != v {
			tt = strings.ReplaceAll(tt, v, vv)
		}
	}
	return
}

func RectifyQuantity(num string) (ss string) {
	if strings.Contains(num, ".") {
		b := []rune(num)
		n := len(b)
		k := n - 1
		for i := k; i >= 0; i-- {
			ch := b[i]
			if (ch == '1') || (ch == '2') || (ch == '3') || (ch == '4') || (ch == '5') ||
				(ch == '6') || (ch == '7') || (ch == '8') || (ch == '9') || (ch == '.') {
				k = i
				if ch == '.' {
					k--
				}
				break
			}
		}
		for i := k; i >= 0; i-- {
			ss = string(b[i]) + ss
		}
	}
	return
}

/*"1,256,850.00  *\n* 1.3245000  (\n)  456.00 "*/
func RectifyQuantityInText(ss string) (tt string) {
	tt = ss
	r, _ := regexp.Compile("[0-9]+\\.[0-9]+0")
	rr := r.FindAllString(ss, -1)
	for _, v := range rr {
		vv := RectifyQuantity(v)
		if vv != v {
			tt = strings.ReplaceAll(tt, v, vv)
		}
	}
	return
}

func RectifyWenHao(ss string) (tt string) {
	tt = strings.ReplaceAll(ss, " ", "")
	tt = strings.ReplaceAll(tt, "[", "〔")
	tt = strings.ReplaceAll(tt, "]", "〕")
	tt = strings.ReplaceAll(tt, "【", "〔")
	tt = strings.ReplaceAll(tt, "】", "〕")
	return
}

// "a[2019] 123号  tty【2013】345 号"
func RectifyWHInText(ss string) (tt string) {
	tt = ss
	r, _ := regexp.Compile("[【|\\[]*\\s*[0-9]+\\s*[\\]|】]*\\s*[0-9]+\\s*号")
	rr := r.FindAllString(ss, -1)
	for _, v := range rr {
		vv := RectifyWenHao(v)
		if vv != v {
			tt = strings.ReplaceAll(tt, v, vv)
		}
	}
	return
}

// time converion
func Millisecond2time(millisecond int64) time.Time {
	sec := millisecond / 1000
	nsec := (millisecond - sec*1000) * 1000000
	return time.Unix(sec, nsec)
}

func Millisecond2str19(millisecond int64) string {
	tm := Millisecond2time(millisecond)
	return tm.Format("2006-01-02 15:04:05")
}

func Time2Millisecondstr(tm time.Time) string {
	return strconv.FormatInt(tm.UnixNano()/1e6, 10)
}

func TrimDate(tm string) (ss string) {
	ss = tm
	if tm == "0000-00-00" {
		ss = ""
	}
	return
}

func ZeroDate(tm string) (flag bool) {
	flag = (tm == "0000-00-00")
	return
}

func NullTime(v string) bool {
	return (v == "0000-00-00 00:00:00" || v == ZERO_TIME)
}

func TrimZero(t string) (ss string) {
	ss = t
	if t == "0" {
		ss = ""
	}
	return
}

func IsTime19(ss string) (flag bool) {
	flag = false
	pattern := "^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}$"
	flag, _ = regexp.MatchString(pattern, ss)
	return
}

func IsZero(tm time.Time) (flag bool) {
	flag = (tm.Year() == 0 && tm.Month() == 1 && tm.Day() == 1 &&
		tm.Hour() == 0 && tm.Minute() == 0 && tm.Second() == 0)
	//t := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	//flag = tm.Equal(t)
	//fmt.Println(flag, tm, t)
	return
}

func WhetherExpired(expiredate10 string) (flag bool) {
	flag = true
	if len(expiredate10) > 0 {
		tm, e := Str2time(expiredate10)
		if e == nil {
			tm.AddDate(0, 0, 1)
			flag = time.Now().After(tm)
		}
	}
	return
}

func Strtime2unix(v string) (unixtime int64) {
	tm, e := Str2time(v)
	if e == nil {
		unixtime = tm.Unix()
	}
	return
}

func Str2time(v string) (tm time.Time, err error) {
	err = errors.New("time format error")
	n := len(v)
	if n == 29 { //2024-06-30T10:47:28.000+08:00
		i := strings.Index(v, ".")
		if i == 19 {
			b := []byte(v)
			ss := string(b[:i])
			ss = strings.ReplaceAll(ss, "T", " ")
			tm, err = Str19time(ss)
		}
	} else if n == 20 { //0000-01-01T00:00:00Z
		if v == "0000-01-01T00:00:00Z" {
			tm, err = Str19time(ZERO_TIME)
		} else {
			tm, err = Str20time(v)
		}
	} else if n == 19 {
		tm, err = Str19time(v)
	} else if n == 10 {
		tm, err = Str10time(v)
	} else if n == 0 {
		tm, err = Str19time(ZERO_TIME)
	}
	return
}

func Str19time(v string) (tm time.Time, err error) { //v := "2016-03-02 12:59:59"
	loc, _ := time.LoadLocation("Local")
	tm, err = time.ParseInLocation("2006-1-2 15:4:5", v, loc)
	return
}

func Str10time(v string) (tm time.Time, err error) { //v := "2016-03-02"
	loc, _ := time.LoadLocation("UTC")
	tm, err = time.ParseInLocation("2006-1-2", v, loc)
	return
}

func UnixtimeStr10(unixtime string) (tm string) {
	ut := Str2int64(unixtime)
	tm = time.Unix(ut, 0).Format("2006-01-02")
	return
}

func UnixtimePreviousday10(unixtime string) (tm string) {
	ut := Str2int64(unixtime)
	previousday := time.Unix(ut, 0).AddDate(0, 0, -1)
	tm = previousday.Format("2006-01-02")
	return
}

func Str10Unixtime(v string) (unixtm string) {
	if len(v) == 10 {
		tm, e := time.Parse("2006-01-02", v)
		if e == nil {
			unixtm = strconv.FormatInt(tm.Unix(), 10)
		}
	}
	return
}

func Today() (ss string) {
	tm := time.Now()
	ss = tm.Format("2006-01-02")
	return
}

func Yesterday() (ss string) {
	tm := time.Now().AddDate(0, 0, -1)
	ss = tm.Format("2006-01-02")
	return
}

func TodayUnixtime() (unixtm string) {
	year, month, day := time.Now().Date()
	tm := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	unixtm = strconv.FormatInt(tm.Unix(), 10)
	return
}

func Str19millisecond(v string) string {
	ms := ""
	tm, err := Str19time(v)
	if err == nil {
		ms = Time2Millisecondstr(tm)
	}
	return ms
}

func Str20time(v string) (tm time.Time, err error) { //0000-01-01T00:00:00Z
	if len(v) == 20 {
		/*    ANSIC       = "Mon Jan _2 15:04:05 2006"
		      UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
		      RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
		      RFC822      = "02 Jan 06 15:04 MST"
		      RFC822Z     = "02 Jan 06 15:04 -0700" // 使用数字表示时区的RFC822
		      RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
		      RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
		      RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // 使用数字表示时区的RFC1123
		      RFC3339     = "2006-01-02T15:04:05Z07:00"
		      RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
		      Kitchen     = "3:04PM"
		      Stamp      = "Jan _2 15:04:05"
		      StampMilli = "Jan _2 15:04:05.000"
		      StampMicro = "Jan _2 15:04:05.000000"
		      StampNano  = "Jan _2 15:04:05.000000000"*/
		loc, _ := time.LoadLocation("Local")
		tm, err = time.ParseInLocation("2006-01-02T15:04:05Z", v, loc)
	}
	return
}

/*
yyyy=2006,
mmmm=January,
MMM=Jan
mm=01
dd=02
m=1
d=2
HH=15
hh=03
nn=04
ss=05
yyyy-mm-dd,2006-01-02
yyyy/mm/dd,2006/01/02
dd-mm-yyyy,02-01-2006
dd/mm/yyyy,02/01/2006
dd mmmm yyyy,02 January 2006
dd.mm.yyyy,02.01.2006
d.m.yyyy,2.1.2006
mm-dd-yyyy,01-02-2006
mm/dd/yyyy,01/02/2006
`MMM dd, yyyy`,`Jan 02, 2006`
`mmmm dd, yyyy`,`January 02, 2006`
----------------------------------
HH:nn:ss,15:04:05
HH:nn,15:04
hh:nn:ss am/pm,03:04:05 pm
hh:nn am/pm,03:04 pm
*/
func DateTimeLayout(ymd_hns string) (format string) { //ymd_hns: yyyy-mm-dd HH:nn:ss
	format = ymd_hns
	layout := []string{"yyyy=2006", "mmmm=January", "MMM=Jan", "mm=01", "dd=02",
		"m=1", "d=2", "HH=15", "hh=03", "nn=04", "ss=05", "am/pm=pm", "AM/PM=PM"}
	n := len(layout)
	for i := 0; i < n; i++ {
		lolo := strings.Split(layout[i], "=")
		if len(lolo) == 2 {
			format = strings.ReplaceAll(format, lolo[0], lolo[1])
		}
	}
	return
}

/*
1分钟之内<60秒	刚刚		just now
1小时之内<3600秒	x分钟前	x minutes ago
今天24点之前	今天23:59	23:59 today
昨天			昨天00:00	00:00 yesterday
今年			mm-dd 		按账户偏好展现（省略年份）
去年及以前	yyyy-mm-dd 	按账户偏好展现
TIP_JUST	=	刚刚
TIP_XM_AGO	=	%d 几分钟前
TIP_HN_TODAY	=	今天 HH:nn
TIP_HN_YESTERDAY	=	昨天 HH:nn
*/
func HumanizedTime(tm time.Time, dateformat, timeformat, clientlanguage_id string) (dt string) {
	if !tm.IsZero() && !IsZero(tm) {
		now := time.Now()
		if tm.After(now) {
			dt = tm.Format(DateTimeLayout(dateformat + " " + timeformat))
		} else {
			dm, _ := time.ParseDuration("-1m")
			if tm.After(now.Add(dm)) {
				dt = GetConfigurationLanguage("TIP_JUST", clientlanguage_id)
			} else {
				dh, _ := time.ParseDuration("-1h")
				if tm.After(now.Add(dh)) {
					xma := GetConfigurationLanguage("TIP_XM_AGO", clientlanguage_id)
					dt = fmt.Sprintf(xma, int(now.Sub(tm).Minutes()))
				} else {
					nyd := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
					if tm.After(nyd) {
						md := strings.Trim(strings.ReplaceAll(dateformat, "2006", ""), "/- .")
						dt = tm.Format(DateTimeLayout(md)) //extract year
					} else {
						dt = tm.Format(DateTimeLayout(dateformat))
					}
				}
			}
		}
	}
	return
}

func StrMD5(ss string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(ss))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

func FileMD5(filename string) (md5str string) {
	file, err := os.Open(filename)
	if err == nil {
		hMD5 := md5.New()
		_, err = io.Copy(hMD5, file)
		if err == nil {
			md5str = hex.EncodeToString(hMD5.Sum(nil))
		}
		file.Close()
	}
	return
}

func Now2MD5() string {
	return StrMD5(strconv.FormatInt(time.Now().UnixNano(), 10))
}

func NowSeedMD5(seed string) string {
	return StrMD5(seed + strconv.FormatInt(time.Now().UnixNano(), 10))
}

func ShowPrice(currency_code string, cny_price int) string {
	s_left, s_right := "$", ""
	var price float64 = float64(cny_price)
	if c, ok := MapCurrency[currency_code]; ok {
		price = float64(cny_price) * c.Exchange_rate
		s_left = c.Symbol_left
		s_right = c.Symbol_right
	}
	return s_left + fmt.Sprintf("%.2f", price) + s_right
}

func ExPrice(currency_code string, cny_price float64) (s_left, s_right string, price float64) {
	s_left = "$"
	s_right = ""
	price = cny_price
	if c, ok := MapCurrency[currency_code]; ok {
		price = price * c.Exchange_rate
		s_left = c.Symbol_left
		s_right = c.Symbol_right
	}
	price, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", price), 64)
	return
}

func Round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

func ExCurrency(amount float64, from_currency, to_currency string) (amt float64) {
	amt = amount
	if from_currency != to_currency { //x USD = 10 EUR/0.13688129*0.15408321
		var f_curr, t_curr float64 = 1.0, 1.0
		if c, ok := MapCurrency[from_currency]; ok {
			if c.Exchange_rate > 0 {
				f_curr = c.Exchange_rate
			}
		}
		if c, ok := MapCurrency[to_currency]; ok {
			if c.Exchange_rate > 0 {
				t_curr = c.Exchange_rate
			}
		}
		amt = amount / f_curr * t_curr
	}
	return
}

// AbsoluteHyper("www.abc.com/cake/a.php?p=1&g=2","../index.php") "b.php"
func AbsoluteHyper(baseUrl, hyperlink string) string {
	newlink := hyperlink
	bsUrl := baseUrl
	urlprefix := "http"
	if !strings.HasPrefix(bsUrl, urlprefix) {
		bsUrl = urlprefix + "://" + bsUrl
	}
	aurl, err := url.Parse(bsUrl)
	if err == nil {
		burl, err := url.Parse(hyperlink)
		if err == nil {
			if !burl.IsAbs() {
				burl = aurl.ResolveReference(burl)
				newlink = burl.String()
			}
		}
	}
	return newlink
}

const (
	base64Table   = "0DSAC8EFG7HIBKJ_LMNPQROTU4g6XWY9VZbd.aceh1fiklmnt5opqrjsvwxuy2z3"
	IDbase64Table = "0DSAC8EFG7HIBKJ~LMNPQROTU4g6XWY9VZbd.aceh1fiklmnt5opqrjsvwxuy2z3"
)

//ID	 use alphabet letter,digit or -, begin with a letter and not accept _ , * , / , \

func EncodeID(id int64) (sr string) {
	sr = "id" + encodeParamAppointTable(strconv.FormatInt(id, 10), IDbase64Table)
	return
}

func DecodeID(src string) (id int64) {
	ss := src
	if strings.HasPrefix(src, "id") {
		ss = src[2:]
	}
	id = Str2int64(decodeParamAppointTable(ss, IDbase64Table))
	return
}

func DecodeParam(src string) (sr string) {
	return decodeParamAppointTable(src, base64Table)
}

func EncodeParam(src string) (sr string) {
	return encodeParamAppointTable(src, base64Table)
}

func decodeParamAppointTable(src, table string) (sr string) {
	coder := base64.NewEncoding(table)
	des, e := coder.DecodeString(strings.Replace(src, "-", "=", -1))
	if e == nil {
		sr = string(des)
	} else {
		//log.Println("coder.DecodeString", e)
	}
	return
}

func encodeParamAppointTable(src, table string) (sr string) {
	coder := base64.NewEncoding(table)
	sr = strings.Replace(coder.EncodeToString([]byte(src)), "=", "-", -1)
	return
}

func keyTable(keys []int) (tbl string) {
	var codeTable = []byte(base64Table)
	var n int = len(keys) / 2
	for i := 0; i < n; i++ {
		a := keys[2*i]
		b := keys[2*i+1]
		ch := codeTable[a]
		codeTable[a] = codeTable[b]
		codeTable[b] = ch
	}
	return string(codeTable)
}

func DecodeByKey(src string, keys []int) (sr string) { //keys=[10,32,22,14,04,45]
	coder := base64.NewEncoding(keyTable(keys))
	des, e := coder.DecodeString(strings.Replace(src, "-", "=", -1))
	if e == nil {
		sr = string(des)
	} else {
		//log.Println("coder.DecodeString", e)
	}
	return
}

func EncodeByKey(src string, keys []int) (sr string) {
	coder := base64.NewEncoding(keyTable(keys))
	sr = strings.Replace(coder.EncodeToString([]byte(src)), "=", "-", -1)
	return
}

var keyss = [][]int{
	{0, 32, 22, 14, 4, 45},
	{1, 13, 24, 34, 6, 39},
	{2, 10, 11, 22, 9, 23},
	{3, 22, 15, 10, 24, 9},
	{4, 63, 7, 22, 14, 15},
	{5, 52, 12, 44, 18, 53},
	{6, 43, 33, 11, 21, 42},
	{7, 35, 54, 18, 38, 25},
	{8, 28, 61, 22, 58, 25},
	{9, 13, 47, 14, 63, 2}}

func Encode(scheme, str string) (ss string) {
	ss = str
	switch scheme {
	case "base64":
		if len(str) > 0 {
			ss = base64.StdEncoding.EncodeToString([]byte(str))
		}
	case "salt-b":
		v := fmt.Sprintf("%d", len(str))
		ending := string(v[len(v)-1])
		ss = ending + EncodeByKey(str, keyss[Str2int(ending)])
	}
	return
}

func Decode(scheme, str string) (ss string) {
	ss = str
	switch scheme {
	case "base64":
		if len(str) > 0 {
			b, e := base64.StdEncoding.DecodeString(str)
			if e == nil {
				ss = string(b)
			}
		}
	case "salt-b":
		n := len(str)
		if n > 0 {
			head := str[0]
			if IsDigit(head) {
				i := Str2int(string(head))
				ss = DecodeByKey(str[1:], keyss[i])
			}
		}
	}
	return
}

func Briefing(message string, max int) (brief string) {
	suffix := ""
	msg := []rune(message)
	m := len(msg)
	if m > max {
		m = max
		suffix = "..."
	}
	brief = string(msg[:m]) + suffix
	return
}

func ByteFormat(bytes int64) (ss string) {
	dictionary := []string{"bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	i := 0
	nn := float32(bytes)
	n := len(dictionary)
	for i = 0; i < n; i++ {
		if nn < 1024 {
			break
		}
		nn = nn / 1024
	}
	ss = fmt.Sprintf("%.2f", nn) + " " + dictionary[i]
	return
}

func ThisUrl_escaped(r *http.Request) (theurl string) {
	/*scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}*/
	protocol := GetConfigurationSimple("SYS_PROTOCOL")
	site_address := GetConfigurationSimple("SITE_ADDRESS")
	theurl = url.QueryEscape(protocol + "://" + site_address + r.RequestURI)
	return
}

func ThisUrl(r *http.Request) (theurl string) {
	/*	scheme := "http://"
		if r.TLS != nil {
			scheme = "https://"
		}*/
	protocol := GetConfigurationSimple("SYS_PROTOCOL")
	site_address := GetConfigurationSimple("SITE_ADDRESS")

	theurl = protocol + "://" + site_address + r.RequestURI
	return
}

func Padleft(num int64, width int, padchar byte) (ss string) {
	ss = fmt.Sprintf("%d", num)
	n := width - len(ss)
	if n > 0 {
		ss = strings.Repeat(string(padchar), n) + ss
	}
	return
}

/*use in meta .table file*/
func CellQuotation(str string) (ss string) {
	ss = str
	if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") ||
		strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) ||
		strings.HasPrefix(str, "`") && strings.HasSuffix(str, "`") {
	} else {
		if strings.Contains(str, "'") || strings.Contains(str, `"`) ||
			strings.Contains(str, ",") {
			ss = "`" + str + "`"
		}
	}
	return
}

// 6.00% --->  0.06
func CalcPercent(str string, precision int) (ss string) {
	if strings.HasSuffix(str, "%") {
		ss = strings.TrimSuffix(str, "%")
		ss = fmt.Sprintf("%."+strconv.Itoa(precision)+"f", Str2float64(ss)/100)
		if strings.Contains(ss, ".") && strings.HasSuffix(ss, "0") {
			b := []byte(ss)
			for i := len(b) - 1; i >= 0; i-- {
				if b[i] != '0' {
					ss = string(b[:i+1])
					break
				}
			}
		}
	} else {
		ss = str
	}
	return
}
