package base

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

func InitSysLang() {
	fname := filepath.Join(dirRes, "res", "language.sys")
	fi, err := os.Open(fname)
	if err == nil {
		i := 1
		eof := false
		r := bufio.NewReader(fi)
		for {
			str, e := r.ReadString('\n')
			if e == io.EOF {
				eof = true
			}
			if len(str) > 0 {
				k, v := SplitKV(TrimBLANK(str), ".")
				ss := strconv.Itoa(i) //1,2,3,...
				mapLang[k] = ss
				mapLangTag[ss] = v + " [" + k + "]"
				mapLangFlag[ss] = ""
				i++
			}
			if eof {
				break
			}
		}
		fi.Close()
	}

	sys_accept_language, _ = ReadLanguageConfigurationFromFile()
	fname = filepath.Join(dirRes, "res", "multilanguage")
	if IsExists(fname) {
		b, e := ioutil.ReadFile(fname)
		if e == nil {
			sys_acceptmultiplelanguage = (TrimBLANK(string(b)) == "1")
		}
	} else {
		if strings.Count(sys_accept_language, ",") > 0 {
			sys_acceptmultiplelanguage = true
		}
	}
}

func SetSysAcceptMultipleLanguage(multilanguage bool) {
	sys_acceptmultiplelanguage = multilanguage
}

func SysAcceptMultipleLanguage() (flag bool) {
	flag = sys_acceptmultiplelanguage
	return
}
func Clientlanguage_id(r *http.Request) (id string) {
	id = Language_id(PreferredLanguage(r.Header["Accept-Language"]))
	return
}

func PreferredLanguage(al []string) (lang string) { //zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3
	lang = Baselanguage_code()
	if len(al) > 0 {
		acl := AcceptLanguageCodes() //[]string
		exists, _ := In_array(lang, acl)
		if !exists {
			acl = append(acl, lang)
		}
		ll := strings.Split(al[0], ",")
		n := len(ll)
		for i := 0; i < n; i++ {
			nl := strings.Split(ll[i], ";")
			if len(nl) > 0 {
				cl := strings.Split(nl[0], "-")
				if len(cl) > 0 {
					code := cl[0]
					if len(code) > 0 {
						exists, _ = In_array(code, acl)
						if exists {
							lang = code
							break
						}
					}
				}
			}
		}
	}
	/*	lang = "zh" //zh-CN,zh;q=0.8
		if len(al) > 0 {
			ss := strings.Split(al[0], ";")
			if len(ss) > 0 {
				if !strings.HasPrefix(ss[0], "zh") {
					lang = "en"
				}
			}
		}*/
	return
}

func GetLanguages() (langs []string) {
	for k, _ := range mapLang {
		langs = append(langs, k)
	}
	return
}

func GetLanguageids() (ids []string) {
	for _, id := range mapLang {
		ids = append(ids, id)
	}
	return
}

func Language_id(lang string) string {
	id := ""
	if i_d, ok := mapLang[lang]; ok {
		id = i_d
	}
	return id
}

func Language_code(id string) (code string) {
	for k, v := range mapLang {
		if v == id {
			code = k
			break
		}
	}
	return
}

func BaseLanguage_id() (id string) {
	lang := ""
	kv := GetConfigurationSimple("SYS_BASE_LANGUAGE")
	if len(kv) > 0 {
		lang = kv
	} else {
		ss := strings.Split(sys_accept_language, ",")
		if len(ss) > 0 {
			lang = ss[0]
		}
	}
	if strings.Contains(lang, ":") {
		id, _ = SplitKV(lang, ":")
	} else {
		id = lang
	}
	return
}

func Baselanguage_code() (code string) {
	code = Language_code(BaseLanguage_id())
	return
}

func Language_flag(id string) (flag string) {
	flag, _ = mapLangFlag[id]
	return
}

func Language_tag(id string) (tag string) {
	tag, _ = mapLangTag[id]
	return
}

