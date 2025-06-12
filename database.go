package base

import (
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/svcbase/configuration"
	"github.com/svcbase/ipaddr"
	"github.com/svcbase/ipblacklist"
	"github.com/tidwall/gjson"
)

var db *sql.DB
var DB_database string

const (
	SQLite              = 1
	MySQL               = 2
	Oracle              = 3
	DB_Connecting       = 1
	DB_Initializing     = 2
	DB_Loading          = 3
	DB_Troubleshooting  = 4
	DB_Undermaintenance = 5
	DB_Running          = 6
	DB_Stop             = 9
)

var DB_type int //sqlite,mysql,oracle,sqlserver
var DB_cyclenode int

func DB() (d_b *sql.DB) {
	d_b = db
	return
}

func SetDB(d_b *sql.DB) {
	db = d_b
}

func DBclose() {
	if db != nil {
		db.Close()
		db = nil
	}
}

func DBname() (database string) {
	database = DB_database
	return
}

func SetDBtype(ss string) int {
	DB_type = DBtype(ss)
	return DB_type
}

func DBtype(ss string) (dbtype int) {
	switch ss {
	case "SQLite":
		dbtype = SQLite
	case "MySQL":
		dbtype = MySQL
	case "Oracle":
		dbtype = Oracle
	}
	return
}

func UseMySQL() bool {
	return (DB_type == MySQL)
}

func DoDBconnect(language_id, db_type, db_file, db_host, db_port, db_user, db_pswd, db_database string) (d_b *sql.DB, err error) {
	switch db_type {
	case "sqlite":
		if len(db_file) > 0 {
			d_b, err = sql.Open("sqlite3", db_file+"?cache=shared")
		} else {
			err = errors.New(GetConfigurationLanguage("TIP_NO_DBFILE", language_id) + db_file)
		}
	case "mysql":
		if !IsDigital(db_port) {
			err = errors.New(GetConfigurationLanguage("TIP_DBPORT_MUST", language_id))
		} else if len(db_host) > 0 && len(db_user) > 0 && len(db_database) > 0 {
			AccessLogger.Println("Start connecting MySQL server:", db_host, ":", db_port)
			n, itv, ntrytimes := 20, 3, 0 //total wait 1 minute.
			for i := 0; i < n; i++ {
				ntrytimes++
				d_b, err = sql.Open("mysql", db_user+":"+db_pswd+"@tcp("+db_host+":"+db_port+")/"+db_database+"?charset=utf8&parseTime=true")
				if err == nil { //d_b.SetMaxOpenConns(2000)  d_b.SetMaxIdleConns(100)
					err = d_b.Ping()
					break
				} else {
					ss := err.Error()
					AccessLogger.Println("err:", ss)
					if strings.Contains(ss, "DB connect dial") && strings.Contains(ss, "connection refused") {
						//DB connect dial tcp [::1]:3306: connect: connection refused
						//wait while mysql.service not ready
						AccessLogger.Println("wait", itv, "seconds to detect mysql.service ready: the", i+1, "time.")
						tc := time.After(time.Second * time.Duration(itv))
						<-tc
					}
				}
			}
			AccessLogger.Println("Try connect MySQL", ntrytimes, "times.")
			if n == ntrytimes {
				ErrorLogger.Println("Connect MySQL server overtime!")
			}
		} else {
			err = errors.New(GetConfigurationLanguage("TIP_NO_MYSQL_PARAM", language_id))
		}
	default:
		err = errors.New(GetConfigurationLanguage("TIP_DB_ERRTYPE", language_id))
	}
	return
}

func DBconnect() (d_b *sql.DB, err error) {
	d_b = nil
	_, db_type, db_file, db_host, db_port, db_user, db_pswd, db_database := GetDBparameter("online", "1")
	fmt.Println(db_type, db_file, db_host, db_port, db_user, db_database)
	db, err = DoDBconnect(BaseLanguage_id(), db_type, db_file, db_host, db_port, db_user, db_pswd, db_database)
	if err != nil {
		ErrorLogger.Println("DB connect", err.Error())
	} else {
		d_b = db
		switch db_type {
		case "sqlite":
			DB_type = SQLite
		case "mysql":
			DB_type = MySQL
		}
		DB_database = db_database
	}
	return
}

func ExistTables() (tables int) {
	if db != nil {
		var asql string
		switch DB_type {
		case SQLite:
			asql = "SELECT count(1) ct FROM sqlite_master where type='table' and tbl_name not like 'sqlite_%'"
		case MySQL:
			asql = "SELECT count(1) ct FROM information_schema.tables WHERE TABLE_SCHEMA='" + DB_database + "'"
		}
		row := db.QueryRow(asql)
		if row != nil {
			e := row.Scan(&tables)
			if e != nil {
				ErrorLogger.Println(asql, e.Error())
			}
		}
	}
	return
}

func ReadOneRow(asql, groupby string, nproperties int, instance_id int64) (vals []sql.NullString, err error, n int) {
	n = 0
	vals = make([]sql.NullString, nproperties)
	if instance_id > 0 {
		if db != nil {
			ss := asql + " where a.id=?"
			if strings.Contains(asql, "GROUP_CONCAT") && len(groupby) > 0 {
				ss += groupby
			}
			//fmt.Println("ReadOneRow")
			//fmt.Println(ss)
			//select a.id _rmi_,IFNULL(GROUP_CONCAT(CONCAT(ml.language_id,'#',LEFT(replace(ml.name,'\r\n',''),80)) SEPARATOR '\n'),replace(a.name,'\r\n','')),a.productcategory_id,a.model,a.unitofmeasurement_id,a.merchant_id,a.brand_id,a.placeoforigin,a.barcode,a.propertyjson,a.piclayout,a.picturedimension,a.skudimension,a.pricedimension,IFNULL(al.description,a.description) from product a left join product_languages ml on a.id=ml.product_id left join product_languages al on a.id=al.product_id and al.language_id=2 where a.id=? group by a.id
			//Error 1140: In aggregated query without GROUP BY, expression #1 of SELECT list contains nonaggregated column 'dt.a.id'; this is incompatible with sql_mode=only_full_group_by
			//ui.Form: Error 1055: Expression #15 of SELECT list is not in GROUP BY clause and contains nonaggregated column 'dt.al.description' which is not functionally dependent on columns in GROUP BY clause; this is incompatible with sql_mode=only_full_group_by
			row := db.QueryRow(ss, instance_id)
			if row != nil {
				cans := make([]interface{}, nproperties)
				for i := 0; i < nproperties; i++ {
					cans[i] = &vals[i]
				}
				e := row.Scan(cans...)
				//fmt.Println("ReadOneRow:", e, asql+" where a.id=?", instance_id)
				if e != nil {
					if !NoRowsError(e) {
						err = e
					}
				} else {
					nn := 0
					for i := 0; i < nproperties; i++ {
						nn += len(vals[i].String)
					}
					if nn > 0 {
						n = 1
					}
				}
			}
		} else {
			err = errors.New("database not connected!")
		}
	}
	return
}

func NoRowsError(e error) (flag bool) {
	flag = false
	if e != nil {
		flag = (e.Error() == "sql: no rows in result set")
	}
	return
}

type tableT struct {
	Name, Size, Rows string
}

func DBtables() (tables []tableT, err error) {
	if db != nil {
		var rows *sql.Rows
		var asql string
		switch DB_type {
		case SQLite:
			tablenames := []string{}
			asql = "select tbl_name from sqlite_master where type='table' and tbl_name not like 'sqlite_%'"
			rows, err = db.Query(asql)
			if err == nil {
				var table_name sql.NullString
				for rows.Next() {
					err = rows.Scan(&table_name)
					if err == nil {
						tablenames = append(tablenames, table_name.String)
					} else {
						ErrorLogger.Println(err.Error())
						break
					}
				}
				rows.Close()
			}
			n := len(tablenames)
			for i := 0; i < n; i++ {
				asql = "select count(1) ct from `" + tablenames[i] + "`"
				row := db.QueryRow(asql)
				if row != nil {
					var table_rows string
					err = row.Scan(&table_rows)
					if err == nil {
						tables = append(tables, tableT{tablenames[i], "", table_rows})
					}
				}
			}
		case MySQL:
			asql = "SELECT table_name,table_rows, CONCAT(TRUNCATE(data_length/1024/1024,2),'MB') AS data_size"
			asql += " FROM information_schema.tables WHERE TABLE_SCHEMA='" + DB_database + "'"
			rows, err = db.Query(asql)
			if err == nil {
				var table_name, data_size sql.NullString
				var table_rows string
				for rows.Next() {
					err = rows.Scan(&table_name, &table_rows, &data_size)
					if err == nil {
						tables = append(tables, tableT{table_name.String, data_size.String, table_rows})
					} else {
						ErrorLogger.Println(err.Error())
						break
					}
				}
				rows.Close()
			}
		}
	} else {
		err = errors.New("database not connected!")
	}
	if err != nil {
		ErrorLogger.Println(err.Error())
	}
	return
}

