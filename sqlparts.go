package base

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func SQL_emptytable(tablename string) (err error) {
	if db != nil {
		asql := ""
		switch DB_type {
		case SQLite:
			asql = "DELETE FROM " + tablename
			_, err = db.Exec(asql)
			if err == nil {
				asql = "DELETE FROM sqlite_sequence WHERE name = '" + tablename + "'"
				_, err = db.Exec(asql)
			}
		case MySQL:
			asql = "truncate table " + tablename
			_, err = db.Exec(asql)
		}
	} else {
		err = errors.New("db not open")
	}
	return
}

func SQL_dropindex(idxname, tablename string) (err error) {
	if db != nil {
		asql := "drop index " + idxname
		switch DB_type {
		case SQLite:
		case MySQL:
			asql += " on " + tablename
		}
		_, err = db.Exec(asql)
	} else {
		err = errors.New("db not open")
	}
	return
}

func SQL_fromdual() (fd string) {
	switch DB_type {
	case SQLite:
	case MySQL:
		fd = " FROM DUAL"
	}
	return
}

func SQL_pagelimit(ipage int) (ss string) {
	return SQL_pagelimitN(ipage, PAGE_ROWS)
}

func SQL_pagelimitN(ipage, pagesize int) (ss string) { //ipage start at: 1
	switch DB_type {
	case SQLite:
		ss = " limit " + fmt.Sprintf("%d", pagesize) + " OFFSET " + fmt.Sprintf("%d", (ipage-1)*pagesize)
	case MySQL:
		ss = " limit " + fmt.Sprintf("%d", (ipage-1)*pagesize) + "," + fmt.Sprintf("%d", pagesize) //+ fmt.Sprintf("%d", ipage*PAGE_ROWS)
	}
	return
}

func SQL_limitN(n int) (ss string) {
	switch DB_type {
	case SQLite:
		ss = " limit " + fmt.Sprintf("%d", n) + " OFFSET 0"
	case MySQL:
		ss = " limit 0," + fmt.Sprintf("%d", n)
	}
	return
}

func SQL_unixtimestamp(field string) (ss string) {
	switch DB_type {
	case SQLite:
		ss = "strftime('%s', " + field + ",'UTC')" /*no UTC, x hours local difference*/
	case MySQL:
		ss = "unix_timestamp(" + field + ")"
	}
	return
}

func SQL_Ymd2date(strfield string) (ss string) {
	switch DB_type {
	case SQLite:
		ss = "date('" + strfield + "')"
	case MySQL:
		ss = "STR_TO_DATE('" + strfield + "','%Y-%m-%d')"
	}
	return
}

func SQL_dateformatYmd(field string) (ss string) {
	switch DB_type {
	case SQLite:
		ss = "strftime('%Y-%m-%d'," + field + ")"
	case MySQL:
		ss = "date_format(" + field + ",'%Y-%m-%d')"
	}
	return
}

func SQL_dateformat19(field string) (ss string) {
	txt := field
	layout := "2006-01-02 15:04:05"
	_, e := time.Parse(layout, field)
	if e != nil {
		txt = time.Now().Format(layout)
	}
	/*layout := "2006-01-02 15:04:05"
	t, e := time.ParseInLocation(layout, field, time.Now().Location())
	if e == nil {
		txt = t.UTC().Format(layout)//会导致扣减时区默认小时数
	}*/
	switch DB_type {
	case SQLite:
		ss = "strftime('%Y-%m-%d %H:%M:%S','" + txt + "')"
	case MySQL:
		ss = "date_format('" + txt + "','%Y-%m-%d %H:%i:%s')"
	}
	return
}

func SQL_timeformat19(tm time.Time) (ss string) {
	txt := tm.Format("2006-01-02 15:04:05")
	switch DB_type {
	case SQLite:
		ss = "strftime('%Y-%m-%d %H:%M:%S','" + txt + "')"
	case MySQL:
		ss = "date_format('" + txt + "','%Y-%m-%d %H:%i:%s')"
	}
	return
}

func SQL_timedefault() (ss string) {
	switch DB_type {
	case SQLite:
		ss = "strftime('%Y-%m-%d %H:%M:%S','" + ZERO_TIME + "')"
	case MySQL:
		ss = "date_format('" + ZERO_TIME + "','%Y-%m-%d %H:%i:%s')"
	}
	return
}

func SQL_beforenow(interval string) (ss string) {
	ss = SQL_dateformatYmd(SQL_datesub("now", interval))
	return
}

func SQL_now() (now string) {
	switch DB_type {
	case SQLite:
		now = "current_timestamp"
	case MySQL:
		now = "now()"
	}
	return
}

// NNN day/hour/minute/second/month/year
func SQL_datesub(field, interval string) (ss string) { //interval 1 day/8 hour
	fld := field
	if field == "now" {
		fld = SQL_now()
	}
	switch DB_type {
	case SQLite:
		ss = "datetime(" + fld + ",'-" + interval + "')"
	case MySQL:
		ss += "date_sub(" + fld + ",interval " + interval + ")"
	}
	return
}

// NNN day/hour/minute/second/month/year
func SQL_dateadd(field, interval string) (ss string) { //interval: 1 day/8 hour
	fld := field
	if field == "now" {
		fld = SQL_now()
	}
	nnn := Str2int(interval)
	si := ""
	switch DB_type {
	case SQLite:
		si = strings.ToLower(interval)
		if nnn > 1 {
			si += "s"
		}
		ss = "datetime(" + fld + ",'+" + si + "')"
	case MySQL:
		si = strings.ToUpper(interval)
		ss += "date_add(" + fld + ",interval " + si + ")"
	}
	return
}

