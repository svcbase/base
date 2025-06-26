package base

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"

	"github.com/svcbase/configuration"
	"github.com/svcbase/ipaddr"

	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"fmt"
	"image"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/natefinch/lumberjack"
)

var software /*GetDir*/, version, preferredMAC, software_title, HTTP_port string
var issuer, issueremail, issuerdomain, manageremail string
var softwaresuperior_id int
var softwaresuperior_language, softwaresuperior_email string

// var SM2_privatekey string
var membership, organCode, organName string
var membershipExpiration bool
var cookieStore *sessions.CookieStore
var administratorAuthority string
var situation string
var mapEcho map[string]string
var mapKV map[string]string
var CacheDone chan bool
var DDSM2_publicKeyPem string
var DDSM2_privateKeyPem string

const (
	DailyDecodingQuota   = 100
	FloatPrecisionFactor = 0.0000001
)

var TodayDecodingNumber = 0

const (
	PAGE_ROWS           = 20
	BLANK_CUTSET        = "\r\n\t "
	defaultBufSize      = 4096
	DEFAULT_STRING_SIZE = "255"
	DEFAULT_IPV4_SIZE   = "16"
	DEFAULT_IPV6_SIZE   = "40"
	DEFAULT_DOTIDS_SIZE = "80"
	ZERO_TIME           = "0000-01-01 00:00:00"
)

/*setup wizard situation*/
const (
	INITIAL_STATE            = "under_deployment"
	SETADMINISTRATOR_STATE   = "welcome"
	SUPPORTEDLANGUAGES_STATE = "language"
	SETDBMS_STATE            = "dbms"
	OPENREGISTRATION_STATE   = "registration"
	SETSMTP_STATE            = "smtp"
	INSERVICE_STATE          = "inservice"
)

const (
	Usertype_customer      = 1
	Usertype_partner       = 2
	Usertype_staffer       = 8
	Usertype_director      = 9
	Usertype_master        = 99
	Usertype_administrator = 100
)

type dependencyT struct {
	CSS []string
	JS  []string
}

var aLogger, eLogger, vLogger lumberjack.Logger
var AccessLogger, VisitLogger, ErrorLogger *log.Logger
var dirRun, dirRes, dirLog string

var mapUsertype map[int]string   //1:customer 8:staffer
var mapLang map[string]string    //en:1	zh:2
var mapLangTag map[string]string //1:English [en]  2:Chinese 中文 [zh]
var mapLangFlag map[string]string
var mapLangCurr map[string]string
var mapWidget map[string]dependencyT

// var mapDateFormat map[string]string
// var mapTimeFormat map[string]string
var sys_acceptmultiplelanguage bool
var sys_accept_language string
var Meta_data Metadata
var Salesequence Sequence
var Tokens TokenSet

func init() {
	membership = "basic"
	db = nil
	Tokens.Init()
	mapUsertype = make(map[int]string)
	mapLang = make(map[string]string)
	mapLangTag = make(map[string]string)
	mapLangFlag = make(map[string]string)
	mapLangCurr = make(map[string]string)
	mapWidget = make(map[string]dependencyT)
	//mapDateFormat = make(map[string]string)
	//mapTimeFormat = make(map[string]string)
	mapKV = make(map[string]string)
	sys_acceptmultiplelanguage = false
	CacheDone = make(chan bool)

	init_channel()
}

func DailyDecodingQuotaStr() (quota string) {
	quota = strconv.Itoa(DailyDecodingQuota)
	return
}

func Membership() (ms string) {
	ms = membership
	return
}

func Setmembership(member string) {
	membership = member
}

func SetDDSM2KeyPem(gy, sy string) {
	DDSM2_publicKeyPem = DecodeByKey(gy, DeepdataKeys())
	DDSM2_privateKeyPem = DecodeByKey(sy, DeepdataKeys())
}

func Expiration() (flag bool) {
	flag = membershipExpiration
	return
}

func Setexpiration(expiration bool) {
	membershipExpiration = expiration
}

func Manageremail() (me string) {
	me = manageremail
	return
}

func Setmanageremail(email string) {
	manageremail = email
}

func Issuerdomain() (id string) {
	id = issuerdomain
	return
}

func Setissuerdomain(domain string) {
	issuerdomain = domain
}

func Issueremail() (ie string) {
	ie = issueremail
	return
}

func Setissueremail(email string) {
	issueremail = email
}

func Issuer() (i string) {
	i = issuer
	return
}

func Setissuer(i string) {
	issuer = i
}

func SetV(key, val string) {
	mapKV[key] = val
}

func GetV(key string) (val string) {
	val = mapKV[key]
	return
}

/*func SetSM2privatekey(privatekey string) {
	SM2_privatekey = privatekey
}

func SM2privatekey() (privatekey string) {
	privatekey = SM2_privatekey
	return
}*/

func Setorgancode(code string) {
	organCode = code
}

func Getorgancode() string {
	return organCode
}

func Setorganname(name string) {
	organName = name
}

func Getorganname() string {
	return organName
}

func Hasright(memberlevel string) (flag bool) {
	flag = false
	l1, l2, l3 := "basic", "advanced", "professional"
	if len(memberlevel) > 0 {
		switch Membership() {
		case l1:
			flag = (memberlevel == l1)
		case l2:
			flag = (memberlevel == l1 || memberlevel == l2)
		case l3:
			flag = (memberlevel == l1 || memberlevel == l2 || memberlevel == l3)
		}
	} else {
		flag = true
	}
	return
}

func Stsongttf() (fn string) {
	fn = filepath.Join(dirRes, "res", "stsong.ttf")
	return
}

func Simfangttf() (fn string) {
	fn = filepath.Join(dirRes, "res", "simfang.ttf")
	return
}

func Simheittf() (fn string) {
	fn = filepath.Join(dirRes, "res", "simhei.ttf")
	return
}

func NewCookieStore(software string) (cs *sessions.CookieStore) {
	cookieStore = sessions.NewCookieStore([]byte(software + "-authentication-key"))
	cs = cookieStore
	return
}

func CreateSubdirs(dirs string) {
	ss := strings.Split(dirs, ",")
	for _, name := range ss {
		SetRunSubdir(name)
	}
}
func SetParam(sw_title string) {
	preferredMAC = RecordMAC()
	software_title = sw_title
	dirLog = filepath.Join(dirRun, "logs")
	CreateSubdirs("logs,configure,data,download,illustration,temp,uploads,wx")
	filename := filepath.Join(dirRes, "res", "stsong.ttf")
	if IsExists(filename) {
		//SongParser.Parse(filename)
	} else {
		fmt.Println("file not exist:", filename)
	}
	loadAdministratorAuthority()
	LoadMetadata()
	Salesequence.Open(dirRun, "sale")
}

func LoadMetadata() {
	e := Meta_data.Loadfromfile()
	if e != nil {
		fmt.Println("metadata reset error:", e.Error())
	} else {
		fmt.Println(EchoTxt("t_metasuccess"))
	}
}

