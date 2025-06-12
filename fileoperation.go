package base

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/svcbase/hierpath"
)

func MakeDir(dir string) {
	if !IsExists(dir) {
		os.Mkdir(dir, 0755) //os.ModeDir
	}
}

func IsExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func Filesize(filename string) (size int64) {
	file, err := os.Open(filename)
	if err == nil {
		stat, err := file.Stat()
		if err == nil {
			size = stat.Size()
		}
		file.Close()
	}
	return
}

func Filemodtime(filename string) (tm time.Time) {
	file, err := os.Open(filename)
	if err == nil {
		stat, err := file.Stat()
		if err == nil {
			tm = stat.ModTime()
		}
		file.Close()
	}
	return
}

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, errors.New(src + " is not a regular file")
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if err == nil {
		err = os.Remove(dst)
		if err != nil {
			return 0, err
		}
	}
	destination, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceFileStat.Mode())
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func AppendFile(mainfile, appendfile string) (err error) {
	var mfile, afile *os.File
	mfile, err = os.OpenFile(mainfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err == nil {
		afile, err = os.OpenFile(appendfile, os.O_RDONLY, 0600)
		if err == nil {
			_, err = mfile.Seek(0, os.SEEK_END)
			if err == nil {
				n := 0
				b := make([]byte, defaultBufSize)
				for {
					n, err = afile.Read(b)
					if err == nil {
						if n > 0 {
							_, err = mfile.Write(b[0:n])
						}
						if err != nil || n < defaultBufSize {
							break
						}
					} else {
						break
					}
				}
			}
			afile.Close()
		}
		mfile.Close()
	}
	return
}

func FileBase64ToString(filename string) string {
	var databuffer bytes.Buffer
	f, e := os.Stat(filename)
	if e == nil {
		w := base64.NewEncoder(base64.StdEncoding, &databuffer)
		size := f.Size()
		if size > 0 {
			fi, err := os.Open(filename)
			if err == nil {
				for {
					b := make([]byte, defaultBufSize)
					n, e := fi.Read(b)
					if e == nil {
						w.Write(b[0:n])
					} else {
						break
					}
				}
			}
			defer fi.Close()
		}
		w.Close()
	}
	return databuffer.String()
}

func Clean_dir(path string, beforehours int) (n int) {
	n = 0
	d, _ := time.ParseDuration(fmt.Sprintf("-%dh", beforehours))
	timepoint := time.Now().Add(d)

	dir, err := ioutil.ReadDir(path)
	if err == nil {
		for _, fi := range dir {
			if !fi.IsDir() {
				mt := fi.ModTime()
				if mt.Before(timepoint) {
					n++
					os.Remove(filepath.Join(path, fi.Name()))
				}
			}
		}
	}
	return
}

func CleanDirBySuffix(path, suffix string) (n int) {
	n = 0
	if IsExists(path) {
		dir, err := ioutil.ReadDir(path)
		if err == nil {
			for _, fi := range dir {
				if !fi.IsDir() {
					fname := fi.Name()
					if strings.HasSuffix(fname, suffix) {
						os.Remove(filepath.Join(path, fname))
						n++
					}
				}
			}
		}
	}
	return
}

func ReadDirBySuffix(path, suffix string) (files []string) {
	if IsExists(path) {
		dir, err := ioutil.ReadDir(path)
		if err == nil {
			for _, fi := range dir {
				if !fi.IsDir() {
					fname := fi.Name()
					if strings.HasSuffix(fname, suffix) {
						files = append(files, fname)
					}
				}
			}
		}
	}
	return
}

func Upload2AbsolutePath(upath string) (apath string) {
	if strings.HasPrefix(upath, "/u?n=") {
		b := []byte(upath)
		apath = hierpath.MappingHierarchicalPath(3, GetUploadsDir(), string(b[5:]))
	}
	return
}

func IllustrationAbsolutePath(ipath string) (apath string) {
	if strings.HasPrefix(ipath, "/i/") {
		b := []byte(ipath)
		fpath := strings.ReplaceAll(string(b[3:]), "/", string(os.PathSeparator))
		apath = filepath.Join(dirRun, "illustration", fpath)
	}
	return
}

func IllustrationPath() (apath string) {
	apath = filepath.Join(dirRun, "illustration")
	return
}

func TempPath() (apath string) {
	apath = filepath.Join(dirRun, "temp")
	return
}

func ImgPath() (apath string) {
	apath = filepath.Join(dirRes, "web", "img")
	return
}

func HasReadWritePermission(spath string) (flag bool) {
	flag = false
	file_info, err := os.Stat(spath)
	if err == nil {
		var rw uint32 = 384 //-rw- --- ---
		file_perm := file_info.Mode().Perm()
		flag = (uint32(file_perm&os.FileMode(rw)) == rw)
	}
	return
}

func Unzip(zipfile, dirOut string) (err error) {
	if filepath.Ext(zipfile) == ".zip" {
		MakeDir(dirOut)
		r, e := zip.OpenReader(zipfile)
		if e == nil {
			for _, innerFile := range r.File {
				fname := filepath.Join(dirOut, innerFile.Name)
				if innerFile.FileInfo().IsDir() {
					err = os.Mkdir(fname, 0755)
				} else {
					if err = os.MkdirAll(filepath.Dir(fname), 0755); err != nil {
						fmt.Println(fname, "*", err.Error())
						break
					} else {
						tm := innerFile.FileInfo().ModTime()
						w, e := os.Create(fname)
						if e != nil {
							err = errors.New("create: " + fname + " " + e.Error())
							break
						} else {
							defer w.Close()
							rc, e := innerFile.Open()
							if e != nil {
								err = errors.New("inner file open: " + innerFile.Name + " " + e.Error())
								break
							} else {
								defer rc.Close()
								_, err = io.Copy(w, rc)
								if err != nil {
									break
								} else {
									os.Chtimes(fname, tm, tm)
								}
							}
						}
					}
				}
			}
			r.Close()
		} else {
			err = errors.New("open: " + zipfile + " " + e.Error())
		}
	} else {
		err = errors.New("this is not a zip file: " + zipfile)
	}
	return
}

func ReadOneSQL(filename string, ipos int) (sqltxt, comment string, nextpos int, err error) {
	nextpos = -1 //over
	f, e := os.Stat(filename)
	if e == nil {
		size := int(f.Size())
		if ipos < size {
			var fi *os.File
			fi, err = os.Open(filename)
			if err == nil {
				bLine := bytes.NewBuffer([]byte{})
				b := make([]byte, defaultBufSize)
				istart := ipos
				bnext := true
				for bnext {
					_, err = fi.Seek(int64(istart), os.SEEK_SET) //func (f *File) Seek(offset int64, whence int) (ret int64, err error)
					if err == nil {
						n := 0
						n, err = fi.Read(b) //func (f *File) Read(b []byte) (n int, err error) {
						if err == nil {
							for i := 0; i < n; i++ {
								if b[i] == '\n' {
									if istart+i+1 < size {
										nextpos = istart + i + 1
									}
									bLine.Write(b[0:i])
									bnext = false
									break
								}
							}
							if bnext {
								if n == defaultBufSize {
									bLine.Write(b)
								} else {
									bLine.Write(b[0:n])
								}
								istart += n
								if istart >= size {
									bnext = false //nextpos=-1
								}
							}
						} else {
							bnext = false
						}
					} else {
						bnext = false
					}
				}
				fi.Close()
				sqltxt, comment = SplitSQL(bLine.String())
			}
		}
	} else {
		err = errors.New("SQL file not find, must be restart!")
	}
	return
}

func ReadNSQL(filename string, ipos, needs int) (sqltxtt, commentt []string, nextpos int, err error) {
	nextpos = -1 //over
	if needs > 0 {
		f, e := os.Stat(filename)
		if e == nil {
			size := int(f.Size())
			if ipos < size {
				var fi *os.File
				fi, err = os.Open(filename)
				if err == nil {
					bLine := bytes.NewBuffer([]byte{})
					b := make([]byte, defaultBufSize)
					istart := ipos
					bnext := true
					for bnext {
						_, err = fi.Seek(int64(istart), os.SEEK_SET) //func (f *File) Seek(offset int64, whence int) (ret int64, err error)
						if err == nil {
							n := 0
							n, err = fi.Read(b) //func (f *File) Read(b []byte) (n int, err error) {
							if err == nil {
								k := 0
								for i := 0; i < n; i++ {
									if b[i] == '\n' {
										if istart+i+1 < size {
											nextpos = istart + i + 1
										}
										bLine.Write(b[k:i])
										if bLine.Len() > 0 {
											sqltxt, comment := SplitSQL(bLine.String())
											sqltxtt = append(sqltxtt, sqltxt)
											commentt = append(commentt, comment)
											bLine.Reset()
										}
										k = i + 1
										needs--
										if needs == 0 {
											bnext = false
											break
										}
									}
								}
								if bnext {
									bLine.Write(b[k:n])
									istart += n
									if istart >= size {
										bnext = false //nextpos=-1
									}
								}
							} else {
								bnext = false
							}
						} else {
							bnext = false
						}
					}
					fi.Close()
				}
			}
		} else {
			err = errors.New("SQL file not find, must be restart!")
		}
	}
	return
}

func dircopy(src, des string) (e error) {
	var dirr []fs.FileInfo
	dirr, e = ioutil.ReadDir(src)
	if e == nil {
		for _, fi := range dirr {
			fname := fi.Name()
			if fi.IsDir() {
				sdir := filepath.Join(src, fname)
				ddir := filepath.Join(des, fname)
				if !IsExists(ddir) {
					e = os.Mkdir(ddir, 0755)
					if e != nil {
						e = errors.New("mkdir:" + ddir + " " + e.Error())
					}
				}
				dircopy(sdir, ddir)
			} else {
				if !strings.HasPrefix(fname, ".") {
					srcfile := filepath.Join(src, fname)
					desfile := filepath.Join(des, fname)
					_, e = CopyFile(srcfile, desfile)
					if e != nil {
						e = errors.New("copy:" + srcfile + "," + desfile + " " + e.Error())
					}
				}
			}
		}
	}
	return
}

func Xcopy(src, des string) (e error) {
	if !IsExists(des) {
		e = os.MkdirAll(des, 0755)
		if e != nil {
			e = errors.New("mkdir:" + des + " " + e.Error())
		}
	}
	if e == nil {
		e = dircopy(src, des)
	}
	return
}

// src/_xxx is the subdir to merge in des.
func Xmergecopy(src, des, merge string) (e error) {
	if !IsExists(des) {
		e = os.MkdirAll(des, 0755)
		if e != nil {
			e = errors.New("mkdir:" + des + " " + e.Error())
		}
	}
	var dir []os.FileInfo
	dir, e = ioutil.ReadDir(src)
	if e == nil {
		for _, fi := range dir {
			fname := fi.Name()
			if fi.IsDir() {
				if !strings.HasPrefix(fname, "_") {
					sdir := filepath.Join(src, fname)
					ddir := filepath.Join(des, fname)
					if !IsExists(ddir) {
						e = os.Mkdir(ddir, 0755)
						if e != nil {
							e = errors.New("mkdir:" + ddir + " " + e.Error())
						}
					}
					e = dircopy(sdir, ddir)
				}
			} else {
				if !strings.HasPrefix(fname, ".") {
					srcfile := filepath.Join(src, fname)
					desfile := filepath.Join(des, fname)
					_, e = CopyFile(srcfile, desfile)
					if e != nil {
						e = errors.New("copy " + srcfile + " , " + desfile + " " + e.Error())
					}
				}
			}
		}
		sdir := filepath.Join(src, "_"+merge)
		if IsExists(sdir) {
			e = dircopy(sdir, des)
		}
	}
	return
}
