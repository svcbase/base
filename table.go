package base

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
)

type FieldinfoT struct {
	ctype        string
	notnull      bool
	defaultvalue string
	primarykey   bool
}

type TableInfoT struct {
	MapField   map[string]FieldinfoT
	MapIndexON map[string]string
}

/*
TABLE_CATALOG,TABLE_SCHEMA,TABLE_NAME,COLUMN_NAME,ORDINAL_POSITION,COLUMN_DEFAULT
IS_NULLABLE,DATA_TYPE,CHARACTER_MAXIMUM_LENGTH,CHARACTER_OCTET_LENGTH,NUMERIC_PRECISION
NUMERIC_SCALE,DATETIME_PRECISION,CHARACTER_SET_NAME,COLLATION_NAME,COLUMN_TYPE,COLUMN_KEY
EXTRA,PRIVILEGES,COLUMN_COMMENT
mysql> select column_name,data_type,column_type,is_nullable,column_default,column_key from information_schema.columns
    -> where table_schema='lsjc' and table_name='entity';
+-------------------+-----------+--------------+-------------+---------------------+------------+
| column_name       | data_type | column_type  | is_nullable | column_default      | column_key |
+-------------------+-----------+--------------+-------------+---------------------+------------+
| id                | int       | int(11)      | NO          | NULL                | PRI        |
| time_created      | datetime  | datetime     | YES         | 0000-01-01 00:00:00 |            |
| time_updated      | datetime  | datetime     | YES         | 0000-01-01 00:00:00 |            |
| code              | varchar   | varchar(64)  | YES         |                     | MUL        |
| name              | varchar   | varchar(255) | YES         |                     |            |
| description       | text      | text         | YES         | NULL                |            |
| enableflag        | int       | int(11)      | YES         | 1                   | MUL        |
| ordinalposition   | int       | int(11)      | YES         | 0                   | MUL        |
| authority         | varchar   | varchar(64)  | YES         | system              |            |
| creator           | varchar   | varchar(64)  | YES         |                     |            |
| thumbnail         | varchar   | varchar(96)  | YES         |                     |            |
| self_hierarchy    | int       | int(11)      | YES         | 0                   |            |
| self_cite         | int       | int(11)      | YES         | 0                   |            |
| multiple_language | int       | int(11)      | YES         | 0                   |            |
| metatype_id       | int       | int(11)      | YES         | 4                   | MUL        |
| definition        | text      | text         | YES         | NULL                |            |
| eventtrigger      | text      | text         | YES         | NULL                |            |
| remark            | text      | text         | YES         | NULL                |            |
+-------------------+-----------+--------------+-------------+---------------------+------------+*/
/*pragma table_info('entity')
cid,name,type,notnull,dflt_value,pk
0|id|INTEGER|1||1
1|time_created|datetime|0|'0000-01-01 00:00:00'|0
2|time_updated|datetime|0|'0000-01-01 00:00:00'|0
3|code|varchar(64)|0|''|0
4|name|varchar(255)|0|''|0
5|description|text|0||0
6|enableflag|INTEGER|0|'1'|0
7|ordinalposition|INTEGER|0|'0'|0
8|authority|varchar(64)|0|'system'|0
9|creator|varchar(64)|0|''|0
10|thumbnail|varchar(96)|0|''|0
11|self_hierarchy|INTEGER|0|'0'|0
12|self_cite|INTEGER|0|'0'|0
13|multiple_language|INTEGER|0|'0'|0
14|metatype_id|INTEGER|0|'4'|0
15|definition|text|0||0
16|eventtrigger|text|0||0
17|remark|text|0||0
alter table tablename ADD COLUMN column-name column-type
*/
func (ti *TableInfoT) ReadFields(tablename string) (err error) {
	ti.MapField = make(map[string]FieldinfoT)
	var rws *sql.Rows
	switch DB_type {
	case SQLite:
		asql := "pragma table_info('" + tablename + "')"
		rws, err = db.Query(asql)
		if err == nil {
			for rws.Next() {
				var id, nm, ty, nn, dv, pk sql.NullString //cid,name,type,notnull,dflt_value,pk
				err = rws.Scan(&id, &nm, &ty, &nn, &dv, &pk)
				if err == nil {
					ti.MapField[nm.String] = FieldinfoT{ty.String, nn.String == "1", dv.String, pk.String == "1"}
				} else {
					break
				}
			}
			rws.Close()
		}
	case MySQL:
		asql := "select column_name,column_type,is_nullable,column_default,column_key from information_schema.columns "
		asql += "where table_schema='" + DB_database + "' and table_name='" + tablename + "'"
		rws, err = db.Query(asql)
		if err == nil {
			for rws.Next() {
				var nm, ty, nn, dv, pk sql.NullString
				err = rws.Scan(&nm, &ty, &nn, &dv, &pk)
				if err == nil {
					t := strings.ReplaceAll(ty.String, "int(11)", "int")
					t = strings.ReplaceAll(t, "bigint(20)", "bigint")
					ti.MapField[nm.String] = FieldinfoT{t, nn.String == "NO", dv.String, pk.String == "PRI"}
				} else {
					break
				}
			}
			rws.Close()
		}
	}
	return
}