func GetCookieStore() (cs *sessions.CookieStore) {
	cs = cookieStore
	return
}

func SetPreferredMAC() (mac string) {
	macs := GetMacAddrs()
	nmacs := len(macs)
	if nmacs > 0 {
		sort.Strings(macs)
		mac = StrMD5(macs[0])
	}
	preferredMAC = mac
	return
}

func GetPreferredMAC() (mac string) {
	mac = preferredMAC
	return
}
func GetRunDir() (ss string) {
	ss = dirRun
	return
}

func SetRunDir(ss string) {
	dirRun = ss
}

func SetVersion(ss string) {
	version = ss
}

func GetVersion() (ss string) {
	ss = version
	return
}

func GetResDir() (ss string) {
	ss = dirRes
	return
}

func GetLogDir() (ss string) {
	ss = dirLog
	return
}

func Superior() (superior_id int, superior_language, superior_email string) {
	superior_id = softwaresuperior_id
	superior_language = softwaresuperior_language
	superior_email = softwaresuperior_email
	return
}

func GetUploadsDir() (ss string) {
	ss = filepath.Join(dirRun, "uploads")
	return
}

func GetTempDir() (ss string) {
	ss = filepath.Join(dirRun, "temp")
	return
}

func SetRunSubdir(sub string) (spath string) {
	spath = filepath.Join(dirRun, sub)
	if !IsExists(spath) {
		MakeDir(spath)
	}
	return
}

func GetSoftware() string {
	return software
}

func SoftwareTitle() (title string) {
	title = software_title
	return
}

func GetAccessLog() *log.Logger {
	return AccessLogger
}

func GetVisitLog() *log.Logger {
	return VisitLogger
}

func GetErrorLog() *log.Logger {
	return ErrorLogger
}

func SetHttpport(http_port string) {
	HTTP_port = http_port
}

func writeMacfile(filename, mac string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open the file", err.Error())
	} else {
		file.WriteString(StrMD5(mac))
		file.Close()
	}
}

func RecordMAC() (mac string) {
	mac = ""
	macs := GetMacAddrs()
	nmacs := len(macs)
	if nmacs > 0 {
		sort.Strings(macs)
	}
	macsfile := filepath.Join(dirRun, "macs")
	file, err := os.OpenFile(macsfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Open file failure: ", macsfile)
		return
	} else {
		file.WriteString(time.Now().Format("2006-01-02 15:04:05") + "\n")
		file.WriteString(strings.Join(macs, ",") + "\n")
	}
	defer file.Close()

	//fmt.Println("MACS:", strings.Join(macs, ","))
	macfile := filepath.Join(dirRun, "identifier")
	if !IsExists(macfile) {
		if nmacs > 0 {
			mac = StrMD5(macs[0])
			writeMacfile(macfile, mac)
		}
	} else {
		flag := false
		bb, err := ioutil.ReadFile(macfile)
		if err == nil {
			m_a_c := TrimBLANK(string(bb))
			n := len(macs)
			for i := 0; i < n; i++ {
				if StrMD5(macs[i]) == m_a_c {
					flag = true
					mac = m_a_c
					break
				}
			}
		}
		if !flag && nmacs > 0 {
			mac = StrMD5(macs[0])
			writeMacfile(macfile, mac)
		}
	}
	preferredMAC = mac
	return
}

func IsInServiceState() (flag bool) {
	flag = (Situation() == INSERVICE_STATE)
	return
}

func SituationMOK(w http.ResponseWriter, r *http.Request) (flag bool) {
	flag = false
	if IsBlackIP(r) {
		http.Redirect(w, r, "/noright", http.StatusFound)
	} else {
		switch Situation() {
		case INSERVICE_STATE:
			flag = true
		case INITIAL_STATE:
			http.Redirect(w, r, "/administrator", http.StatusFound)
		}
	}
	return
}

func SituationOK(w http.ResponseWriter, r *http.Request) (flag bool) {
	flag = false
	if IsBlackIP(r) {
		http.Redirect(w, r, "/noright", http.StatusFound)
	} else {
		switch Situation() {
		case INSERVICE_STATE:
			if Undermaintenance() {
				http.Redirect(w, r, "/undermaintenance", http.StatusFound)
			} else {
				flag = true
			}
		case INITIAL_STATE:
			http.Redirect(w, r, "/administrator", http.StatusFound)
		default:
			http.Redirect(w, r, "/underconstruction", http.StatusFound)
		}
	}
	return
}

func SetAdministratorAuthority(authority string) {
	administratorAuthority = authority
	filename := filepath.Join(dirRun, "administrator")
	err := ioutil.WriteFile(filename, []byte(administratorAuthority), 0600)
	if err != nil {
		ErrorLogger.Println("write administrator:", err.Error())
	}
}

type organT struct {
	ID          string
	Code        string
	Name        string
	Type        string
	Gateway     string
	Environment string
	Tenantcode  string
}

func SetOrgan2file(orgId, orgCode, orgName, orgType, orgGateway, orgEnv, tenantcode string) {
	v := organT{orgId, orgCode, orgName, orgType, orgGateway, orgEnv, tenantcode}
	dt, _ := json.Marshal(v)
	data := EncodeParam(string(dt))
	filename := filepath.Join(dirRun, "organ")
	err := ioutil.WriteFile(filename, []byte(data), 0600)
	if err != nil {
		ErrorLogger.Println("write organ:", err.Error())
	}
}

func LoadOrganfile() (orgId, orgCode, orgName, orgType, orgGateway, orgEnv, tenantCode string) {
	filename := filepath.Join(dirRun, "organ")
	if IsExists(filename) {
		b, e := ioutil.ReadFile(filename)
		if e == nil {
			if len(b) > 0 {
				dt := DecodeParam(string(b))
				var v organT
				e = json.Unmarshal([]byte(dt), &v)
				if e == nil {
					orgId = v.ID
					orgCode = v.Code
					orgName = v.Name
					orgType = v.Type
					orgGateway = v.Gateway
					orgEnv = v.Environment
					tenantCode = v.Tenantcode
				}
			}
		}
	}
	return
}

func GetSystemUserID() string {
	return "1"
}

func GetUsertype(usertype_id int) (usertype string) {
	var ok bool
	if usertype, ok = mapUsertype[usertype_id]; !ok {
		usertype = "free"
	}
	return
}

func GetAdministratorAuthority() (authority string) {
	authority = administratorAuthority
	return
}

func loadAdministratorAuthority() {
	filename := filepath.Join(dirRun, "administrator")
	if IsExists(filename) {
		b, err := ioutil.ReadFile(filename)
		if err == nil {
			administratorAuthority = string(b)
		}
	}
	return
}

func IsAdministratorOpened() (opened bool) {
	opened = (len(administratorAuthority) > 0)
	return
}

