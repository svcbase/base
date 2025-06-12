package base

import (
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

func SplitIDTag(str string) (id int, tag string) { // id:code.name
	id = 0
	dd := strings.SplitN(str, ":", 2)
	n := len(dd)
	if n > 0 {
		if IsDigital(dd[0]) {
			id = Str2int(dd[0])
		}
	}
	if n > 1 {
		tag = dd[1]
	}
	return
}

func SplitIDCodeName(str string) (id int, code, name string) { // id:code.name
	id = 0
	dd := strings.SplitN(str, ":", 2)
	n := len(dd)
	if n > 0 {
		if IsDigital(dd[0]) {
			id = Str2int(dd[0])
		}
	}
	if n > 1 {
		ss := strings.SplitN(dd[1], ".", 2)
		n = len(ss)
		if n > 0 {
			code = ss[0]
		}
		if n > 1 {
			name = ss[1]
		}
	}
	return
}

func SplitCodeName(str string) (code, name string) { // code.name
	ss := strings.SplitN(str, ".", 2)
	n := len(ss)
	if n > 0 {
		code = ss[0]
	}
	if n > 1 {
		name = ss[1]
	}
	return
}

func SplitKV(str, separator string) (k, v string) {
	ss := strings.SplitN(str, separator, 2)
	n := len(ss)
	if n == 1 {
		v = TrimBLANK(ss[0])
	} else if n == 2 {
		k = TrimBLANK(ss[0])
		v = TrimBLANK(ss[1])
	}
	return
}

func SplitK_V(str, separator string) (k, v string) {
	ss := strings.SplitN(str, separator, 2)
	n := len(ss)
	if n == 1 {
		k = TrimBLANK(ss[0])
	} else if n == 2 {
		k = TrimBLANK(ss[0])
		v = TrimBLANK(ss[1])
	}
	return
}

func SplitXYZ(str, separator string) (x, y, z string) {
	ss := strings.SplitN(str, separator, 3)
	n := len(ss)
	if n == 1 {
		x = TrimBLANK(ss[0])
	} else if n == 2 {
		x = TrimBLANK(ss[0])
		y = TrimBLANK(ss[1])
	} else if n == 3 {
		x = TrimBLANK(ss[0])
		y = TrimBLANK(ss[1])
		z = TrimBLANK(ss[2])
	}
	return
}

/*func SplitWH(str, separator string) (w, h string) {
	if len(str) > 0 {
		k, v := SplitK_V(str, separator)
		if len(v) > 0 {
			w = k
			h = v
		} else {
			w = k
			h = k
		}
	}
	return
}*/

func SplitCND(str, separator string) (code, name, desc string) {
	vv := strings.Split(str, separator)
	n := len(vv)
	if n > 0 {
		code = TrimBLANK(vv[0])
	}
	if n > 1 {
		name = TrimBLANK(vv[1])
	}
	if n > 2 {
		desc = TrimBLANK(vv[2])
	}
	return
}
func Splitfield(txt, init_seps, separator string) (sep string, fields []string, err error) {
	separators := ",\t;/|"
	if len(init_seps) > 0 {
		separators = init_seps
	}
	value := ""
	sep = separator
	i, state := 0, 0
	ptxt := []rune(txt)
	n := len(ptxt)
	for i < n {
		char := string(ptxt[i])
		if len(sep) == 0 && strings.Contains(separators, char) {
			sep = char
		}
		switch state {
		case 0:
			switch char {
			case " ", "\r", "\n": /*trim space*/
			case "'":
				state = 2
				value += char
			case `"`:
				state = 3
				value += char
			case "`":
				state = 4
				value += char
			default:
				if char == sep {
					fields = append(fields, value)
					value = ""
				} else {
					state = 1
					value += char
				}
			}
		case 1:
			if char == "\r" || char == "\n" { /*trim*/
			} else if char == sep {
				fields = append(fields, value)
				value = ""
				state = 0
			} else {
				value += char
			}
		case 2:
			if char == "'" {
				v := []rune(value)
				fields = append(fields, string(v[1:]))
				value = ""
				state = 50
			} else if char == "\\" {
				value += char
				state = 10
			} else {
				value += char
			}
		case 3:
			if char == `"` {
				v := []rune(value)
				fields = append(fields, string(v[1:]))
				value = ""
				state = 50
			} else if char == "\\" {
				value += char
				state = 20
			} else {
				value += char
			}
		case 4:
			if char == "`" {
				v := []rune(value)
				fields = append(fields, string(v[1:]))
				value = ""
				state = 50
			} else if char == "\\" {
				value += char
				state = 30
			} else {
				value += char
			}
		case 10:
			value += char
			if char == "'" {
				state = 11
			} else {
				state = 2
			}
		case 11:
			if char == "'" {
				v := []rune(value)
				fields = append(fields, string(v[1:]))
				value = ""
				state = 50
			} else if char == sep {
				v := []rune(value)
				fields = append(fields, string(v[1:len(v)-1]))
				value = ""
				state = 0
			} else {
				value += char
				state = 2
			}
		case 20:
			value += char
			if char == `"` {
				state = 21
			} else {
				state = 3
			}
		case 21:
			if char == `"` {
				v := []rune(value)
				fields = append(fields, string(v[1:]))
				value = ""
				state = 50
			} else if char == sep {
				v := []rune(value)
				fields = append(fields, string(v[1:len(v)-1]))
				state = 0
				value = ""
			} else {
				value += char
				state = 3
			}
		case 30:
			value += char
			if char == "`" {
				state = 31
			} else {
				state = 4
			}
		case 31:
			if char == "`" {
				v := []rune(value)
				fields = append(fields, string(v[1:]))
				value = ""
				state = 50
			} else if char == sep {
				v := []rune(value)
				fields = append(fields, string(v[1:len(v)-1]))
				state = 0
				value = ""
			} else {
				value += char
				state = 4
			}
		case 50:
			if char == sep {
				state = 0
			} else if char == " " || char == "\r" || char == "\n" {

			} else {
				err = errors.New("quote right adhesive character:" + char)
				state = 99
				i = n
			}
		}
		i++
	}
	if state == 2 || state == 10 || state == 3 || state == 20 || state == 4 || state == 30 {
		err = errors.New("unpaired quotation marks:" + value)
	} else if state == 11 || state == 21 || state == 31 {
		v := []rune(value)
		value = string(v[1 : len(v)-1])
	}
	if len(value) > 0 {
		fields = append(fields, value)
	} else {
		if state == 0 {
			fields = append(fields, "")
		}
	}
	return
}
func SplitSQL(line string) (sqltxt, comment string) {
	ss := strings.Split(line, "#*#")
	if len(ss) == 2 {
		sqltxt = ss[0]
		switch DB_type {
		case SQLite:
			df := "DATE_FORMAT(now(),'%b %d %Y')"
			if strings.Contains(sqltxt, df) {
				sqltxt = strings.ReplaceAll(sqltxt, df, "strftime('%m-%d-%Y')")
			}
			df = "DATE_FORMAT(now(),'%Y年%c月%e日')"
			if strings.Contains(sqltxt, df) {
				sqltxt = strings.ReplaceAll(sqltxt, df, "strftime('%Y年%m月%d日')")
			}
			df = "concat(name,' [',code,']')"
			if strings.Contains(sqltxt, df) {
				sqltxt = strings.ReplaceAll(sqltxt, df, "name||' ['||code||']'")
			}
			df = " FROM DUAL"
			if strings.Contains(sqltxt, df) {
				sqltxt = strings.ReplaceAll(sqltxt, df, "")
			}
			if strings.Contains(sqltxt, "now()") {
				sqltxt = strings.ReplaceAll(sqltxt, "now()", "current_timestamp")
			}
			if strings.Contains(sqltxt, "_NOW_") {
				sqltxt = strings.ReplaceAll(sqltxt, "_NOW_", "current_timestamp")
			}
			if strings.Contains(sqltxt, "\\'") {
				sqltxt = strings.ReplaceAll(sqltxt, "\\'", "''")
			}
		case MySQL:
			if strings.Contains(sqltxt, "_NOW_") {
				sqltxt = strings.ReplaceAll(sqltxt, "_NOW_", "now()")
			}
		}
		comment = ss[1]
	}
	return
}

// /cloneinstances?from=productcategory&scene=property&pid={{value}}&to=product&sub=property&rmi={{instance_id}}
func SplitURL(theurl string) (action string, mapKV map[string]string) {
	mapKV = make(map[string]string)
	ss := strings.SplitN(theurl, "?", 2)
	if len(ss) == 2 {
		action = ss[0]
		pp := strings.Split(ss[1], "&")
		for _, kv := range pp {
			k_v := strings.SplitN(kv, "=", 2)
			if len(k_v) == 2 {
				mapKV[k_v[0]] = k_v[1]
			}
		}
	}
	return
}

func SplitParam(param, seg_separator, kv_separator string) (mapKV map[string]string) {
	mapKV = make(map[string]string)
	segments := strings.Split(param, seg_separator) // "|"
	for _, segment := range segments {
		k, v := SplitK_V(segment, kv_separator) // ":"
		mapKV[k] = v
	}
	return
}

func ExtractCNmobile(str string) (mobile string) {
	pattern := "1\\d{10}" //china mobile format
	r, e := regexp.Compile(pattern)
	if e == nil {
		mobilee := r.FindAllString(str, -1)
		if len(mobilee) > 0 {
			fonee := []string{}
			for _, fone := range mobilee {
				exists, _ := In_array(fone, fonee)
				if !exists {
					fonee = append(fonee, fone)
				}
			}
			mobile = strings.Join(fonee, ",")
		}
	}
	return
}

func ExtractCNphone(str string) (phone string) {
	ss := str
	mobilee := strings.Split(ExtractCNmobile(str), ",")
	for _, mobile := range mobilee {
		ss = strings.ReplaceAll(ss, mobile, " ")
	}
	pattern := "(0[1-9]\\d{1,2}-\\d{7,8})|(0[1-9]\\d{1,2}\\d{7,8})|(0[1-9]\\d{1,2}-9\\d{4})|9\\d{4}|400[0-9|\\s|\\-]{7,9}" //china phone format
	pattern0 := "[2-8]\\d{6,7}"
	r, e := regexp.Compile(pattern)
	if e == nil {
		phonee := r.FindAllString(ss, -1)
		if len(phonee) == 0 {
			ss = strings.ReplaceAll(ss, " ", "")
			phonee = r.FindAllString(ss, -1)
		}
		if len(phonee) == 0 {
			dx := []string{"１", "２", "３", "４", "５", "６", "７", "８", "９", "０", "—"}
			xx := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-"}
			for i, n := 0, len(dx); i < n; i++ {
				ss = strings.ReplaceAll(ss, dx[i], xx[i])
			}
			phonee = r.FindAllString(ss, -1)
		}
		fonee := []string{}
		if len(phonee) > 0 {
			for _, fone := range phonee {
				if strings.HasPrefix(fone, "400") {
					fone = strings.ReplaceAll(fone, " ", "")
					fone = strings.ReplaceAll(fone, "-", "")
				}
				exists, _ := In_array(fone, fonee)
				if !exists {
					fonee = append(fonee, fone)
				}
			}
		} else {
			r, e = regexp.Compile(pattern0)
			if e == nil {
				fonee = r.FindAllString(ss, -1)
			}
		}
		if len(fonee) > 0 {
			phone = strings.Join(fonee, ",")
		}
	}
	return
}

func Extractemail(str string) (email string) {
	pattern := "\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*"
	r, e := regexp.Compile(pattern)
	if e == nil {
		emaill := r.FindAllString(str, -1)
		if len(emaill) > 0 {
			email = strings.Join(emaill, ",")
		}
	}
	return
}

func SplitBanknameAccount(str string) (bankname, bankaccount string) {
	pattern := "[0-9]{10,}"
	r, e := regexp.Compile(pattern)
	if e == nil {
		bankaccount = r.FindString(str)
	}
	bankname = TrimBLANK(strings.ReplaceAll(str, bankaccount, ""))
	return
}

func SplitAddressTelephone(str string) (address, telephone string) {
	address = str
	mobile := ExtractCNmobile(str)
	telephone = ExtractCNphone(str)
	if len(telephone) > 0 {
		address = strings.ReplaceAll(address, telephone, "")
	}
	if len(mobile) > 0 {
		address = strings.ReplaceAll(address, mobile, "")
		telephone += " " + mobile
	}
	address = TrimBLANK(address)
	telephone = TrimBLANK(telephone)
	return
}
func SplitAddressTelephoneMobile(str string) (address, telephone, mobile string) {
	address = str
	mobile = ExtractCNmobile(str)
	telephone = ExtractCNphone(str)
	if len(telephone) > 0 {
		address = strings.ReplaceAll(address, telephone, "")
	}
	if len(mobile) > 0 {
		address = strings.ReplaceAll(address, mobile, "")
	}
	address = TrimBLANK(address)
	return
}

func SplitRoadmapid(rmi string) (roadmapids []string) {
	var rmirmi []string
	if strings.Contains(rmi, ",") {
		rmirmi = strings.Split(rmi, ",")
	} else {
		rmirmi = strings.Split(rmi, ".")
	}
	for _, v := range rmirmi {
		if IsDigital(v) {
			roadmapids = append(roadmapids, v)
		}
	}
	return
}

// key:"MP-mpid" val:"Appid,Appsecret"
func SplitV2map(key string, bs64 bool) (mapKV map[string]string) {
	mapKV = make(map[string]string)
	val := GetV(key)
	if len(val) > 0 {
		if bs64 {
			b, e := base64.StdEncoding.DecodeString(val)
			if e == nil {
				val = string(b)
			}
		}
		o := gjson.Parse(val)
		if o.Exists() {
			o.ForEach(func(k, v gjson.Result) bool {
				mapKV[k.String()] = v.String()
				return true
			})
		}
	}
	return
}