func MatchSignature() (flag bool, err error) {
	flag = false
	if db != nil {
		signature := Meta_data.Signature()
		asql := "select signature from deepdata limit 1"
		row := db.QueryRow(asql)
		if row != nil {
			var dd_signature sql.NullString
			err = row.Scan(&dd_signature)
			if err != nil {
				if NoRowsError(err) {
					err = errors.New("no deepdata signature!")
				} else {
					ErrorLogger.Println(asql, "scan", err.Error())
				}
			} else {
				flag = (signature == dd_signature.String)
			}
		}
	}
	return
}

/*func GetDeepDataSignature() (signature string, err error) {
	if db != nil {
		otables := []string{}
		var rows *sql.Rows
		var asql string
		switch DB_type {
		case SQLite:
			asql = "SELECT tbl_name FROM sqlite_master WHERE type='table' and tbl_name not like 'sqlite_%'"
		case MySQL:
			asql = "SELECT table_name FROM information_schema.tables WHERE TABLE_SCHEMA='" + DB_database + "'"
		}
		rows, err = db.Query(asql)
		if err == nil {
			var table_name sql.NullString
			for rows.Next() {
				err = rows.Scan(&table_name)
				if err == nil {
					otables = append(otables, table_name.String)
				} else {
					ErrorLogger.Println(err.Error())
					break
				}
			}
			rows.Close()
		}
		idid := []string{}
		//identifier->code
		asql = "select code from entity where creator='deepdata' order by code"
		rows, err = db.Query(asql)
		if err == nil {
			var identifier sql.NullString
			for rows.Next() {
				err = rows.Scan(&identifier)
				if err == nil {
					id := identifier.String
					exists, _ := In_array(id, otables)
					if exists {
						idid = append(idid, id)
					} else {
						err = errors.New("table " + id + " not find!")
					}
				} else {
					ErrorLogger.Println(err.Error())
					break
				}
			}
			rows.Close()
		}
		if err == nil {
			if len(idid) == 0 {
				err = errors.New("zero entity table created!")
			} else {
				signature = StrMD5(strings.Join(idid, "="))
			}
		}
	} else {
		err = errors.New("database not connected!")
	}
	return
}*/

func TableExists(name string) (flag bool) {
	flag = false
	if db != nil && len(name) > 0 {
		var asql string
		switch DB_type {
		case SQLite:
			asql = "select count(1) from sqlite_master where type='table' and tbl_name='" + name + "'"
		case MySQL:
			if len(DB_database) > 0 {
				//information_schema:Each MySQL user has the right to access these tables, but can see only the rows in the tables that correspond to objects for which the user has the proper access privileges.
				asql = "select count(1) from information_schema.tables where table_schema='" + DB_database + "' and table_name='" + name + "'"
			}
		}
		if len(asql) > 0 {
			row := db.QueryRow(asql)
			if row != nil {
				ct := 0
				if row.Scan(&ct) == nil {
					flag = (ct == 1)
				}
			}
		}
	}
	return
}

func EmptyDatabase() (err error) {
	if db != nil {
		var tables []tableT
		tables, err = DBtables()
		n := len(tables)
		for i := 0; i < n; i++ {
			asql := "drop table `" + tables[i].Name + "`"
			_, err = db.Exec(asql)
			if err != nil {
				break
			}
		}
	} else {
		err = errors.New("database not connected!")
	}
	return
}

func RemoveDBconnection(signature string) (err error) {
	filename := filepath.Join(dirRun, "configure", "database")
	if IsExists(filename) {
		infile, _ := os.Open(filename)
		Decoder := gob.NewDecoder(infile)
		var M map[string]map[string]string
		err = Decoder.Decode(&M)
		if err != nil {
			ErrorLogger.Println(err.Error())
		} else {
			infile.Close()
			outfile, e := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if e != nil {
				err = e
				ErrorLogger.Println("Failed to open the file", err.Error())
			} else {
				for connection_name, mapParam := range M {
					if v, ok := mapParam["signature"]; ok && (v == signature) {
						delete(M, connection_name)
					}
				}
				encoder := gob.NewEncoder(outfile)
				if err = encoder.Encode(M); err != nil {
					ErrorLogger.Println("DB connection save", err.Error())
				}
				outfile.Close()
			}
		}
	}
	return
}

func LiveDBconnection(signature string) (err error) {
	filename := filepath.Join(dirRun, "configure", "database")
	if IsExists(filename) {
		infile, _ := os.Open(filename)
		Decoder := gob.NewDecoder(infile)
		var M map[string]map[string]string
		err = Decoder.Decode(&M)
		if err != nil {
			ErrorLogger.Println(err.Error())
		} else {
			infile.Close()
			outfile, e := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if e != nil {
				err = e
				ErrorLogger.Println("Failed to open the file", err.Error())
			} else {
				MM := make(map[string]map[string]string)
				for connection_name, mapParam := range M {
					o_signature := mapParam["signature"]
					key := "online"
					if o_signature == signature {
						mapParam[key] = "1"
					} else {
						if _, ok := mapParam[key]; ok {
							delete(mapParam, key)
						}
					}
					MM[connection_name] = mapParam
				}
				encoder := gob.NewEncoder(outfile)
				if err = encoder.Encode(MM); err != nil {
					ErrorLogger.Println("DB connection save", err.Error())
				}
				outfile.Close()
			}
		}
	} else {
		err = errors.New("NO DB configure file")
	}
	return
}

func WriteMySQLDBMS(desPath, dbtitle, dbhost, dbport, dbuser, dbpswd, dbdatabase string) (err error) {
	filename := filepath.Join(desPath, "configure", "database")
	var file *os.File
	file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err == nil {
		info := make(map[string]map[string]string)
		aDB := make(map[string]string)
		aDB["online"] = "1"
		aDB["db_host"] = dbhost
		aDB["db_port"] = dbport
		aDB["db_user"] = dbuser
		aDB["db_pswd"] = EncodeByKey(dbpswd, MacKeyGen(1111))
		aDB["db_database"] = dbdatabase
		aDB["db_type"] = "mysql"
		aDB["signature"] = makeConnectionSignature(aDB)
		info[dbtitle] = aDB
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(info)
		file.Close()
	}
	return
}

func DBlaunch() (err error) {
	filename := filepath.Join(dirRun, "configure", "database")
	if !IsExists(filename) {
		var file *os.File
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			ErrorLogger.Println("Failed to open the file", err.Error())
		} else {
			mysqlpath := filepath.Join(dirRun, "mysql")
			info := make(map[string]map[string]string)
			aDB := make(map[string]string)
			aDB["online"] = "1"
			dbtype := ""
			if IsExists(mysqlpath) {
				dbtype = "mysql"
				aDB["db_host"] = "localhost"
				aDB["db_port"] = "3306"
				aDB["db_user"] = software
				aDB["db_pswd"] = EncodeByKey(software+fmt.Sprintf("%d", TotalASCII(software)), MacKeyGen(1111))
				aDB["db_database"] = software
			} else {
				dbtype = "sqlite"
				aDB["db_file"] = filepath.Join("[SYS]", software+".sqlite3")
			}
			aDB["db_type"] = dbtype
			aDB["signature"] = makeConnectionSignature(aDB)
			info["Native integration ["+dbtype+"]"] = aDB
			encoder := gob.NewEncoder(file)
			if err = encoder.Encode(info); err != nil {
				ErrorLogger.Println("DB launch", err.Error())
			}
			file.Close()
		}
	}
	return
}