/*
sqlite3
	NNN days
	NNN hours
	NNN minutes
	NNN.NNNN seconds
	NNN months
	NNN years

mysql:
DATE_ADD(date,INTERVAL expr type)
	MICROSECOND
	SECOND
	MINUTE
	HOUR
	DAY
	WEEK
	MONTH
	QUARTER
	YEAR
*/

// fieldname: "time_created"  or  constant: "'2023-06-16 12:12:12'"
func SQL_datediff(fielda, fieldb string) (ss string) { /* days */
	afld, bfld := fielda, fieldb
	if afld == "now" {
		afld = SQL_now()
	}
	if bfld == "now" {
		bfld = SQL_now()
	}
	switch DB_type {
	case SQLite:
		ss = "cast(julianday(" + afld + ")-julianday(" + bfld + ") as int)"
	case MySQL:
		ss = "datediff(" + afld + "," + bfld + ")"
	}
	return
}

func SQL_concat(abc ...string) (ss string) {
	switch DB_type {
	case SQLite:
		ss = strings.Join(abc, "||")
	case MySQL:
		ss = "CONCAT(" + strings.Join(abc, ",") + ")"
	default:
		ss = "CONCAT(" + strings.Join(abc, ",") + ")"
	}
	return
}

func SQL_concat_column(alias, column string) (ss string) { //id*frequencyofusage~valuetext
	abc := []string{}
	b := []rune(column)
	n := len(b)
	s := ""
	for i := 0; i < n; i++ {
		ch := b[i]
		if IsAtomChar(ch) {
			s += string(ch)
		} else {
			if len(s) > 0 {
				abc = append(abc, alias+"."+s)
			}
			abc = append(abc, "'"+string(ch)+"'")
			s = ""
		}
	}
	if len(s) > 0 {
		abc = append(abc, alias+"."+s)
	}
	ss = SQL_concat(abc...)
	return
}

func SQL_left(field string, n int) (ss string) {
	switch DB_type {
	case SQLite:
		ss = "SUBSTR(" + field + ",1," + strconv.Itoa(n) + ")"
	case MySQL:
		ss = "LEFT(" + field + "," + strconv.Itoa(n) + ")"
	}
	return
}

func SQL_CRLF() (ss string) {
	switch DB_type {
	case SQLite:
		ss = "char(13)||char(10)"
	case MySQL:
		ss = "'\\r\\n'"
	}
	return
}

func SQL_LF() (ss string) {
	switch DB_type {
	case SQLite:
		ss = "char(10)"
	case MySQL:
		ss = "'\\n'"
	}
	return
}

func SQL_groupconcat(field, separator, orderby string) (ss string) {
	ss = "GROUP_CONCAT("
	switch DB_type {
	case SQLite:
		/*not accept ORDER BY,so need alternative method in code.*/
		ss += field + "," + separator
	case MySQL:
		ss += field
		if len(orderby) > 0 {
			ss += " ORDER BY " + orderby
		}
		ss += " SEPARATOR " + separator
	}
	ss += ")"
	return
}

func SQL_escape(ss string) (rr string) {
	switch DB_type {
	case SQLite:
		rr = SQLiteEscape(ss)
	case MySQL:
		rr = MySQLEscape(ss)
	}
	return
}

func MySQLEscape(ss string) (rr string) {
	pChar := []rune(ss)
	n := len(pChar)
	rr = "'"
	for i := 0; i < n; i++ {
		ch := pChar[i]
		switch ch {
		case '\'':
			rr += "\\'"
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
	rr += "'"
	return
}

func SQLiteEscape(ss string) (rr string) {
	rr = ""
	pChar := []rune(ss)
	n := len(pChar)
	lastchartype := 0 //1:x'0d',x'0a'  2:default
	for i := 0; i < n; i++ {
		ch := pChar[i]
		switch ch {
		case '\'':
			switch lastchartype {
			case 1:
				rr += "||"
			case 2:
				rr += "'||"
			}
			rr += "x'27'"
			lastchartype = 1
		case '\r':
			switch lastchartype {
			case 1:
				rr += "||"
			case 2:
				rr += "'||"
			}
			rr += "x'0d'"
			lastchartype = 1
		case '\n':
			switch lastchartype {
			case 1:
				rr += "||"
			case 2:
				rr += "'||"
			}
			rr += "x'0a'"
			lastchartype = 1
		default:
			switch lastchartype {
			case 0:
				rr += "'"
			case 1:
				rr += "||'"
			}
			rr += string(ch)
			lastchartype = 2
		}
	}
	switch lastchartype {
	case 0:
		rr += "''"
	case 2:
		rr += "'"
	}
	return
}

func SQL_default(valtype string) (str string) {
	switch valtype {
	case "string", "text", "password", "ipv4", "ipv6", "dots":
		str = "''"
	case "int", "long", "decimal", "float":
		str = "0"
	case "time":
		str = SQL_timedefault()
	}
	return
}

func SQL_value(valtype, val string) (str string) {
	switch valtype {
	case "string", "text", "password", "ipv4", "ipv6", "dots":
		str = "'" + val + "'"
	case "int", "long", "decimal", "float":
		if IsDigital(val) {
			str = val
		} else {
			str = "0"
		}
	case "time":
		tm, e := Str2time(strings.Trim(val, "'"))
		if e == nil {
			str = SQL_timeformat19(tm)
		} else {
			str = SQL_timedefault()
		}
	}
	return
}
