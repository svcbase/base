package base

import (
	"bytes"
	"fmt"
	"strings"
)

func BetweenHeadTail(b []byte, head, tail string) (dt string) {
	i := bytes.Index(b, []byte(head))
	if i >= 0 {
		i += len(head)
		bb := b[i:]
		i = bytes.Index(bb, []byte(tail))
		if i > 0 {
			dt = string(bb[:i])
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