func SaveDBconnection(in_signature, online, db_connection, db_type, db_file, db_host, db_port, db_user, db_pswd, db_database string) (out_signature string) {
	M := make(map[string]map[string]string)
	filename := filepath.Join(dirRun, "configure", "database")
	if IsExists(filename) {
		infile, err := os.Open(filename)
		if err == nil {
			Decoder := gob.NewDecoder(infile)
			e := Decoder.Decode(&M)
			if e != nil {
				ErrorLogger.Println(e.Error())
			}
			infile.Close()
		}
	}
	outfile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		ErrorLogger.Println("Failed to open the file", err.Error())
	} else {
		if in_signature != "new" {
			for connection_name, mapParam := range M {
				if v, ok := mapParam["signature"]; ok && (v == in_signature) {
					delete(M, connection_name)
				}
			}
		}
		mapParam := make(map[string]string)
		mapParam["db_type"] = db_type
		switch db_type {
		case "sqlite":
			ss := db_file
			prefix := filepath.Join(dirRun, "data")
			if strings.HasPrefix(ss, prefix) {
				ss = strings.Replace(ss, prefix, "[SYS]", 1)
			}
			mapParam["db_file"] = ss
		case "mysql":
			mapParam["db_host"] = db_host
			mapParam["db_port"] = db_port
			mapParam["db_user"] = db_user
			mapParam["db_pswd"] = EncodeByKey(db_pswd, MacKeyGen(1111))
			mapParam["db_database"] = db_database
		}
		out_signature = makeConnectionSignature(mapParam)
		mapParam["signature"] = out_signature
		if online == "1" {
			mapParam["online"] = online
		}
		M[db_connection] = mapParam
		encoder := gob.NewEncoder(outfile)
		if err = encoder.Encode(M); err != nil {
			ErrorLogger.Println("DB connection save", err.Error())
		}
		outfile.Close()
	}
	return
}

type dbconnectionT struct {
	Name      string
	Signature string
	Online    bool
}

func GetDBconnections() (db_connections []dbconnectionT, onlineParam map[string]string) {
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
				signature := ""
				online := false
				oParam := make(map[string]string)
				oParam["db_connection"] = connection_name
				for k, v := range mapParam {
					switch k {
					case "online":
						online = (v == "1")
					case "signature":
						signature = v
					case "db_file":
						v = strings.Replace(v, "[SYS]", filepath.Join(dirRun, "data"), -1)
					case "db_pswd":
						if len(v) > 0 {
							v = DecodeByKey(v, MacKeyGen(1111))
						}
					}
					oParam[k] = v
				}
				if online {
					onlineParam = oParam
				}
				db_connections = append(db_connections, dbconnectionT{connection_name, signature, online})
			}
		}
	}
	return
}
func ReadUsertype() (err error) {
	if db != nil {
		mapUsertype[100] = "administrator"
		asql := "select id,code from usertype"
		var rows *sql.Rows
		rows, err = db.Query(asql)
		if err == nil {
			for rows.Next() {
				var usertype_id int
				var code sql.NullString
				err = rows.Scan(&usertype_id, &code)
				if err == nil {
					mapUsertype[usertype_id] = code.String
				} else {
					ErrorLogger.Println(asql, "scan", err.Error())
				}
			}
			rows.Close()
		} else {
			ErrorLogger.Println(asql, err.Error())
		}
	}
	return
}
func ReadLang() (err error) {
	if db != nil {
		asql := "select id,code,name,flagimage,preferredcurrency from language"
		var rows *sql.Rows
		rows, err = db.Query(asql)
		if err == nil {
			for rows.Next() {
				var language_id string
				var code, name, flagimage, curr sql.NullString
				err = rows.Scan(&language_id, &code, &name, &flagimage, &curr)
				if err == nil {
					mapLang[code.String] = language_id
					mapLangTag[language_id] = name.String + " [" + code.String + "]"
					mapLangFlag[language_id] = flagimage.String
					mapLangCurr[language_id] = curr.String
				} else {
					ErrorLogger.Println(asql, "scan", err.Error())
				}
			}
			rows.Close()
		} else {
			ErrorLogger.Println(asql, err.Error())
		}
	}
	return
}

func ReadCurrency() (err error) {
	if db != nil {
		asql := "select id,code,name,exchangerate,symbolleft,symbolright from currency where enableflag=1"
		var rows *sql.Rows
		rows, err = db.Query(asql)
		if err == nil {
			for rows.Next() {
				var currency_id int
				var exchangerate float64
				var code, name, symbolleft, symbolright sql.NullString
				err = rows.Scan(&currency_id, &code, &name, &exchangerate, &symbolleft, &symbolright)
				if err == nil {
					MapCurrency[code.String] = CurrencyT{currency_id, code.String, name.String, exchangerate, symbolleft.String, symbolright.String}
				} else {
					ErrorLogger.Println(asql, "scan", err.Error())
				}
			}
			rows.Close()
		} else {
			ErrorLogger.Println(asql, err.Error())
		}
	}
	return
}

/*func ReadDateFormat() (err error) {
	if db != nil {
		asql := "select name,parameter from dateformat"
		var rows *sql.Rows
		rows, err = db.Query(asql)
		if err == nil {
			for rows.Next() {
				var name, parameter sql.NullString
				err = rows.Scan(&name, &parameter)
				if err == nil {
					mapDateFormat[name.String] = parameter.String
				} else {
					ErrorLogger.Println(asql, "scan", err.Error())
				}
			}
			rows.Close()
		} else {
			ErrorLogger.Println(asql, err.Error())
		}
	}
	return
}

func ReadTimeFormat() (err error) {
	if db != nil {
		asql := "select name,parameter from timeformat"
		var rows *sql.Rows
		rows, err = db.Query(asql)
		if err == nil {
			for rows.Next() {
				var name, parameter sql.NullString
				err = rows.Scan(&name, &parameter)
				if err == nil {
					mapTimeFormat[name.String] = parameter.String
				} else {
					ErrorLogger.Println(asql, "scan", err.Error())
				}
			}
			rows.Close()
		} else {
			ErrorLogger.Println(asql, err.Error())
		}
	}
	return
}*/

func ReadConfiguration() (err error) {
	if db != nil {
		sys_config_buf_size := 100
		asql := "select `value` from configuration where `key`='SYS_CONFIG_BUF_SIZE'"
		row := db.QueryRow(asql)
		if row != nil {
			var v sql.NullString
			err = row.Scan(&v)
			if err == nil {
				size := Str2int(v.String)
				if size > sys_config_buf_size { //get system memory size and limit this value
					sys_config_buf_size = size
				}
			} else {
				ErrorLogger.Println(asql, "scan", err.Error())
			}
		}
		configuration.SetMaxBufSize(sys_config_buf_size)
		asql = "select c.id,c.`key`,c.`value`,c.valueconsistent,c.valueencryption,ifnull(cg.priority,0) from configuration c"
		asql += " left join configurationgroup cg on c.configurationgroup_id=cg.id"
		asql += " order by c.id"
		var rows *sql.Rows
		rows, err = db.Query(asql)
		if err == nil {
			for rows.Next() {
				var id, valueconsistent, valueencryption, priority int
				var key, value sql.NullString
				err = rows.Scan(&id, &key, &value, &valueconsistent, &valueencryption, &priority)
				if err == nil {
					enkey := ConfigurationKeyGen(int64(id))
					needencryption := (valueencryption == 1)
					configuration.SetKey(key.String, id, needencryption)
					if valueconsistent == 1 || !SysAcceptMultipleLanguage() {
						v := value.String
						if needencryption && len(v) > 0 {
							v = DecodeByKey(v, enkey)
						}
						configuration.SetSimple(id, priority, v)
					} else {
						if SysAcceptMultipleLanguage() {
							asq := "select id,language_id,`value` from configuration_languages where configuration_id=?"
							var rws *sql.Rows
							rws, err = db.Query(asq, id)
							if err == nil {
								for rws.Next() {
									var cl_id, language_id int
									var val sql.NullString
									err = rws.Scan(&cl_id, &language_id, &val)
									if err == nil {
										v := val.String
										if needencryption && len(v) > 0 {
											v = DecodeByKey(v, enkey)
										}
										configuration.Set(id, language_id, cl_id, priority, v)
									} else {
										ErrorLogger.Println(asq, "scan", err.Error())
										break
									}
								}
								rws.Close()
								if err != nil {
									break
								}
							} else {
								ErrorLogger.Println(asq, err.Error())
								break
							}
						}
					}
				} else {
					ErrorLogger.Println(asql, "scan", err.Error())
					break
				}
			}
			rows.Close()
			sal := GetConfigurationSimple("SYS_ACCEPT_LANGUAGE")
			if len(sal) > 0 {
				sys_accept_language = sal
				if strings.Contains(sys_accept_language, ",") {
					sys_acceptmultiplelanguage = true
				}
			}
		} else {
			ErrorLogger.Println(asql, err.Error())
		}
	}
	return
}
func SetConfigurationSimple(key, value string) {
	//relate frame.Json_saveconfigurationHandler
	if db != nil {
		iid := InstanceRetrieve("configuration", "key", key)
		if iid == 0 {
			asql := "insert into configuration(`key`,`value`,time_created,time_updated) values(?,?," + SQL_now() + "," + SQL_now() + ")"
			rslt, e := db.Exec(asql, key, value)
			if e == nil {
				iid, e = rslt.LastInsertId()
				if e == nil {
					configuration.SetSimple(int(iid), 10, value) //10: top priority
				}
			}
		} else {
			asql := "select c.id,c.valueencryption,cg.priority from configuration c"
			asql += " left join configurationgroup cg on c.configurationgroup_id=cg.id"
			asql += " where c.`key`=?"
			row := db.QueryRow(asql, key)
			if row != nil {
				var configuration_id, valueencryption, priority int
				if row.Scan(&configuration_id, &valueencryption, &priority) == nil {
					configuration.SetSimple(configuration_id, priority, value)
					save_value := value
					if valueencryption == 1 {
						save_value = EncodeByKey(value, ConfigurationKeyGen(int64(configuration_id)))
					}
					asql := "update configuration set `value`=?,time_updated=" + SQL_now() + " where id=?"
					_, e := db.Exec(asql, save_value, configuration_id)
					if e != nil {
						ErrorLogger.Println(asql, save_value, configuration_id, e.Error())
					}
				}
			}
		}
	}
}