func makeConnectionSignature(mapParam map[string]string) (signature string) {
	keys := []string{}
	for k, _ := range mapParam {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	tt := ""
	n := len(keys)
	for i := 0; i < n; i++ {
		k := keys[i]
		tt += k + "=" + mapParam[k]
	}
	signature = StrMD5(tt)
	return
}

func SendNotice(subject, body string) {
	NoticeChannel <- SysNoticeT{subject, body} //send notice by frame.NoticeTask
}

func GetDBparameter(key, value string) (db_connection, db_type, db_file, db_host, db_port, db_user, db_pswd, db_database string) {
	filename := filepath.Join(dirRun, "configure", "database")
	if IsExists(filename) {
		file, _ := os.Open(filename)
		defer file.Close()
		Decoder := gob.NewDecoder(file)
		var M map[string]map[string]string
		e := Decoder.Decode(&M)
		if e != nil {
			ErrorLogger.Println(e.Error())
		} else {
			for connection_name, mapParam := range M {
				if v, ok := mapParam[key]; ok && (v == value) {
					db_type = mapParam["db_type"]
					switch db_type {
					case "sqlite":
						db_file = strings.Replace(mapParam["db_file"], "[SYS]", filepath.Join(dirRun, "data"), -1)
					case "mysql":
						db_host = mapParam["db_host"]
						db_port = mapParam["db_port"]
						db_user = mapParam["db_user"]
						if db_pswd, ok = mapParam["db_pswd"]; ok && (len(db_pswd) > 0) {
							db_pswd = DecodeByKey(db_pswd, MacKeyGen(1111))
						}
						db_database = mapParam["db_database"]
					}
					db_connection = connection_name
					break
				}
			}
		}
	}
	return
}

func DB_in_Connecting() {
	DB_cyclenode = DB_Connecting
}

func DB_in_Initializing() {
	DB_cyclenode = DB_Initializing
}

func DB_in_Loading() {
	DB_cyclenode = DB_Loading
}

func DB_in_Troubleshooting() {
	DB_cyclenode = DB_Troubleshooting
}

func DB_in_Undermaintenance() {
	DB_cyclenode = DB_Undermaintenance
}

func DB_in_Running() {
	DB_cyclenode = DB_Running
}

func DB_in_Stop() {
	DB_cyclenode = DB_Stop
}

func Is_DB_in_Troubleshooting() bool {
	return (DB_cyclenode == DB_Troubleshooting)
}

func Is_DB_in_Running() bool {
	return (DB_cyclenode == DB_Running)
}

/*
	func ReadHint() {
		langpath := filepath.Join(dirRes, "res")
		dir, err := ioutil.ReadDir(langpath)
		if err == nil {
			for _, fd := range dir {
				fname := fd.Name()
				if !fd.IsDir() && strings.HasSuffix(fname, ".lang") {
					lang := ""
					ss := strings.Split(fname, ".")
					if len(ss) == 2 {
						lang = ss[0]
					}
					fname = filepath.Join(langpath, fname)
					fi, err := os.Open(fname)
					if err == nil {
						defer fi.Close()
						eof := false
						r := bufio.NewReader(fi)
						for {
							s, e := r.ReadString('\n')
							if e == io.EOF {
								eof = true
							}
							if len(s) > 0 {
								ss := strings.SplitN(strings.TrimRight(s, "\r\n"), "=", 2)
								if len(ss) == 2 {
									k := strings.TrimRight(ss[0], " \t")
									v := strings.TrimLeft(ss[1], " \t")
									mapHint[lang+k] = v
								}
							}
							if eof {
								break
							}
						}
					}
				}
			}
		}
	}
*/
func SigninRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/signin?r="+url.QueryEscape(r.RequestURI), http.StatusFound)
}

func LoginRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login?r="+url.QueryEscape(r.RequestURI), http.StatusFound)
	//http: superfluous(多余的) response.WriteHeader call from base.LoginRedirect
}

func ReadDualLogin(r *http.Request, cs *sessions.CookieStore) (signed bool, user_id, username string, user_type int) {
	user_id = "0"
	signed, username, user_type = ReadManagerLogin(r, cs)
	if !signed {
		var u_id int64
		signed, username, _, u_id, _, _, _, user_type = ReadLogin(r, cs)
		user_id = strconv.FormatInt(u_id, 10)
	}
	return
}

func ReadManagerLogin(r *http.Request, cs *sessions.CookieStore) (signed bool, user_name string, user_type int) {
	session, err := cs.Get(r, software+"-manage")
	if err == nil {
		if s, ok := session.Values["user"]; ok {
			user_name = s.(string)
		}
		if s, ok := session.Values["signed"]; ok {
			signed = s.(bool)
		}
		if s, ok := session.Values["user_type"]; ok {
			user_type = s.(int)
		}
	}
	return
}

func SessionValues(r *http.Request, cs *sessions.CookieStore) (values string) {
	vv := []string{}
	session, err := cs.Get(r, software+"-common")
	if err == nil {
		if s, ok := session.Values["user"]; ok {
			vv = append(vv, "user_name="+s.(string))
		}
		if s, ok := session.Values["user_id"]; ok {
			vv = append(vv, "user_id="+strconv.FormatInt(s.(int64), 10))
		}
		if s, ok := session.Values["user_type"]; ok {
			vv = append(vv, "user_type="+strconv.Itoa(s.(int)))
		}
		if s, ok := session.Values["usertype"]; ok {
			vv = append(vv, "usertype="+s.(string))
		}
		if s, ok := session.Values["manager_id"]; ok {
			vv = append(vv, "manager_id="+strconv.FormatInt(s.(int64), 10))
		}
		if s, ok := session.Values["superior_id"]; ok {
			vv = append(vv, "superior_id="+strconv.FormatInt(s.(int64), 10))
		}
		if s, ok := session.Values["orgchart_id"]; ok {
			vv = append(vv, "orgchart_id="+strconv.Itoa(s.(int)))
		}
		if s, ok := session.Values["orgchart_dotids"]; ok {
			vv = append(vv, "orgchart_dotids="+s.(string))
		}
		if s, ok := session.Values["currency_preferred"]; ok {
			vv = append(vv, "currency_preferred="+s.(string))
		}
		if s, ok := session.Values["dateformat_preferred"]; ok {
			vv = append(vv, "dateformat_preferred="+s.(string))
		}
		if s, ok := session.Values["timeformat_preferred"]; ok {
			vv = append(vv, "timeformat_preferred="+s.(string))
		}
		if s, ok := session.Values["clientlanguage_id"]; ok {
			vv = append(vv, "clientlanguage_id="+s.(string))
		}
	}
	if len(vv) == 0 {
		session, err = cs.Get(r, software+"-manage")
		if err == nil {
			vv = append(vv, "user_id=0")
			if s, ok := session.Values["user"]; ok {
				vv = append(vv, "user_name="+s.(string))
			}
			if s, ok := session.Values["user_type"]; ok {
				vv = append(vv, "user_type="+strconv.Itoa(s.(int)))
			}
			if s, ok := session.Values["usertype"]; ok {
				vv = append(vv, "usertype="+s.(string))
			}
			if s, ok := session.Values["clientlanguage_id"]; ok {
				vv = append(vv, "clientlanguage_id="+s.(string))
			}
		}
	}
	values = strings.Join(vv, "&")
	return
}