func (ti *TableInfoT) FieldExists(fieldname string) (flag bool) {
	_, flag = ti.MapField[fieldname]
	return
}

func (ti *TableInfoT) SameProperty(fieldname, propertySQL string) (flag bool) {
	if fi, ok := ti.MapField[fieldname]; ok {
		ff := "`" + fieldname + "` "
		ff += fi.ctype
		if fi.primarykey {
			ff += " PRIMARY KEY"
			ff += " NOT NULL"
		} else {
			switch DB_type {
			case SQLite:
				if len(fi.defaultvalue) > 0 {
					ff += " DEFAULT " + fi.defaultvalue //defaultvalue contain ''
				}
			case MySQL:
				if fi.defaultvalue != "NULL" {
					if len(fi.defaultvalue) > 0 {
						ff += " DEFAULT '" + fi.defaultvalue + "'"
					} else {
						if fi.ctype != "datetime" && fi.ctype != "mediumblob" {
							ff += " DEFAULT ''"
						}
					}
				}
			}
			if fi.notnull {
				ff += " NOT NULL"
			}
		}
		flag = (ff == propertySQL)
		if !flag {
			fmt.Println(ff, "=", propertySQL)
		}
	}
	return
}

/*sqlite:
select name,sql from sqlite_master where type='index' and tbl_name='entity';
CREATE INDEX `idx_entity_metatype_id` ON `entity`(`metatype_id`)
CREATE INDEX `idx_entity_ordinalposition` ON `entity`(`ordinalposition`)
CREATE INDEX `idx_entity_enableflag` ON `entity`(`enableflag`)
CREATE INDEX `idx_entity_code` ON `entity`(`code`)
CREATE INDEX `idx_user_recentappendix_u_s_u` ON `user_recentappendix`(`user_id`,`scene`,`time_updated` DESC)
MySQL:

*/
//DROP INDEX index_name

func (ti *TableInfoT) ReadIndexes(tablename string) (err error) {
	ti.MapIndexON = make(map[string]string)
	var rws *sql.Rows
	switch DB_type {
	case SQLite:
		asql := "select name,sql from sqlite_master where type='index' and tbl_name='" + tablename + "'"
		rws, err = db.Query(asql)
		if err == nil {
			for rws.Next() {
				var nm, ci sql.NullString
				err = rws.Scan(&nm, &ci)
				if err == nil {
					b := []byte(ci.String)
					i := bytes.IndexRune(b, '(')
					j := bytes.IndexRune(b, ')')
					if i > 0 && j > i {
						onfield := strings.ReplaceAll(string(b[i+1:j]), "`", "")
						onfield = strings.ReplaceAll(onfield, " ASC", "")
						onfield = strings.ReplaceAll(onfield, " DESC", "")
						onfield = strings.ReplaceAll(onfield, " asc", "")
						onfield = strings.ReplaceAll(onfield, " desc", "")
						onfield = strings.ReplaceAll(onfield, " ", "")
						ti.MapIndexON[onfield] = nm.String //map[`code`,`name` DESC]=idx_entity_code
					}
				} else {
					break
				}
			}
			rws.Close()
		}
	case MySQL:
		asql := "select index_name,column_name,collation from information_schema.statistics"
		asql += " where table_schema='" + DB_database + "' and table_name='" + tablename + "'"
		asql += " order by seq_in_index"
		rws, err = db.Query(asql)
		if err == nil {
			mapICC := make(map[string][]string)
			for rws.Next() {
				var index_name, column_name, collation sql.NullString
				err = rws.Scan(&index_name, &column_name, &collation)
				if err == nil {
					i_n := index_name.String
					if i_n != "PRIMARY" {
						fieldnames := mapICC[i_n]
						//fieldnames = append(fieldnames, "`"+column_name.String+"`")
						fieldnames = append(fieldnames, column_name.String)
						mapICC[i_n] = fieldnames
					}
				} else {
					break
				}
			}
			rws.Close()
			for i_n, fns := range mapICC {
				ti.MapIndexON[strings.Join(fns, ",")] = i_n
			}
		}
	}
	return
}

func (ti *TableInfoT) IndexExists(index_on_fields string) (flag bool) {
	i_o_f := strings.ReplaceAll(index_on_fields, "`", "")
	_, flag = ti.MapIndexON[i_o_f]
	return
}