func GetConfigurationSimple(key string) (value string) {
	value = configuration.GetSimple(key)
	re := regexp.MustCompile(`^@[0-9]+@$`)
	if re.MatchString(value) {
		id := strings.Trim(value, "@")
		value = ""
		if db != nil {
			asql := "select `value` from configuration where id=?"
			row := db.QueryRow(asql, id)
			if row != nil {
				var v sql.NullString
				if row.Scan(&v) == nil {
					value = v.String
				}
			}
		}
	}
	return
}

func GetConfigurationMultiple(key string) (value string) {
	ids, _ := AcceptLanguageSet()
	idid := strings.Split(ids, ",")
	vals := []string{}
	for _, id := range idid {
		vals = append(vals, id+"#"+GetConfigurationLanguage(key, id))
	}
	return strings.Join(vals, "\n")
}

func GetConfigurationLanguage(key, language_id string) (value string) {
	if SysAcceptMultipleLanguage() {
		value = configuration.Get(key, language_id)
	} else {
		value = configuration.GetSimple(key)
	}
	configuration_id := 0
	re := regexp.MustCompile(`^@[0-9]+-[0-9]+@$`)
	if re.MatchString(value) {
		idid := strings.Split(strings.Trim(value, "@"), "-")
		value = ""
		if len(idid) == 2 {
			configuration_id = Str2int(idid[0])
			id := idid[1]
			asql := "select `value` from configuration_languages where id=?"
			row := db.QueryRow(asql, id)
			if row != nil {
				var v sql.NullString
				if row.Scan(&v) == nil {
					value = v.String
				}
			}
		}
	} else if strings.HasPrefix(value, "@") && strings.HasSuffix(value, "@") {
		id := strings.Trim(value, "@")
		if IsDigital(id) {
			configuration_id = Str2int(id)
			asql := "select `value` from configuration where id=?"
			row := db.QueryRow(asql, configuration_id)
			if row != nil {
				var v sql.NullString
				if row.Scan(&v) == nil {
					value = v.String
				}
			}
		}
	}
	if configuration_id > 0 && len(value) > 0 && configuration.NeedEncryption(configuration_id) {
		value = DecodeByKey(value, ConfigurationKeyGen(int64(configuration_id)))
	}
	return
}

func configurationVariablefilling(ss string) (tt string) {
	tt = ss
	re, e := regexp.Compile("{{[A-Z0-9_]*}}")
	if e == nil {
		vv := re.FindAllString(ss, -1)
		for _, v := range vv {
			key := strings.Trim(v, "{}")
			val := configuration.GetSimple(key)
			if len(val) > 0 {
				tt = strings.ReplaceAll(tt, v, val)
			}
		}
	}
	return
}

func Loadwidget() (err error) {
	db := DB()
	if db != nil {
		asql := "select w.code,d.pageelementtype_id,d.url from widget w left join widget_dependency d on w.id=d.widget_id order by d.widget_id,d.ordinalposition"
		rows, e := db.Query(asql)
		if e == nil {
			var pageelementtype_id int
			var code, url sql.NullString
			for rows.Next() {
				e = rows.Scan(&code, &pageelementtype_id, &url)
				if e == nil {
					var w dependencyT
					var ok bool
					exists := false
					if w, ok = mapWidget[code.String]; ok {
						switch pageelementtype_id {
						case 1:
							exists, _ = In_array(url.String, w.CSS)
						case 2:
							exists, _ = In_array(url.String, w.JS)
						}
					}
					if !exists {
						switch pageelementtype_id {
						case 1:
							w.CSS = append(w.CSS, url.String)
						case 2:
							w.JS = append(w.JS, configurationVariablefilling(url.String))
						}
						mapWidget[code.String] = w
					}
				} else {
					err = e
					break
				}
			}
			rows.Close()
		} else {
			err = e
		}
	} else {
		err = errors.New("DB not connected!")
	}
	return
}