func ReadLogin(r *http.Request, cs *sessions.CookieStore) (signed bool, user_name, email_address string, user_id, manager_id, superior_id int64, active, user_type int) {
	signed = false
	user_name = ""
	user_id, user_type, manager_id, superior_id, active = 0, 0, 0, 0, 0
	session, err := cs.Get(r, software+"-common")
	if err == nil {
		if s, ok := session.Values["user"]; ok {
			user_name = s.(string)
			email_address = user_name
		}
		if s, ok := session.Values["signed"]; ok {
			signed = s.(bool)
		}
		if s, ok := session.Values["user_id"]; ok {
			user_id = s.(int64)
		}
		if s, ok := session.Values["user_type"]; ok {
			user_type = s.(int)
		}
		if s, ok := session.Values["manager_id"]; ok {
			manager_id = s.(int64)
		}
		if s, ok := session.Values["superior_id"]; ok {
			superior_id = s.(int64)
		}
		if s, ok := session.Values["active"]; ok {
			active = s.(int)
		}
	}
	return
}

func ReadOrgchart(r *http.Request, cs *sessions.CookieStore) (signed bool, orgchart_id int, orgchart_dotids string) {
	signed = false
	session, err := cs.Get(r, software+"-common")
	if err == nil {
		if s, ok := session.Values["signed"]; ok {
			signed = s.(bool)
		}
		if s, ok := session.Values["orgchart_id"]; ok {
			orgchart_id = s.(int)
		}
		if s, ok := session.Values["orgchart_dotids"]; ok {
			orgchart_dotids = s.(string)
		}
	}
	return
}

func LeadingCount(ss string, leading rune) (count int) { //	TAB: 9
	b := []rune(ss)
	n := len(b)
	for i := 0; i < n; i++ {
		if b[i] == leading {
			count++
		} else {
			break
		}
	}
	return
}

func TotalASCII(ss string) (n int64) { //total string each byte value
	n = 0
	b := []byte(ss)
	for _, ch := range b {
		n += int64(ch)
	}
	return n
}

func IsDigit(ch byte) bool {
	return (ch == '0') || (ch == '1') || (ch == '2') || (ch == '3') || (ch == '4') ||
		(ch == '5') || (ch == '6') || (ch == '7') || (ch == '8') || (ch == '9')
}

func IsDigital(str string) bool {
	n := 0
	flag := true
	p := []rune(str)
	for i := range p {
		n++
		ch := p[i]
		if !((ch == '0') || (ch == '1') || (ch == '2') || (ch == '3') || (ch == '4') || (ch == '5') ||
			(ch == '6') || (ch == '7') || (ch == '8') || (ch == '9')) {
			if i == 0 && ch == '-' {
			} else {
				flag = false
				break
			}
		}
	}
	if n == 0 {
		flag = false
	}
	return flag
}

func IsNumber(str string) bool {
	n := 0
	flag := true
	p := []rune(str)
	for i := range p {
		n++
		ch := p[i]
		if !((ch == '0') || (ch == '1') || (ch == '2') || (ch == '3') || (ch == '4') || (ch == '5') ||
			(ch == '6') || (ch == '7') || (ch == '8') || (ch == '9') || (ch == '.')) {
			if i == 0 && ch == '-' {
			} else {
				flag = false
				break
			}
		}
	}
	if n == 0 {
		flag = false
	}
	return flag
}

func isSpecials(ch rune) bool {
	return (ch == '(') || (ch == ')') || (ch == '<') || (ch == '>') || (ch == '@') ||
		(ch == ',') || (ch == ';') || (ch == ':') || (ch == '\\') || (ch == '"') ||
		(ch == '[') || (ch == ']') || (ch == '=') || (ch == '&') || (ch == '/') ||
		(ch == '!') || (ch == '#') || (ch == '$') || (ch == '%') || (ch == '^') ||
		(ch == '*') || (ch == '+') || (ch == '{') || (ch == '}') || (ch == '\'') ||
		(ch == '?') || (ch == '~') || (ch == '`') || (ch == '|')
}

func IsAtomChar(ch rune) bool {
	return ch > 32 && ch < 127 && !isSpecials(ch)
}

func In_array(val string, array []string) (exists bool, index int) {
	exists = false
	index = -1

	for i, v := range array {
		if val == v {
			index = i
			exists = true
			return
		}
	}
	return
}

func In_arrayint(val int, array []int) (exists bool, index int) {
	exists = false
	index = -1

	for i, v := range array {
		if val == v {
			index = i
			exists = true
			return
		}
	}
	return
}

func In_arrayint64(val int64, array []int64) (exists bool, index int) {
	exists = false
	index = -1

	for i, v := range array {
		if val == v {
			index = i
			exists = true
			return
		}
	}
	return
}

func SubMonth(year, month, n int) (y, m int) {
	var yy int = n / 12
	var mm int = n - yy*12
	y = year - yy
	m = month - mm
	if m <= 0 {
		y -= 1
		m += 12
	}
	return
}

func AddMonth(year, month, n int) (y, m int) {
	var yy int = n / 12
	var mm int = n - yy*12
	y = year + yy
	m = month + mm
	if m > 12 {
		y += 1
		m -= 12
	}
	return
}

// sample: Weed_array(arr,"123,456")
func Weed_array(array []string, weed string) (newarray []string) {
	weedd := strings.Split(weed, ",")
	n := len(array)
	for i := 0; i < n; i++ {
		v := array[i]
		exists, _ := In_array(v, weedd)
		if !exists {
			newarray = append(newarray, v)
		}
	}
	return
}

func PswdValid(password string) bool {
	flag := true
	n := len(password) //Password must be between 6 to 16 characters.
	if n < 6 || n > 16 {
		flag = false
	}
	return flag
}

func IsWeakpassword(password string) bool {
	flag := false
	if db != nil {
		asql := "select count(1) ct from weakpassword where password=?"
		row := db.QueryRow(asql, password)
		if row != nil {
			var ct int = 0
			if row.Scan(&ct) == nil {
				flag = (ct > 0)
			}
		}
	}
	return flag
}

// ---
func CharType(ch rune) (ty int) { //1: 汉 2: digit 3: letter 4: space 5: return 6: newline 0: other
	ty = 0
	if len(string(ch)) > 1 {
		ty = 1
	} else if ch >= 48 && ch <= 57 {
		ty = 2
	} else if (ch >= 65 && ch <= 90) || (ch >= 97 && ch <= 122) {
		ty = 3
	} else if ch == '\t' || ch == ' ' {
		ty = 4
	} else if ch == '\r' {
		ty = 5
	} else if ch == '\n' {
		ty = 6
	}
	return
}