func Language_Currency(id string) (curr string) {
	curr, _ = mapLangCurr[id]
	return
}

func Language_label(data, language_id string) (str string) {
	str = GetLanguageVersionText(data, Baselanguage_code(), Language_code(language_id), ";")
	return
}

func LanguageLabel(data, language string) (str string) {
	str = GetLanguageVersionText(data, Baselanguage_code(), language, ";")
	/*en := data
	ss := strings.Split(data, ";")
	for _, v := range ss {
		vv := strings.Split(v, ":")
		if len(vv) == 2 {
			if vv[0] == "en" {
				en = vv[1]
			}
			if vv[0] == language {
				str = vv[1]
			}
		}
	}
	if len(str) == 0 {
		str = en
	}*/
	return
}

type LanguageT struct {
	Language_id   string
	Language_tag  string
	Language_flag string
	Selected      bool
}

type simplelanguageT struct {
	Lang_id  int
	Lang_tag string
	Notfirst int
}

func AllLanguages() (languages []simplelanguageT) {
	ids := []int{}
	for k, _ := range mapLangTag {
		ids = append(ids, Str2int(k))
	}
	sort.Ints(ids)
	n := len(ids)
	for i := 0; i < n; i++ {
		id := ids[i]
		notfirst := 1
		if i == 0 {
			notfirst = 0
		}
		languages = append(languages, simplelanguageT{id, mapLangTag[strconv.Itoa(id)], notfirst})
	}
	return
}

func AcceptLanguages(multiple int) (languages []LanguageT) { //1:English [en],2:Chinese 中文 [zh]
	baselanguage_id := BaseLanguage_id()
	if multiple == 1 {
		accept_languages := strings.Split(GetConfigurationSimple("SYS_ACCEPT_LANGUAGE"), ",")
		n := len(accept_languages)
		for i := 0; i < n; i++ {
			id, tag := SplitKV(accept_languages[i], ":")
			languages = append(languages, LanguageT{id, tag, Language_flag(id), id == baselanguage_id})
		}
	} else {
		languages = append(languages, LanguageT{baselanguage_id, Language_tag(baselanguage_id), Language_flag(baselanguage_id), true})
	}
	return
}

func AcceptLanguageCodes() (codes []string) {
	al := GetConfigurationSimple("SYS_ACCEPT_LANGUAGE")
	if len(al) > 0 {
		accept_languages := strings.Split(al, ",")
		n := len(accept_languages)
		if n > 0 {
			for i := 0; i < n; i++ {
				id /*tag*/, _ := SplitKV(accept_languages[i], ":")
				codes = append(codes, Language_code(id))
			}
		}
	} else {
		al, _ = ReadLanguageConfigurationFromFile()
		if len(al) > 0 {
			accept_languages := strings.Split(al, ",")
			n := len(accept_languages)
			if n > 0 {
				for i := 0; i < n; i++ {
					id := accept_languages[i]
					if IsDigital(id) {
						codes = append(codes, Language_code(id))
					}
				}
			}
		}
	}
	return
}

func AcceptLanguageSet() (ids, jsontxt string) {
	type languageT struct {
		Code  string
		Tag   string
		Label string
		Flag  string
	}
	idid := []string{}
	mapLD := make(map[string]languageT)
	codee := AcceptLanguageCodes()
	n := len(codee)
	for i := 0; i < n; i++ {
		code := codee[i]
		lid := Language_id(code)
		tag := Language_tag(lid)
		label, _ := SplitK_V(tag, " [")
		lflag := Language_flag(lid)
		mapLD[lid] = languageT{code, tag, label, lflag}
		idid = append(idid, lid)
	}
	b, e := json.Marshal(mapLD)
	if e == nil {
		jsontxt = string(b)
	}
	ids = strings.Join(idid, ",")
	return
}