func IsBlackIP(r *http.Request) bool {
	return ipblacklist.Inblacklist(ipaddr.Ip2long(ipaddr.RemoteIp(r)))
}
func ReadIPblacklist() (err error) {
	if db != nil {
		asql := "select ipstart,ipterminus from ipblacklist order by ipstart"
		var rows *sql.Rows
		rows, err = db.Query(asql)
		if err == nil {
			for rows.Next() {
				var ipstart, ipterminus int64
				err = rows.Scan(&ipstart, &ipterminus)
				if err == nil {
					ipblacklist.AddSegment(uint32(ipstart), uint32(ipterminus))
				} else {
					break
				}
			}
			rows.Close()
		} else {
			ErrorLogger.Println(asql, err.Error())
		}
	}
	return
}
func AddOperation(r *http.Request, user_id int64, entity_id int, instance_id int64, action, brief string) {
	if db != nil {
		ip_address := ipaddr.RemoteIp(r)
		ipaddress := int64(ipaddr.Ip2long(ip_address))
		asql := "insert into user_operation(user_id,entity_id,instance_id,ip,action,brief,time_created) values(?,?,?,?,?,?," + SQL_now() + ")"
		_, e := db.Exec(asql, user_id, entity_id, instance_id, ipaddress, action, brief)
		if e != nil {
			ErrorLogger.Println(asql, user_id, entity_id, ipaddress, action, brief, e.Error())
		}
	}
}
func CountryPreferred(countrycode string) (country_id, preferredlanguage_id, preferredcurrency_id, preferreddateformat_id, preferredtimeformat_id int) {
	if db != nil {
		asql := "select id,preferredlanguage,preferredcurrency,preferreddateformat,preferredtimeformat from country"
		asql += " where code='" + strings.ToUpper(countrycode) + "'"
		row := db.QueryRow(asql)
		if row != nil {
			var preferredlanguage, preferredcurrency, preferreddateformat, preferredtimeformat sql.NullString
			e := row.Scan(&country_id, &preferredlanguage, &preferredcurrency, &preferreddateformat, &preferredtimeformat)
			if e == nil {
				preferredlanguage_id, _, _ = SplitIDCodeName(preferredlanguage.String)
				preferredcurrency_id, _, _ = SplitIDCodeName(preferredcurrency.String)
				preferreddateformat_id, _, _ = SplitIDCodeName(preferreddateformat.String)
				preferredtimeformat_id, _, _ = SplitIDCodeName(preferredtimeformat.String)
			}
		}
	}
	return
}
func Servicepreparation() {
	if db != nil {
		base_language, base_language_code, base_currency, launch_date := "", "", "", ""
		sys_accept_language, _ = ReadLanguageConfigurationFromFile()
		baselanguage_id := ReadBaselanguageConfigurationFromFile()
		if strings.Count(sys_accept_language, ",") > 0 {
			sys_acceptmultiplelanguage = true
		}
		asql := "select id,code,name from language where id in (" + sys_accept_language + ")"
		asql += " order by id"
		rows, e := db.Query(asql)
		if e == nil {
			ss := []string{}
			for rows.Next() {
				var id string
				var code, name sql.NullString
				err := rows.Scan(&id, &code, &name)
				if err == nil {
					lang := id + ":" + name.String + " [" + code.String + "]"
					if id == baselanguage_id {
						base_language = lang
						base_language_code = code.String
						switch code.String {
						case "zh":
							base_currency = "CNY"
						case "ja":
							base_currency = "JPY"
						case "de":
							base_currency = "EUR"
						case "fr":
							base_currency = "EUR"
						case "it":
							base_currency = "EUR"
						case "es":
							base_currency = "EUR"
						case "ko":
							base_currency = "KRW"
						case "ru":
							base_currency = "RUB"
						default:
							base_currency = "USD"
						}
					}
					ss = append(ss, lang)
				} else {
					ErrorLogger.Println(err.Error())
					break
				}
			}
			rows.Close()
			sys_accept_language = strings.Join(ss, ",")
		} else {
			ErrorLogger.Println(asql, e.Error())
		}
		asql = "update configuration set `value`=? where `key`='SYS_BASE_LANGUAGE'"
		_, e = db.Exec(asql, base_language)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
		asql = "update configuration set `value`=? where `key`='SYS_ACCEPT_LANGUAGE'"
		_, e = db.Exec(asql, sys_accept_language)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
		if len(base_currency) > 0 {
			asql = "select id,code from currency where code=?"
			row := db.QueryRow(asql, base_currency)
			if row != nil {
				var id string
				var code sql.NullString
				if row.Scan(&id, &code) == nil {
					base_currency = id + "." + code.String
					asql = "update configuration set `value`=? where `key`='SYS_BASE_CURRENCY'"
					_, e = db.Exec(asql, base_currency)
					if e != nil {
						ErrorLogger.Println(asql, e.Error())
					}
				}
			}
		}
		switch DB_type {
		case SQLite:
			if base_language_code == "zh" {
				launch_date = "strftime('%Y年%m月%d日')"
			} else {
				launch_date = "strftime('%m-%d-%Y')"
			}
		case MySQL:
			if base_language_code == "zh" {
				launch_date = "DATE_FORMAT(now(),'%Y年%c月%e日')"
			} else {
				launch_date = "DATE_FORMAT(now(),'%b %d %Y')"
			}
		}
		asql = "update configuration set `value`=" + launch_date + " where `key`='SITE_LAUNCH_DATE'"
		_, e = db.Exec(asql)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
		open := ""
		filename := filepath.Join(dirRun, "configure", "registration")
		if IsExists(filename) {
			b, err := ioutil.ReadFile(filename)
			if err == nil {
				open = string(b)
			}
		}
		if len(open) == 0 {
			open = "0"
		}
		asql = "update configuration set `value`=? where `key`='SITE_OPEN_REGISTRATION'"
		_, e = db.Exec(asql, open)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
		logowidth, headerheight := "76", "61"
		logofile := filepath.Join(ImgPath(), "logo_default.jpg")
		w, h := ImagePixelsWidthHeight(logofile)
		if w > 0 {
			logowidth = strconv.Itoa(w)
		}
		titlefile := filepath.Join(ImgPath(), "title_default.jpg")
		_, ht := ImagePixelsWidthHeight(titlefile)
		if ht > h {
			h = ht
		}
		if h > 0 {
			headerheight = strconv.Itoa(h + 2)
		}
		asql = "update configuration set `value`=? where `key`='UI_LOGO_WIDTH'"
		_, e = db.Exec(asql, logowidth)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
		asql = "update configuration set `value`=? where `key`='UI_HEADER_HEIGHT'"
		_, e = db.Exec(asql, headerheight)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
		desktoplogowidth, desktopheaderheight := "76", "38"
		logofile = filepath.Join(ImgPath(), "logo_desktop.jpg")
		w, h = ImagePixelsWidthHeight(logofile)
		if w > 0 {
			desktoplogowidth = strconv.Itoa(w)
		}
		titlefile = filepath.Join(ImgPath(), "title_desktop.jpg")
		_, ht = ImagePixelsWidthHeight(titlefile)
		if ht > h {
			h = ht
		}
		if h > 0 {
			desktopheaderheight = strconv.Itoa(h + 2)
		}
		asql = "update configuration set `value`=? where `key`='UI_DESKTOPLOGO_WIDTH'"
		_, e = db.Exec(asql, desktoplogowidth)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
		asql = "update configuration set `value`=? where `key`='UI_DESKTOPHEADER_HEIGHT'"
		_, e = db.Exec(asql, desktopheaderheight)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
		managerheaderheight := "38"
		logomanager := filepath.Join(ImgPath(), "logo_manager.jpg")
		_, h = ImagePixelsWidthHeight(logomanager)
		titlemanager := filepath.Join(ImgPath(), "title_manager.jpg")
		_, ht = ImagePixelsWidthHeight(titlemanager)
		if ht > h {
			h = ht
		}
		if h > 0 {
			managerheaderheight = strconv.Itoa(h + 2)
		}
		asql = "update configuration set `value`=? where `key`='UI_MANAGERHEADER_HEIGHT'"
		_, e = db.Exec(asql, managerheaderheight)
		if e != nil {
			ErrorLogger.Println(asql, e.Error())
		}
	}
}

func WriteEntityRecord(object_identifier string, instance_id int64, nml_fields,
	ins_clauses, nml_clauses []string, nml_values []interface{},
	adaptive_fields map[string][]string, adaptive_values map[string][]interface{}) (ninstance_id int64, err error) {
	var normal_fields, insert_clauses, normal_clauses []string
	var normal_values []interface{}
	n := len(nml_fields)
	for i := 0; i < n; i++ {
		field := nml_fields[i]
		exists, _ := In_array(field, normal_fields)
		if !exists {
			normal_fields = append(normal_fields, field)
			if field == "time_created" || field == "time_updated" {
				//ignore
			} else {
				//interface conversion: interface {} is int64, not string
				val, ok := nml_values[i].(string)
				if ok && val == "_NOW_" {
					insert_clauses = append(insert_clauses, SQL_now())
					normal_clauses = append(normal_clauses, SQL_now())
				} else {
					insert_clauses = append(insert_clauses, ins_clauses[i])
					normal_clauses = append(normal_clauses, nml_clauses[i])
					normal_values = append(normal_values, nml_values[i])
				}
			}
		}
	}
	if db != nil {
		err = nil
		var ct = 0
		if instance_id > 0 { //for object_extension
			csql := "select count(1) from " + Backquote(object_identifier) + " where id=?"
			row := db.QueryRow(csql, instance_id)
			if row != nil {
				row.Scan(&ct)
			}
		}

		n := len(normal_fields)
		if instance_id == 0 || ct == 0 {
			//fmt.Println("AAA", n)
			asql := "insert into " + Backquote(object_identifier) + "("
			for i := 0; i < n; i++ {
				asql += Backquote(normal_fields[i]) + ","
			}
			asql += "time_created,time_updated) values("
			for i := 0; i < n; i++ {
				asql += insert_clauses[i] + ","
			}
			asql += SQL_now() + "," + SQL_now() + ")"
			//fmt.Println("database.go:1182", asql, normal_values)
			rslt, e := db.Exec(asql, normal_values...)
			if e == nil {
				ninstance_id, err = rslt.LastInsertId()
			} else {
				err = errors.New("insert into error:" + object_identifier + ":" + e.Error() + " " + asql)
			}
			//fmt.Println("BBB")
		} else {
			vvalues := append(normal_values, interface{}(instance_id))
			asql := "update " + Backquote(object_identifier) + " set "
			for i := 0; i < n; i++ {
				asql += Backquote(normal_fields[i]) + "=" + normal_clauses[i] + ","
			}
			asql += "time_updated=" + SQL_now() + " where id=?"
			//fmt.Println(asql, vvalues)
			_, err = db.Exec(asql, vvalues...)
			ninstance_id = instance_id
		}

		if err == nil {
			if SysAcceptMultipleLanguage() {
				for language_id, fields := range adaptive_fields {
					if values, ok := adaptive_values[language_id]; ok {
						var instance_languages_id int64 = 0
						if instance_id > 0 {
							asql := "select `id` from " + Backquote(object_identifier+"_languages")
							asql += " where " + Backquote(object_identifier+"_id") + "=? and `language_id`=?"
							row := db.QueryRow(asql, instance_id, language_id)
							if row != nil {
								if row.Scan(&instance_languages_id) != nil {
									ErrorLogger.Println(asql, "scan id error")
								}
							}
						}
						if instance_languages_id == 0 {
							roadmap := strings.Split(object_identifier, "_")
							nroadmaps := len(roadmap)
							if nroadmaps > 1 {
								for i := 0; i < nroadmaps-1; i++ {
									rm_id := strings.Join(roadmap[0:i+1], "_") + "_id"
									exists, idx := In_array(rm_id, nml_fields)
									if exists {
										fields = append(fields, rm_id)
										values = append(values, nml_values[idx])
									}
								}
							}
							nfields := len(fields)
							asql := "insert into  " + Backquote(object_identifier+"_languages") + "(" + Backquote(object_identifier+"_id") + ","
							asql += Backquote("language_id") + "," + Backquote("language_tag")
							for i := 0; i < nfields; i++ {
								asql += "," + Backquote(fields[i])
							}
							asql += ",time_created,time_updated) values(?,?,?" + strings.Repeat(",?", nfields) + "," + SQL_now() + "," + SQL_now() + ")"

							v := make([]interface{}, 3)
							v[0] = interface{}(ninstance_id)
							v[1] = interface{}(language_id)
							v[2] = interface{}(mapLangTag[language_id])
							vv := append(v, values...)
							rslt, e := db.Exec(asql, vv...)
							//fmt.Println(asql, vv)
							if e == nil {
								instance_languages_id, err = rslt.LastInsertId()
								if err != nil {
									err = errors.New("LastInsertId:" + err.Error() + " " + asql)
								}
							} else {
								err = errors.New("insert into error:" + object_identifier + "_languages:" + e.Error())
								//fmt.Println(asql)
							}
						} else {
							nfields := len(fields)
							asql := "update " + Backquote(object_identifier+"_languages") + " set time_updated=" + SQL_now()
							for i := 0; i < nfields; i++ {
								asql += "," + Backquote(fields[i]) + "=?"
							}
							asql += " where id=?"
							vv := append(values, interface{}(instance_languages_id))
							_, err = db.Exec(asql, vv...)
							if err != nil {
								err = errors.New("update error:" + err.Error() + asql)
							}
						}
						if err != nil {
							break
						}
					}
				}
				//fmt.Println("CCC")
			}
		}
	}
	return
}