func StrNextWord(s []rune, i, n int) (j int, curword string) {
	j = -1
	if i < n {
		ch := s[i]
		ty := CharType(ch)
		j = i + 1
		if ty >= 2 {
			for j < n {
				if ty < 5 && CharType(s[j]) == ty {
					j++
				} else if (ty == 2 || ty == 3) && s[j] == '-' {
					j++
				} else {
					break
				}
			}
		}
		curword = string(s[i:j])
		//fmt.Println( []byte(curword) )
		//fmt.Printf( " %x\r\n",ch )
		//if ch == 0x3000 || ch == 0x2003 {	//IdeographicSpace EmSpace
		//	curword = "　"
		//fmt.Println( "***" )
		//}
	}
	return
}

func FirstWords(str string, maxBytes int) (gist string) {
	m := maxBytes
	s := []rune(str)
	i, n := 0, len(s)
	for i < n {
		j, wd := StrNextWord(s, i, n)
		if j > 0 {
			wd = strings.Trim(wd, "\r\n")
			nwd := len(wd)
			if nwd > 0 {
				if m >= nwd {
					m -= nwd
					gist += wd
				} else {
					break
				}
			}
			i = j
		} else {
			break
		}
	}
	return
}

func WordSlice(str string) (words []string) {
	s := []rune(str)
	i, n := 0, len(s)
	for i < n {
		j, wd := StrNextWord(s, i, n)
		if j > 0 {
			wd = strings.Trim(wd, "\t\r\n ")
			if len(wd) > 0 {
				flag := true
				if len(wd) == 1 {
					c := []rune(wd)
					flag = IsAtomChar(c[0])
				}
				/*if flag {
					flag = !IsNumber(wd)
				}*/
				if flag {
					exists, _ := In_array(wd, words)
					if !exists {
						words = append(words, wd)
					}
				}
			}
			i = j
		} else {
			break
		}
	}
	return
}