func GetActualLanguageVersionText(data, language, version_separator string) (str string) {
	ss := strings.Split(data, version_separator)
	for _, v := range ss {
		vv := strings.Split(v, ":")
		if len(vv) == 2 {
			if vv[0] == language {
				str = vv[1]
				break
			}
		}
	}
	return
}

func GetLanguageVersionText(data, base_language, language, version_separator string) (str string) {
	baseversion := data
	ss := strings.Split(data, version_separator)
	for _, v := range ss {
		vv := strings.SplitN(v, ":", 2)
		if len(vv) == 2 {
			if vv[0] == base_language {
				baseversion = vv[1]
			}
			if vv[0] == language {
				str = vv[1]
			}
		}
	}
	if len(str) == 0 {
		str = baseversion
	}
	return
}

func WriteLanguageConfigurationFile(accept_language string) (err error) {
	filename := filepath.Join(dirRun, "configure", "language")
	if IsExists(filename) {
		os.Remove(filename)
	}
	e := ioutil.WriteFile(filename, []byte(accept_language), 0644)
	if e != nil {
		err = errors.New("write language configuration file: " + e.Error())
	}
	sys_accept_language = accept_language
	if strings.Count(sys_accept_language, ",") > 0 {
		sys_acceptmultiplelanguage = true
	}
	return
}

func ReadLanguageConfigurationFromFile() (accept_languages, language_tags string) {
	filename := filepath.Join(dirRun, "configure", "language")
	if IsExists(filename) {
		b, err := ioutil.ReadFile(filename)
		if err == nil {
			accept_languages = string(b)
		}
	} /* else {
		errorLog.Println("file not exists:", filename)
	}*/
	if len(accept_languages) == 0 {
		accept_languages = "2"
	}
	tags := []string{}
	idid := strings.Split(accept_languages, ",")
	for i, n := 0, len(idid); i < n; i++ {
		tags = append(tags, Language_tag(idid[i]))
	}
	language_tags = strings.Join(tags, ",")
	return
}
func WriteBaselanguageConfigurationFile(base_language string) (err error) {
	filename := filepath.Join(dirRun, "configure", "baselanguage")
	if IsExists(filename) {
		os.Remove(filename)
	}
	e := ioutil.WriteFile(filename, []byte(base_language), 0644)
	if e != nil {
		err = errors.New("write baselanguage configuration file: " + e.Error())
	}
	return
}

func ReadBaselanguageConfigurationFromFile() (base_language string) {
	filename := filepath.Join(dirRun, "configure", "baselanguage")
	if IsExists(filename) {
		b, err := ioutil.ReadFile(filename)
		if err == nil {
			base_language = string(b)
		}
	}
	if len(base_language) == 0 {
		base_language = "2"
	}
	return
}

func ReadFileVariables(filename, language string) (mapVar map[string]string) {
	mapVar = make(map[string]string)
	if IsExists(filename) {
		b, err := ioutil.ReadFile(filename)
		if err == nil && len(b) > 0 {
			definition := gjson.ParseBytes(b)
			if definition.Exists() {
				lang := ""
				languages := []string{}
				definition.ForEach(func(k, v gjson.Result) bool {
					languages = append(languages, k.String())
					return true
				})
				exists, _ := In_array(language, languages)
				if !exists {
					ss := strings.Split(language, "_") //zh_TW -> zh
					if len(ss) == 2 {
						exists, _ = In_array(ss[0], languages)
						if exists {
							lang = ss[0]
						}
					}
					if len(lang) == 0 && len(languages) > 0 {
						lang = languages[0]
					}
				} else {
					lang = language
				}
				if len(lang) > 0 {
					o := definition.Get(lang)
					if o.Exists() {
						o.ForEach(func(k, v gjson.Result) bool {
							if v.IsObject() {
							} else if v.IsArray() {
							} else {
								mapVar[k.String()] = v.String()
							}
							return true
						})
					}
				}
			}
		}
	}
	return
}