/*	normal_fields := []string{"value"}
	normal_values := make([]interface{}, 1)
	normal_values[0] = interface{}("0")
	adaptive_fields := []string{"title", "description"}
	adaptive_values := []interface{}{}
	adaptive_values = append(adaptive_values, interface{}("=1="))
	adaptive_values = append(adaptive_values, interface{}("=2="))
	base.WriteMultiLanguageRecord("configuration", "fr", 4, normal_fields, normal_values, adaptive_fields, adaptive_values)*/
//one language record, now use for "configuration"
//need extend roadmap like 'WriteEntityRecord'
func WriteMultiLanguageRecord(object_identifier, language string, instance_id int64,
	normal_fields []string, normal_values []interface{}, adaptive_fields []string, adaptive_values []interface{}) (instance_languages_id int64, err error) {
	/*fmt.Println("WriteMultiLanguageRecord:---")
	fmt.Println("normal_fields:", normal_fields)
	fmt.Println("normal_values:", normal_values)
	fmt.Println("adaptive_fields:", adaptive_fields)
	fmt.Println("adaptive_values:", adaptive_values)*/
	if db != nil {
		baselanguage_id := BaseLanguage_id()
		language_id := baselanguage_id
		if len(language) > 0 {
			language_id = Language_id(language)
		}
		fields := normal_fields
		values := normal_values
		if language_id == baselanguage_id {
			fields = append(fields, adaptive_fields...)
			values = append(values, adaptive_values...)
		}
		n := len(fields)
		if instance_id == 0 {
			asql := "insert into " + Backquote(object_identifier) + "("
			for i := 0; i < n; i++ {
				asql += Backquote(fields[i]) + ","
			}
			asql += "time_created,time_updated) values(" + strings.Repeat("?,", n) + SQL_now() + "," + SQL_now() + ")"
			rslt, e := db.Exec(asql, values...)
			if e == nil {
				id, e := rslt.LastInsertId()
				if e == nil {
					if SysAcceptMultipleLanguage() {
						//fmt.Println("database.go:1301")
						asql = "insert into  " + Backquote(object_identifier+"_languages") + "(" + Backquote(object_identifier+"_id") + ","
						asql += Backquote("language_id") + "," + Backquote("language_tag")
						n = len(adaptive_fields)
						for i := 0; i < n; i++ {
							asql += "," + Backquote(adaptive_fields[i])
						}
						asql += ") values(?,?" + strings.Repeat(",?", n) + ")"
						v := make([]interface{}, 3)
						v[0] = interface{}(id)
						v[1] = interface{}(language_id)
						v[2] = interface{}(mapLangTag[language_id])
						vv := append(v, adaptive_values...)
						//fmt.Println(asql)
						_, e = db.Exec(asql, vv...)
						if e != nil {
							err = errors.New("insert into error2:" + e.Error())
							ErrorLogger.Println(asql, e.Error())
						}
					}
				} else {
					ErrorLogger.Println("LastInsertId", e.Error())
				}
			} else {
				err = errors.New("insert into error2:" + e.Error())
				ErrorLogger.Println(asql, err.Error())
			}
		} else {
			values = append(values, interface{}(instance_id))
			asql := "update " + Backquote(object_identifier) + " set "
			for i := 0; i < n; i++ {
				asql += Backquote(fields[i]) + "=?,"
			}
			asql += "time_updated=" + SQL_now() + " where id=?"
			_, e := db.Exec(asql, values...)
			if e != nil {
				err = errors.New("update error")
				ErrorLogger.Println(asql, e.Error())
			} else {
				if SysAcceptMultipleLanguage() {
					n = len(adaptive_fields)
					if n > 0 {
						instance_languages_id = 0
						asql = "select `id` from " + Backquote(object_identifier+"_languages")
						asql += " where " + Backquote(object_identifier+"_id") + "=? and `language_id`=?"
						row := db.QueryRow(asql, instance_id, language_id)
						if row != nil {
							if row.Scan(&instance_languages_id) != nil {
								ErrorLogger.Println(asql, "scan id error")
							}
						}
						//fmt.Println("instance_languages_id:", instance_languages_id)
						if instance_languages_id == 0 {
							//fmt.Println("database.go:1354")
							asql = "insert into  " + Backquote(object_identifier+"_languages") + "(" + Backquote(object_identifier+"_id") + ","
							asql += Backquote("language_id") + "," + Backquote("language_tag")
							for i := 0; i < n; i++ {
								asql += "," + Backquote(adaptive_fields[i])
							}
							asql += ",time_updated) values(?,?,?" + strings.Repeat(",?", n) + "," + SQL_now() + ")"

							v := make([]interface{}, 3)
							v[0] = interface{}(instance_id)
							v[1] = interface{}(language_id)
							v[2] = interface{}(mapLangTag[language_id])
							vv := append(v, adaptive_values...)
							//fmt.Println(asql)
							rslt, e := db.Exec(asql, vv...)
							if e == nil {
								instance_languages_id, err = rslt.LastInsertId()
								if err != nil {
									ErrorLogger.Println("LastInsertId", e.Error())
								}
							} else {
								err = errors.New("insert into error")
								ErrorLogger.Println(asql, e.Error())
							}
						} else {
							asql = "update " + Backquote(object_identifier+"_languages") + " set time_updated=" + SQL_now()
							for i := 0; i < n; i++ {
								asql += "," + Backquote(adaptive_fields[i]) + "=?"
							}
							asql += " where id=?"
							vv := append(adaptive_values, interface{}(instance_languages_id))
							_, e = db.Exec(asql, vv...)
							//fmt.Println(asql, vv)
							if e != nil {
								err = errors.New("update error")
								ErrorLogger.Println(asql, e.Error())
							}
						}
					}
				}
			}
		}
	} else {
		err = errors.New("DB not connected!")
	}
	return
}

//func EmptyRecursively(identifier string) {
//->entity
// SQL_emptytable(tablename string) (err error)
//}