func FullWordSlice(str string) (wordss []string) {
	words := WordSlice(str)
	wordss = words
	n := len(words)
	for i := 0; i < n; i++ {
		ss := strings.Join(words[:i+1], "")
		ee := strings.Join(words[n-i-1:], "")
		exists, _ := In_array(ss, wordss)
		if !exists {
			wordss = append(wordss, ss)
		}
		exists, _ = In_array(ee, wordss)
		if !exists {
			wordss = append(wordss, ee)
		}
	}
	return
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func FloatEqual(a, b string) (flag bool) {
	flag = false
	if IsNumber(a) && IsNumber(b) {
		af, _ := strconv.ParseFloat(a, 32)
		bf, _ := strconv.ParseFloat(b, 32)
		as := fmt.Sprintf("%.2f", af)
		bs := fmt.Sprintf("%.2f", bf)
		flag = (as == bs)
	}
	return
}

func FloatLessthan(a, b string) (flag bool) {
	flag = false
	if IsNumber(a) && IsNumber(b) {
		af, _ := strconv.ParseFloat(a, 32)
		bf, _ := strconv.ParseFloat(b, 32)
		flag = (af < bf)
	}
	return
}

func CalcCommission(cash_pay float64) (commission float64) {
	commission = 0.0
	var k50 float64 = 50000
	var k10 float64 = 10000
	var k1 float64 = 1000
	if cash_pay >= k50 {
		commission += (cash_pay - k50) * 0.45
		cash_pay = k50
	}
	if cash_pay >= k10 {
		commission += (cash_pay - k10) * 0.4
		cash_pay = k10
	}
	if cash_pay >= k1 {
		commission += (cash_pay - k1) * 0.35
		cash_pay = k1
	}
	commission += cash_pay * 0.3
	return
}

func Pages(nrows, pagesize int) (pages int) {
	psize := pagesize
	if psize == 0 {
		psize = PAGE_ROWS
	}
	pages = (nrows + psize - 1) / psize
	return
}
func Paging(nrows, pagesize int, cur_page string) (ipage, pages int) {
	psize := pagesize
	if psize == 0 {
		psize = PAGE_ROWS
	}
	pages = (nrows + psize - 1) / psize
	ipage = 1
	if IsDigital(cur_page) {
		ipage, _ = strconv.Atoi(cur_page)
		if ipage <= 0 {
			ipage = 1
		}
	}
	return
}

func ToplevelDomain(domain string) (toplevel string) {
	toplevel = domain //localhost
	if domain != "127.0.0.1" {
		ss := strings.Split(domain, ".")
		n := len(ss)
		if n >= 2 {
			toplevel = ss[n-2] + "." + ss[n-1]
		}
	}
	return
}

func GetItemName(items string, itemcode int) (itemname string) {
	mapItem := String2Map(items)
	var ok bool
	itemname, ok = mapItem[strconv.Itoa(itemcode)]
	if !ok {
		itemname, _ = mapItem["*"]
	}
	return
}

func ExeExt() (ext string) {
	file, _ := exec.LookPath(os.Args[0]) //execute file name
	exeFilename, _ := filepath.Abs(file)
	ext = filepath.Ext(exeFilename)
	return
}

func GetDir(soft_ware string) (exeDir, exeFile, runDir, resDir string) {
	software = soft_ware
	file, _ := exec.LookPath(os.Args[0]) //execute file name
	path, _ := filepath.Abs(file)
	exeDir, exeFile = filepath.Split(path) //execute file path and short name
	//exeDir, _ = os.Getwd()//systemctl服务不安全
	if strings.Index(exeDir, "go-build") > 0 {
		fmt.Println("not support go-build mode!")
	} else {
		runDir = filepath.Join(exeDir, "execute", software)
		if !IsExists(runDir) {
			runDir = exeDir
		}
		resDir = filepath.Join(exeDir, "deployment", software)
		if !IsExists(resDir) {
			resDir = exeDir
		}
	}
	//fmt.Println(os.Getenv("LANG"))
	fmt.Println("start "+software+" dir:", exeDir, "OS:", runtime.GOOS) //runtime.GOOS: linux,darwin,...
	//	fmt.Println("run dir:", runDir)
	//	fmt.Println("res dir:", resDir)
	dirRun = runDir
	dirRes = resDir
	filename := filepath.Join(dirRes, "res", "start.echo")
	if IsExists(filename) {
		mapEcho = LoadConfigure(filename)
	} else {
		mapEcho = make(map[string]string)
	}
	filename = filepath.Join(dirRes, "res", "partner.chip")
	if IsExists(filename) {
		src, e := ioutil.ReadFile(filename)
		if e == nil {
			dst := make([]byte, len(src))
			var n int
			n, e = base64.StdEncoding.Decode(dst, src)
			if e == nil {
				type chipT struct {
					User     int
					Language string
					Email    string
				}
				var chip chipT
				e = json.Unmarshal(dst[0:n], &chip)
				if e == nil {
					softwaresuperior_id = chip.User
					softwaresuperior_language = chip.Language
					softwaresuperior_email = chip.Email
				}
			}
		}
	}
	return
}

func EchoTxt(txt string) (echo string) {
	if v, ok := mapEcho[txt]; ok {
		echo = v
	}
	return
}

func Situation() (str string) {
	str = situation
	return
}

func RightSituation(dir_run string) (err error) {
	filename := filepath.Join(dirRun, "situation")
	if IsExists(filename) {
		b, e := ioutil.ReadFile(filename)
		if e != nil {
			err = errors.New("read situation file: " + e.Error())
		} else {
			situation = TrimBLANK(string(b))
			if len(situation) == 0 {
				err = errors.New("empty situation file, please remove it and retry.")
			}
		}
	} else {
		filename = filepath.Join(dirRun, "administrator")
		if IsExists(filename) {
			os.Remove(filename)
		}
		err = WriteSituation(dirRun, INITIAL_STATE)
	}
	return
}

func GetBack2DBMS_state(dir_run string) (err error) {
	err = WriteSituation(dir_run, SETDBMS_STATE)
	return
}

func WriteSituation(dir_run, ss string) (err error) {
	filename := filepath.Join(dir_run, "situation")
	situation = ss
	e := ioutil.WriteFile(filename, []byte(situation), 0666)
	if e != nil {
		err = errors.New("write situation file: " + e.Error())
	}
	return
}

func Software() (sw string) {
	sw = software
	return
}

func Undermaintenance() (flag bool) {
	flag = (GetConfigurationSimple("SYS_UNDER_MAINTENANCE") == "1")
	return
}

func SetLogger() (accessLog, visitLog, errorLog *log.Logger, err error) {
	err = nil
	accesslogfile := filepath.Join(dirLog, "access.log")
	aLogger = lumberjack.Logger{
		Filename:   accesslogfile,
		MaxSize:    1,  // megabytes after which new file is created
		MaxBackups: 20, // number of backups
		MaxAge:     7,  //days
	}
	accessfile, aerr := os.OpenFile(accesslogfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if aerr == nil {
		AccessLogger = log.New(accessfile, "", log.Ldate|log.Ltime|log.Lshortfile)
		AccessLogger.SetOutput(&aLogger)
	} else {
		err = aerr
		fmt.Println(aerr)
	}
	if err == nil {
		errorlogfile := filepath.Join(dirLog, "error.log")
		eLogger = lumberjack.Logger{
			Filename:   errorlogfile,
			MaxSize:    1,  // megabytes after which new file is created
			MaxBackups: 20, // number of backups
			MaxAge:     7,  //days
		}
		errorfile, eerr := os.OpenFile(errorlogfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if eerr == nil {
			ErrorLogger = log.New(errorfile, "", log.Ldate|log.Ltime|log.Lshortfile)
			ErrorLogger.SetOutput(&eLogger)
		} else {
			err = eerr
			fmt.Println(eerr)
		}
		visitlogfile := filepath.Join(dirLog, "visit.log")
		vLogger = lumberjack.Logger{
			Filename:   visitlogfile,
			MaxSize:    1,  // megabytes after which new file is created
			MaxBackups: 50, // number of backups
			MaxAge:     10, //days
		}
		visitfile, verr := os.OpenFile(visitlogfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if verr == nil {
			VisitLogger = log.New(visitfile, "", log.Ldate|log.Ltime|log.Lshortfile)
			VisitLogger.SetOutput(&vLogger)
		} else {
			err = verr
			fmt.Println(verr)
		}
	}
	accessLog = AccessLogger
	visitLog = VisitLogger
	errorLog = ErrorLogger
	return
}

func VisitLog(vlog *log.Logger, r *http.Request, user_id int64) {
	type visitT struct {
		Time_visited    string
		User_id         int64
		Ip_address      string
		Accept_language string
		Referer         string
		User_agent      string
		Uri             string
	}
	var v visitT
	v.Time_visited = time.Now().Format("2006-01-02 15:04:05")
	v.User_id = user_id
	v.Ip_address = ipaddr.RemoteIp(r)
	v.Accept_language = r.Header.Get("Accept-Language")
	v.Referer = r.Header.Get("Referer")
	v.User_agent = r.Header.Get("User-Agent")
	v.Uri = r.RequestURI
	dt, _ := json.Marshal(v)
	vlog.Println(string(dt))
}

func GetMacAddrs() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err == nil {
		for _, netInterface := range netInterfaces {
			if netInterface.Flags&net.FlagUp != 0 && netInterface.Flags&net.FlagLoopback == 0 {
				macAddr := netInterface.HardwareAddr.String()
				if len(macAddr) > 0 {
					addrs, e := netInterface.Addrs()
					if e == nil {
						for _, addr := range addrs {
							ipNet, isValidIpNet := addr.(*net.IPNet)
							if isValidIpNet && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
								if !strings.HasPrefix(ipNet.String(), "169.254") {
									macAddrs = append(macAddrs, macAddr)
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

func KeyGen(mac_seed, salt int64) (keys []int) {
	ak := []byte(ReverseStr(strings.Trim(strconv.FormatInt(int64(mac_seed*720319+salt), 10), "-")))
	var n int = len(ak) / 2
	bk := []byte(ReverseStr(strings.Trim(strconv.FormatInt(int64(mac_seed*201229+salt), 10), "-")))
	var m int = len(bk) / 2
	n = Min(n, m)
	for i := 0; i < n; i++ {
		a := Str2int(string(ak[2*i : 2*i+2]))
		b := Str2int(string(bk[2*i : 2*i+2]))
		if a > 63 {
			a = 100 - a
		}
		if b > 63 {
			b = 100 - b
		}
		keys = append(keys, a)
		keys = append(keys, b)
	}
	return
}

func MacKeyGen(salt int64) (keys []int) {
	keys = KeyGen(TotalASCII(preferredMAC), salt)
	return
}

func ConfigurationKeyGen(salt int64) (keys []int) {
	keys = KeyGen(TotalASCII("DeepData:"+software), salt)
	return
}

func MobileSlice(mobile string) string {
	segments := []string{}
	mb := []byte(mobile)
	n := len(mb)
	var m int = n / 2
	for i := 1; i < m; i++ {
		ss := string(mb[2*i : 2*i+2])
		exists, _ := In_array(ss, segments)
		if !exists {
			segments = append(segments, ss)
		}
		ss = ReverseStr(ss)
		exists, _ = In_array(ss, segments)
		if !exists {
			segments = append(segments, ss)
		}
	}
	m = (n - 1) / 2
	for i := 0; i < m; i++ {
		ss := string(mb[1+2*i : 2*i+3])
		exists, _ := In_array(ss, segments)
		if !exists {
			segments = append(segments, ss)
		}
	}
	m = n / 3
	for i := 0; i < m; i++ {
		ss := string(mb[3*i : 3*i+3])
		exists, _ := In_array(ss, segments)
		if !exists {
			segments = append(segments, ss)
		}
		bb := []byte(ss)
		ch := bb[1]
		bb[1] = bb[2]
		bb[2] = ch
		ss = string(bb)
		exists, _ = In_array(ss, segments)
		if !exists {
			segments = append(segments, ss)
		}
	}
	m = (n - 1) / 3
	for i := 0; i < m; i++ {
		ss := string(mb[1+3*i : 3*i+4])
		exists, _ := In_array(ss, segments)
		if !exists {
			segments = append(segments, ss)
		}
	}
	m = (n - 2) / 3
	for i := 0; i < m; i++ {
		ss := string(mb[2+3*i : 3*i+5])
		exists, _ := In_array(ss, segments)
		if !exists {
			segments = append(segments, ss)
		}
	}
	sort.Strings(segments)
	return strings.Join(segments, " ")
}

func MobileQueryExpression(mobile string) (expression string) {
	mb := []byte(mobile)
	expression = string(mb[0:3]) + " +" + string(mb[3:6]) + " +" + string(mb[6:9]) + " +" + string(mb[9:])
	return
}

func FulltextQueryExpression(q string) (expression string) {
	qq := WordSlice(strings.ToLower(q))
	expression = ""
	for i, w := range qq {
		if i > 0 {
			expression += " "
		}
		expression += "+"
		if len(w) == len([]rune(w)) {
			expression += w
			if w != "use" && w != "non" { //https://sczcx.iteye.com/blog/2145722 停用词
				expression += "*"
			}
		} else {
			expression += w
		}
		if len(expression) > 32 {
			break
		}
	}
	return
}

func SortMapKeys(mp map[string]string) (n int, keys []string) {
	for k, _ := range mp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	n = len(keys)
	return
}

func LoadConfigure(filename string) map[string]string {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()
	return readConfigureFrom(bufio.NewReader(file))
}

func LoadEncodedConfigure(filename string) map[string]string {
	data := ""
	if IsExists(filename) {
		b, e := ioutil.ReadFile(filename)
		if e == nil && len(b) > 0 {
			data = DecodeByKey(string(b), []int{12, 24, 9, 18})
		}
	}
	return readConfigureFrom(bufio.NewReader(strings.NewReader(data)))
}

func readConfigureFrom(buf *bufio.Reader) map[string]string {
	mapParam := make(map[string]string)
	for {
		line, _, err := buf.ReadLine()
		ne := bytes.Trim(line, " \t")
		ln := bytes.Trim(ne, string(BOM))
		if err == io.EOF {
			break
		}
		if !bytes.HasPrefix(ln, []byte{'#'}) && !bytes.Equal(ln, []byte{}) {
			parts := bytes.SplitN(ln, []byte{'='}, 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(string(parts[0]))
				val := strings.Trim(strings.TrimSpace(string(parts[1])), "\"")
				mapParam[key] = val
			}
		}
	}
	return mapParam
}

func ReadAdminConfiguration() {
	respath := filepath.Join(dirRes, "res")
	dir, err := ioutil.ReadDir(respath)
	if err == nil {
		for _, fd := range dir {
			fname := fd.Name()
			if !fd.IsDir() && strings.HasSuffix(fname, "_admin.kv") { //后续需与 *_tip.kv 有效划分，保证准确够用
				lang := ""
				ss := strings.Split(fname, "_")
				if len(ss) == 2 {
					lang = ss[0]
				}
				if langid, ok := mapLang[lang]; ok {
					lang_id := Str2int(langid)
					fname = filepath.Join(respath, fname)
					fi, err := os.Open(fname)
					if err == nil {
						defer fi.Close()
						i := 0
						eof := false
						r := bufio.NewReader(fi)
						for {
							s, e := r.ReadString('\n')
							if e == io.EOF {
								eof = true
							}
							if len(s) > 0 {
								ss := strings.SplitN(strings.TrimRight(s, "\r\n"), "=", 2)
								if len(ss) == 2 {
									k := strings.TrimRight(ss[0], " \t")
									v := strings.TrimLeft(ss[1], " \t")
									id := configuration.KeyID(k)
									if id == 0 {
										i++
										id = i
										configuration.SetKey(k, id, false)
									}
									configuration.Set(id, lang_id, 0, 10, v)
								}
							}
							if eof {
								break
							}
						}
					}
				}
			}
		}
	}
}

// productcategory,property
func CombineRoadmap(identifier, subentities string) (roadmap []string) {
	roadmap = append(roadmap, identifier)
	if len(subentities) > 0 {
		ss := strings.Split(subentities, ".")
		roadmap = append(roadmap, ss...)
	}
	return
}

func GetNsegment(str string, n int) (segment []string) { //100.101.102.103 -> GetNsegment(ss,2) return 100.101
	ss := strings.Split(str, ".")
	m := len(ss)
	for i := 0; i < n; i++ {
		if i < m {
			segment = append(segment, ss[i])
		}
	}
	return
}

func WhereeByRoadmap(roadmap []string, ids []int64) (wheree []string) {
	n := len(ids)
	if len(roadmap) >= n {
		for i := 0; i < n; i++ {
			wheree = append(wheree, "a."+strings.Join(roadmap[0:i+1], "_")+"_id="+strconv.FormatInt(ids[i], 10))
		}
	}
	return
}

func WhereByRoadmap(roadmap []string, ids []string) (where string) {
	wheres := []string{}
	n := len(ids)
	if len(roadmap) >= n {
		for i := 0; i < n; i++ {
			wheres = append(wheres, strings.Join(roadmap[0:i+1], "_")+"_id="+ids[i])
		}
	}
	return strings.Join(wheres, " and ")
}

func WhereByRoadmapExcept(roadmap []string, ids []string, excepts []string) (where string) {
	wheres := []string{}
	n := len(ids)
	if len(roadmap) >= n {
		for i := 0; i < n; i++ {
			ss := strings.Join(roadmap[0:i+1], "_") + "_id"
			exists, _ := In_array(ss, excepts)
			if !exists {
				wheres = append(wheres, ss+"="+ids[i])
			}
		}
	}
	return strings.Join(wheres, " and ")
}

func GetWidgetDependency(widget string) (dependencies []string) {
	if w, ok := mapWidget[widget]; ok {
		n := len(w.CSS)
		for i := 0; i < n; i++ {
			dependencies = append(dependencies, `<link rel="stylesheet" href="`+w.CSS[i]+`">`)
		}
		//rf := ""
		//if GetConfigurationSimple("SITE_ADDRESS") == "localhost" {
		//	rf = "?" + strconv.FormatInt(time.Now().Unix(), 10)
		//}
		n = len(w.JS)
		for i := 0; i < n; i++ {
			dependencies = append(dependencies, `<script type="text/javascript" src="`+w.JS[i]+ /*rf+*/ `"></script>`)
		}
	}
	return
}

func GetWidgetDependencyNames(widget string) (dependencies []string) {
	wi := widget
	if widget == "input" || widget == "text" {
		wi = "fulllanguage"
	}
	if w, ok := mapWidget[wi]; ok {
		dependencies = append(dependencies, w.CSS...)
		dependencies = append(dependencies, w.JS...)
	}
	return
}

func Int64ArrayJoin(ids []int64, separator string) (txt string) {
	idid := []string{}
	for _, id := range ids {
		idid = append(idid, strconv.FormatInt(id, 10))
	}
	txt = strings.Join(idid, separator)
	return
}

func GetWidgetJS(widget string) (jsjs []string) {
	if w, ok := mapWidget[widget]; ok {
		for _, v := range w.JS {
			vv := v
			i := strings.LastIndex(v, "/")
			if i >= 0 {
				b := []byte(v)
				vv = string(b[i+1:])
			}
			jsjs = append(jsjs, vv)
		}
	}
	return
}

func MergeStringArray(all *([]string), add []string) {
	n := len(add)
	for i := 0; i < n; i++ {
		exists, _ := In_array(add[i], *all)
		if !exists {
			*all = append(*all, add[i])
		}
	}
}

func MergeDependency(all *([]string), add []string) {
	MergeStringArray(all, add)
}

func CalcColumnWidth(ncols int) (colwidths []string) {
	width := 100
	avg := width / ncols
	for i := 0; i < ncols; i++ {
		colwidth := avg
		if i == ncols-1 {
			colwidth = width - i*avg
		}
		colwidths = append(colwidths, strconv.Itoa(colwidth))
	}
	return
}

const (
	RESOLUTION_RATIO = 72.0
	MM_PER_INCH      = 25.4
)

func WordPageMargin() (left, top, right, bottom float64) {
	mm := strings.Split(GetConfigurationSimple("UI_WORD_PAGEMARGIN"), "/") //25.4/31.7/25.4/31.7
	if len(mm) == 4 {
		top = Str2float64(mm[0]) * RESOLUTION_RATIO / MM_PER_INCH
		right = Str2float64(mm[1]) * RESOLUTION_RATIO / MM_PER_INCH
		bottom = Str2float64(mm[2]) * RESOLUTION_RATIO / MM_PER_INCH
		left = Str2float64(mm[3]) * RESOLUTION_RATIO / MM_PER_INCH
	}
	return
}

func WordPageHeaderFooter() (header, footer float64) {
	mm := strings.Split(GetConfigurationSimple("UI_WORD_PAGEHEADERFOOTER"), "/") //15/17.5
	if len(mm) == 2 {
		header = Str2float64(mm[0]) * RESOLUTION_RATIO / MM_PER_INCH
		footer = Str2float64(mm[1]) * RESOLUTION_RATIO / MM_PER_INCH
	}
	return
}

func ImagePixelsWidthHeight(filename string) (width, height int) {
	if IsExists(filename) {
		file, err := os.Open(filename)
		if err == nil {
			c, _, e := image.DecodeConfig(file)
			if e == nil {
				width = c.Width
				height = c.Height
			}
			file.Close()
		}
	}
	return
}

func ImageWidthHeight(filename string) (width, height float64) {
	if IsExists(filename) {
		file, err := os.Open(filename)
		if err == nil {
			c, _, e := image.DecodeConfig(file)
			if e == nil {
				ratio := 0.5625 //72.0 / 128.0
				width = float64(c.Width) * ratio
				height = float64(c.Height) * ratio
			}
			file.Close()
		}
	}
	return
}

func Unit2pixel(unit float64) (pixels int) {
	ratio := 0.5625
	pixels = int(unit / ratio)
	return
}

func Pixel2unit(pixel int) (unit float64) {
	ratio := 0.5625
	unit = float64(pixel) * ratio
	return
}

func SiteName() (site_name string) {
	site_name = strings.Trim(GetConfigurationSimple("SITE_NAME"), " ")
	if len(site_name) == 0 || site_name == "sitename" {
		site_name = software_title
	}
	return
}

func SiteAddress(client_ipaddress string) (site_address string) {
	site_address = GetConfigurationSimple("SITE_ADDRESS")
	s_a := strings.ToLower(site_address)
	if strings.HasPrefix(s_a, "http://") {
		b := []byte(site_address)
		site_address = string(b[7:])
	}
	if strings.HasPrefix(site_address, "https://") {
		b := []byte(site_address)
		site_address = string(b[8:])
	}
	site_address = strings.Trim(site_address, "/ ")
	if len(site_address) == 0 || site_address == "localhost" {
		if client_ipaddress == "127.0.0.1" {
			site_address = "localhost"
		} else {
			site_address = ipaddr.LocalIP()
		}
	}
	/*if len(HTTP_port) > 0 && HTTP_port != "80" {
		site_address += ":" + HTTP_port
	}*/
	return
}

func SiteUrlandLogo(client_ipaddress string) (protocol, site_address, logo_image string) {
	protocol = GetConfigurationSimple("SYS_PROTOCOL")
	site_address = SiteAddress(client_ipaddress)
	logo_image = GetConfigurationSimple("SITE_LOGO_IMAGE")
	if len(logo_image) == 0 {
		logo_image = "logo_default.jpg"
	}
	if !strings.HasPrefix(logo_image, "http") {
		logo_image = protocol + "://" + site_address + "/img/" + logo_image
	}
	return
}

func DeepdataKeys() (keys []int) {
	keys = KeyGen(TotalASCII("DeepData-BigData Insights"), 888)
	return
}

// filename format: project_a1bae84094d454870b0012facd6e822f.ddat
func CheckFileSignature(filename string) (e error) {
	if IsExists(filename) {
		i := strings.LastIndex(filename, "_")
		j := strings.LastIndex(filename, ".")
		if i > 0 && j > i {
			b := []byte(filename)
			srcsign := string(b[i+1 : j])
			dessign := FileMD5(filename)
			if srcsign != dessign {
				e = errors.New("not match: " + srcsign + "," + dessign)
			}
		}
	} else {
		e = errors.New(filename + "not exists.")
	}
	return
}

func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func Iconhtml(icon string) (iconhtml string) {
	if strings.HasPrefix(icon, "fa-") {
		iconhtml += `<i class="fa fa-fw ` + icon + `"></i>`
	} else {
		iconhtml += `<img src="/img/` + icon + `">`
	}
	return
}

func USCIdistrictcode(usci string) (districtcode string) {
	if len(usci) == 18 {
		b := []byte(usci)
		districtcode = string(b[2:8])
	}
	return
}

func DaysBetweenTwoDate(dateStart, dateEnd time.Time) (days int) {
	days = int((dateEnd.Unix() - dateStart.Unix()) / (60 * 60 * 24))
	return
}
