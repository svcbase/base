package base

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type metainfoT struct {
	Comment   string
	Signature string
}

type Metadata struct {
	Filenames []string
	mapInfo   map[string]metainfoT
}

func (md *Metadata) Loadfromfile() (e error) {
	if len(md.Filenames) > 0 {
		md.Filenames = make([]string, 0)
	}
	md.mapInfo = make(map[string]metainfoT)
	filename := filepath.Join(dirRes, "res", "meta.data")
	if IsExists(filename) {
		file, ee := os.Open(filename)
		if ee != nil {
			e = errors.New("failed to open file: " + filename + " " + ee.Error())
			return
		} else {
			reader := bufio.NewReader(file)
			eof := false
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						e = errors.New("failed to finish reading the file:" + " " + filename + " " + err.Error())
						return
					} //here reach the last line
					eof = true
				}
				ln := strings.TrimLeft(line, "\r\t ")
				if !strings.HasPrefix(ln, "#") && !strings.HasPrefix(ln, "-- ") {
					ln = TrimBLANK(line)
					if len(ln) > 0 {
						fname := ""
						mm := strings.Split(ln, ";")
						if len(mm) > 0 {
							fname = mm[0]
							filename := filepath.Join(dirRes, "res", fname)
							if IsExists(filename) {
								bb, err := ioutil.ReadFile(filename)
								if err == nil {
									signature := StrMD5(string(bb))
									comment := ""
									if len(mm) > 1 {
										comment = mm[1]
									}
									md.Filenames = append(md.Filenames, fname)
									md.mapInfo[fname] = metainfoT{comment, signature}
								} else {
									e = errors.New(filename + " md5 " + err.Error())
									break
								}
							} else {
								fmt.Println(ln)
								e = errors.New(filename + " not exists!")
								break
							}
						}
					}
				}
				if eof {
					break
				}
			}
			file.Close()
		}
	} else {
		e = errors.New(filename + " not exists!")
	}
	return
}

func (md *Metadata) Signature() (signature string) {
	filenames := []string{}
	for _, fname := range md.Filenames { //remain meta.data order
		filenames = append(filenames, fname)
	}
	sort.Strings(filenames)
	ss := ""
	for _, fname := range filenames {
		if info, ok := md.mapInfo[fname]; ok {
			ss += fname + info.Signature
		}
	}
	signature = StrMD5(ss)
	return
}

func (md *Metadata) Filecomment(fname string) (comment string) {
	if v, ok := md.mapInfo[fname]; ok {
		comment = v.Comment
	}
	return
}

func (md *Metadata) WriteSignature2DeepData() (err error) {
	db := DB()
	if db != nil {
		asql := "update deepdata set signature=?,time_updated=" + SQL_now()
		_, err = db.Exec(asql, md.Signature())
		if err == nil {
			asql = "update `configuration` set `valuestarting`= `value`"
			_, err = db.Exec(asql)
		}
	} else {
		err = errors.New("DB not ready!")
	}
	return
}

func (md *Metadata) Write2Database(db *sql.DB) (err error) {
	asql := "delete from metadata"
	_, err = db.Exec(asql)
	if err == nil {
		filenames := []string{}
		for _, fname := range md.Filenames { //remain meta.data order
			filenames = append(filenames, fname)
		}
		sort.Strings(filenames)
		for _, fname := range filenames {
			if info, ok := md.mapInfo[fname]; ok {
				//filename := filepath.Join(dirRes, "res", "meta.data")
				//tm := Filemodtime(filename)
				//use for datetime field
				extension := strings.Trim(filepath.Ext(fname), ".")
				asql := "insert into metadata(filename,fileextension,filecomment,signature,time_created,time_updated) "
				asql += "values(?,?,?,?," + SQL_now() + "," + SQL_now() + ")"
				_, err = db.Exec(asql, fname, extension, info.Comment, info.Signature)
				if err != nil {
					break
				}
			}
		}
	}
	return
}

func (md *Metadata) Write2DB() (err error) {
	db := DB()
	if db != nil {
		err = md.Write2Database(db)
	} else {
		err = errors.New("DB not ready!")
	}
	return
}

func (md *Metadata) UpgradedMetas() (metas []string, err error) {
	md.Loadfromfile()
	tablename := "metadata"
	if TableExists(tablename) {
		isempty := true
		isempty, err = Tableisempty(tablename)
		if err == nil {
			db := DB()
			if db != nil {
				for _, fname := range md.Filenames {
					skip := (isempty && strings.HasSuffix(fname, ".sql"))
					asql := "select signature from metadata where filename=?"
					row := db.QueryRow(asql, fname)
					if row != nil {
						var signature sql.NullString
						err = row.Scan(&signature)
						if err == nil {
							if info, ok := md.mapInfo[fname]; ok {
								if info.Signature != signature.String {
									if !skip {
										metas = append(metas, fname)
									}
								}
							}
						} else {
							if NoRowsError(err) {
								if !skip {
									metas = append(metas, fname)
								}
								err = nil
							}
						}
					}
				}
			} else {
				err = errors.New("DB not ready!")
			}
		}
	} else {
		for _, fname := range md.Filenames {
			if !strings.HasSuffix(fname, ".sql") { //skip .sql file
				metas = append(metas, fname)
			}
		}
	}
	return
}