func RmvInstanceRecursively(object_identifier string, instance_id int64) (nn int) { //recursively
	if db != nil {
		nn = 0
		r_ids := []int64{} //recursive
		l_ids := []string{strconv.FormatInt(instance_id, 10)}
		asql := "select id,isleaf from " + Backquote(object_identifier) + " where parentid=?"
		rows, e := db.Query(asql, instance_id)
		if e == nil {
			var id int64
			var isleaf int
			for rows.Next() {
				e = rows.Scan(&id, &isleaf)
				if e == nil {
					if isleaf == 1 {
						l_ids = append(l_ids, strconv.FormatInt(id, 10))
					} else {
						r_ids = append(r_ids, id)
					}
				} else {
					ErrorLogger.Println(e.Error())
				}
			}
			rows.Close()
		}
		nleaves := len(l_ids)
		if nleaves > 0 {
			asql = "delete from " + Backquote(object_identifier) + " where id in (" + strings.Join(l_ids, ",") + ")"
			_, e := db.Exec(asql)
			if SysAcceptMultipleLanguage() {
				asql = "delete from " + Backquote(object_identifier+"_languages") + " where " + Backquote(object_identifier+"_id") + " in (" + strings.Join(l_ids, ",") + ")"
				_, e = db.Exec(asql)
			}
			if e != nil {
				ErrorLogger.Println(asql, e.Error())
			}
			nn += nleaves
		}
		n := len(r_ids)
		for i := 0; i < n; i++ {
			nn += RmvInstanceRecursively(object_identifier, r_ids[i])
		}
	}
	return
}

func InstanceRetrieve(tablename, unique_property, unique_value string) (instance_id int64) {
	if db != nil && len(unique_value) > 0 {
		asql := "select id from " + tablename + " where `" + unique_property + "`=?"
		row := db.QueryRow(asql, unique_value)
		if row != nil {
			e := row.Scan(&instance_id)
			if e != nil && !NoRowsError(e) {
				ErrorLogger.Println(asql, e.Error())
			}
		}
	}
	return
}

func InstanceRetrieveEx(tablename, unique_property, unique_value, extra_property string) (instance_id int64, extra_value string) {
	if db != nil && len(unique_value) > 0 {
		asql := "select id," + extra_property + " from " + tablename + " where " + unique_property + "=?"
		row := db.QueryRow(asql, unique_value)
		if row != nil {
			var e_v sql.NullString
			e := row.Scan(&instance_id, &e_v)
			if e != nil && !NoRowsError(e) {
				ErrorLogger.Println(asql, e.Error())
			} else {
				extra_value = e_v.String
			}
		}
	}
	return
}

func InstancePropertyRetrieve(tablename, property string, instance_id int64) (thevalue string) {
	if db != nil {
		asql := "select " + property + " from " + tablename + " where id=?"
		row := db.QueryRow(asql, instance_id)
		if row != nil {
			var v sql.NullString
			e := row.Scan(&v)
			if e != nil && !NoRowsError(e) {
				ErrorLogger.Println(asql, e.Error())
			} else {
				thevalue = v.String
			}
		}
	}
	return
}

func Instance2PropertyRetrieve(tablename, property1, property2 string, instance_id int64) (thevalue1, thevalue2 string) {
	if db != nil {
		asql := "select " + property1 + "," + property2 + " from " + tablename + " where id=?"
		row := db.QueryRow(asql, instance_id)
		if row != nil {
			var v1, v2 sql.NullString
			e := row.Scan(&v1, &v2)
			if e != nil && !NoRowsError(e) {
				ErrorLogger.Println(asql, e.Error())
			} else {
				thevalue1 = v1.String
				thevalue2 = v2.String
			}
		}
	}
	return
}

func Instance3PropertyRetrieve(tablename, property1, property2, property3 string, instance_id int64) (thevalue1, thevalue2, thevalue3 string) {
	if db != nil {
		asql := "select " + property1 + "," + property2 + "," + property3 + " from " + tablename + " where id=?"
		row := db.QueryRow(asql, instance_id)
		if row != nil {
			var v1, v2, v3 sql.NullString
			e := row.Scan(&v1, &v2, &v3)
			if e != nil && !NoRowsError(e) {
				ErrorLogger.Println(asql, e.Error())
			} else {
				thevalue1 = v1.String
				thevalue2 = v2.String
				thevalue3 = v3.String
			}
		}
	}
	return
}

func Instance4PropertyRetrieve(tablename, property1, property2, property3, property4 string, instance_id int64) (thevalue1, thevalue2, thevalue3, thevalue4 string) {
	if db != nil {
		asql := "select " + property1 + "," + property2 + "," + property3 + "," + property4 + " from " + tablename + " where id=?"
		row := db.QueryRow(asql, instance_id)
		if row != nil {
			var v1, v2, v3, v4 sql.NullString
			e := row.Scan(&v1, &v2, &v3, &v4)
			if e != nil && !NoRowsError(e) {
				ErrorLogger.Println(asql, e.Error())
			} else {
				thevalue1 = v1.String
				thevalue2 = v2.String
				thevalue3 = v3.String
				thevalue4 = v4.String
			}
		}
	}
	return
}

// properties: abc,def,ghi
func InstanceMultiPropertiesRetrieve(tablename, properties string, instance_id int64) (values []string) {
	if db != nil {
		asql := "select " + properties + " from " + tablename + " where id=?"
		row := db.QueryRow(asql, instance_id)
		if row != nil {
			m := strings.Count(properties, ",") + 1
			vals := make([]sql.NullString, m)
			cans := make([]interface{}, m)
			for i := 0; i < m; i++ {
				cans[i] = &vals[i]
			}
			if row.Scan(cans...) == nil {
				for i := 0; i < m; i++ {
					values = append(values, vals[i].String)
				}
			}
		}
	}
	return
}

func InstancePropertyUpdate(tablename, extra_property, extra_value string, instance_id int64) (e error) {
	if instance_id > 0 {
		if db != nil {
			asql := "update " + tablename + " set " + extra_property + "=? where id=?"
			_, e = db.Exec(asql, extra_value, instance_id)
		}
	}
	return
}

func InstancePropertyUpdateEx(tablename, unique_property, unique_value, extra_property, extra_value string) (e error) {
	instance_id := InstanceRetrieve(tablename, unique_property, unique_value)
	e = InstancePropertyUpdate(tablename, extra_property, extra_value, instance_id)
	return
}

// int property
func InstanceCount_int(objectidentifier, property_name string, property_value int) (count int) {
	if db != nil {
		tablename := strings.ReplaceAll(objectidentifier, ".", "_")
		asql := "select count(1) from " + tablename
		asql += " where " + property_name + "=" + strconv.Itoa(property_value)
		row := db.QueryRow(asql)
		if row != nil {
			row.Scan(&count)
		}
	}
	return
}

func InstanceIDs(tablename, property_name, property_value string) (iids []int64) {
	if db != nil && len(property_value) > 0 {
		var e error
		var rows *sql.Rows
		asql := "select id from " + tablename + " where " + property_name + "="
		if IsTime19(property_value) {
			asql += SQL_dateformat19(property_value)
			rows, e = db.Query(asql)
		} else {
			asql += "?"
			rows, e = db.Query(asql, property_value)
		}
		if e == nil {
			var id int64
			for rows.Next() {
				e = rows.Scan(&id)
				if e == nil {
					iids = append(iids, id)
				} else {
					ErrorLogger.Println(e.Error())
				}
			}
			rows.Close()
		}
	}
	return
}

func InstanceIDsEx(tablename, property_name, property_value, extra_property string, limit int) (iids []int64, vals []string) {
	if db != nil && len(property_value) > 0 && len(extra_property) > 0 {
		var e error
		var rows *sql.Rows
		limit_sql := ""
		if limit > 0 {
			limit_sql = " limit " + strconv.Itoa(limit)
		}
		asql := "select id," + extra_property + " from " + tablename + " where " + property_name + "="
		if IsTime19(property_value) {
			asql += SQL_dateformat19(property_value) + limit_sql
			rows, e = db.Query(asql)
		} else {
			asql += "?" + limit_sql
			rows, e = db.Query(asql, property_value)
		}
		if e == nil {
			var id int64
			var eval sql.NullString
			for rows.Next() {
				e = rows.Scan(&id, &eval)
				if e == nil {
					iids = append(iids, id)
					vals = append(vals, eval.String)
				} else {
					ErrorLogger.Println(e.Error())
				}
			}
			rows.Close()
		}
	}
	return
}

func InstanceIDsAnB(tablename, property_name, property_value, extra_aproperty, extra_bproperty string) (iids []int64, avals, bvals []string) {
	if db != nil && len(property_value) > 0 && len(extra_aproperty) > 0 && len(extra_bproperty) > 0 {
		var e error
		var rows *sql.Rows
		asql := "select id," + extra_aproperty + "," + extra_bproperty + " from " + tablename + " where " + property_name + "="
		if IsTime19(property_value) {
			asql += SQL_dateformat19(property_value)
			rows, e = db.Query(asql)
		} else {
			asql += "?"
			rows, e = db.Query(asql, property_value)
		}
		if e == nil {
			var id int64
			var aval, bval sql.NullString
			for rows.Next() {
				e = rows.Scan(&id, &aval, &bval)
				if e == nil {
					iids = append(iids, id)
					avals = append(avals, aval.String)
					bvals = append(bvals, bval.String)
				} else {
					ErrorLogger.Println(e.Error())
				}
			}
			rows.Close()
		}
	}
	return
}

func Emptytable(tablename string) (err error) {
	if db != nil {
		_, err = db.Exec("delete from `" + tablename + "`")
	} else {
		err = errors.New("DB not connected!")
	}
	return
}

func Tableisempty(tablename string) (flag bool, err error) {
	flag = false
	count := 0
	count, err = Tablecount(tablename)
	if err == nil {
		flag = (count == 0)
	}
	return
}

func Tablecount(tablename string) (count int, err error) {
	if db != nil {
		row := db.QueryRow("select count(1) from `" + tablename + "`")
		if row != nil {
			err = row.Scan(&count)
		}
	} else {
		err = errors.New("DB not connected!")
	}
	return
}

func Tablecountex(tablename, where string) (count int, err error) {
	if db != nil {
		asql := "select count(1) from `" + tablename + "`"
		if len(where) > 0 {
			asql += " where " + where
		}
		row := db.QueryRow(asql)
		if row != nil {
			err = row.Scan(&count)
		}
	} else {
		err = errors.New("DB not connected!")
	}
	return
}
func InstanceCount(asql string) (count int64) { //asql: select count(1) from user
	if db != nil && len(asql) > 0 {
		row := db.QueryRow(asql)
		if row != nil {
			e := row.Scan(&count)
			if e != nil && !NoRowsError(e) {
				ErrorLogger.Println(asql, e.Error())
			}
		}
	}
	return
}

func InstanceCountEx(tablename, property_name, property_value string) (count int64) {
	if db != nil && len(property_name) > 0 && len(property_value) > 0 {
		asql := `select count(1) from ` + tablename + ` where ` + property_name + `="` + property_value + `"`
		row := db.QueryRow(asql)
		if row != nil {
			e := row.Scan(&count)
			if e != nil && !NoRowsError(e) {
				ErrorLogger.Println(asql, e.Error())
			}
		}
	}
	return
}

func GetRoadmapids(identifier, subentity string, instance_id int64) (roadmapids []int64) {
	if len(subentity) > 0 {
		roadmap := CombineRoadmap(identifier, subentity)
		rmrm := []string{}
		n := len(roadmap) - 1
		for i := 0; i < n; i++ {
			rmrm = append(rmrm, strings.Join(roadmap[:i+1], "_")+"_id")
		}
		if db != nil {
			asql := "select " + strings.Join(rmrm, ",") + " from " + strings.Join(roadmap, "_") + " where id=?"
			row := db.QueryRow(asql, instance_id)
			if row != nil {
				vals := make([]interface{}, n)
				for i := 0; i < n; i++ {
					vals[i] = new(int64)
				}
				if row.Scan(vals...) == nil {
					for i := 0; i < n; i++ {
						roadmapids = append(roadmapids, *(vals[i].(*int64)))
					}
				}
			}
		}
	} else {
		roadmapids = append(roadmapids, instance_id)
	}
	return
}

func Parentids(tablename string, iid int) (parents []string) {
	cii := iid
	for cii > 0 {
		asql := "select parentid from " + tablename + " where id=?"
		row := DB().QueryRow(asql, cii)
		if row != nil {
			var parent_id int
			e := row.Scan(&parent_id)
			if e == nil {
				if parent_id > 0 { //first hierarchy not read
					parents = append(parents, strconv.Itoa(parent_id))
				}
				cii = parent_id
			}
		}
	}
	return
}

func SetUserShoppingcart(db *sql.DB, user_id int64) (shoppingcart_id int64, e error) {
	asql := "select id from user_shoppingcart where id=?"
	row := db.QueryRow(asql, user_id)
	if row != nil {
		e = row.Scan(&shoppingcart_id)
		if e == nil || NoRowsError(e) {
			if shoppingcart_id == 0 {
				asql = "insert into user_shoppingcart(id,time_created,time_updated)"
				asql += " values(?," + SQL_now() + "," + SQL_now() + ")"
				var rslt sql.Result
				rslt, e = db.Exec(asql, user_id)
				if e == nil {
					shoppingcart_id, e = rslt.LastInsertId()
				}
			}
		}
	}
	return
}

func UpdateUserInferiors(db *sql.DB, superior_id int64) (err error) {
	if superior_id > 0 {
		inferiors := 0
		asql := "select IFNULL(count(1),0) from user where superiorid=?"
		row := db.QueryRow(asql, superior_id)
		if row != nil {
			row.Scan(&inferiors)
		}
		if inferiors > 0 {
			asql = "update user set inferiors=? where id=?"
			_, err = db.Exec(asql, inferiors, superior_id)
		}
	}
	return
}

func RecordUserIP(db *sql.DB, user_id, ipaddress int64) {
	var userip_id int64 = 0
	asql := "select id from user_ipaddress where user_id=? and ip=?"
	row := db.QueryRow(asql, user_id, ipaddress)
	if row != nil {
		row.Scan(&userip_id)
		AccessLogger.Println(asql, user_id, ipaddress)
		AccessLogger.Println("userip_id:", userip_id)
	}
	if userip_id == 0 {
		asql = "insert into user_ipaddress(user_id,ip,time_fired) values(?,?," + SQL_now() + ")"
		_, err := db.Exec(asql, user_id, ipaddress)
		if err != nil {
			ErrorLogger.Println(asql, err.Error())
		} else {
			AccessLogger.Println(asql, user_id, ipaddress)
		}
	} else {
		asql = "update user_ipaddress set time_fired=" + SQL_now() + " where id=?"
		_, err := db.Exec(asql, userip_id)
		if err != nil {
			ErrorLogger.Println(asql, err.Error())
		} else {
			AccessLogger.Println(asql, userip_id)
		}
	}
}

func RetrieveCounter(itemkey string) (val int) {
	_, itemval := InstanceRetrieveEx("counter", "itemkey", itemkey, "itemval")
	val = Str2int(itemval)
	return
}

func IncCounter(itemkey string) (e error) {
	counter_id, itemval := InstanceRetrieveEx("counter", "itemkey", itemkey, "itemval")
	if counter_id == 0 {
		if db != nil {
			asql := "insert into counter(itemkey,itemval) values(?,?)"
			_, e = db.Exec(asql, itemkey, "1")
		}
	} else {
		val := Str2int(itemval) + 1
		e = InstancePropertyUpdate("counter", "itemval", strconv.Itoa(val), counter_id)
	}
	return
}

// count(1),sum(amount)
//
//identifier:user,subentity:wxvisit,target:count(1)
func Aggregate(identifier, subentity, target, definition string, mapQuery map[string]string, limit string) (r_val string, e error) {
	roadmap := CombineRoadmap(identifier, subentity)
	tablename := strings.Join(roadmap, "_")
	msg := ""
	wheree := []string{}
	asql := "select " + target + " from " + tablename
	if len(mapQuery) > 0 { //query is the limitation of the data range.
		if gjson.Valid(definition) {
			o_definition := gjson.Parse(definition)
			if len(subentity) > 0 {
				o_definition = o_definition.Get(subentity)
			}
			for key, val := range mapQuery {
				prop := o_definition.Get(key)
				if prop.Exists() {
					wheree = append(wheree, KV2expression(key, val, prop.Get("type").String(), ""))
				} else {
					msg = "query [" + key + "] not exists in object definition"
					break
				}
			}
		} else {
			msg = "definition not valid!"
		}
	}
	if len(limit) > 0 {
		wheree = append(wheree, limit)
	}
	if len(wheree) > 0 {
		asql += " where " + strings.Join(wheree, " and ")
	}
	if len(msg) == 0 {
		if db != nil {
			AccessLogger.Println(asql)
			//fmt.Println(asql)
			row := db.QueryRow(asql)
			if row != nil {
				var v sql.NullString
				e = row.Scan(&v)
				if e == nil {
					r_val = v.String
				}
			}
		}
	} else {
		e = errors.New(msg)
	}
	return
}
