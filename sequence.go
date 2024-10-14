package base

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Sequence struct {
	Identifier string
	Number     int64
	Filename   string
}

func (s *Sequence) Open(dirstorage, identifier string) (e error) {
	if IsExists(dirstorage) {
		s.Filename = filepath.Join(dirstorage, identifier+".seq")
		e = s.Reset()
	} else {
		e = errors.New(dirstorage + " not exists")
	}
	return
}
func (s *Sequence) Reset() (e error) {
	if IsExists(s.Filename) {
		var b []byte
		b, e = os.ReadFile(s.Filename)
		s.Number = Str2int64(string(b))
	} else {
		s.Number = 0
		s.save2file()
	}
	return
}
func (s *Sequence) save2file() (e error) {
	var sfile *os.File
	sfile, e = os.Create(s.Filename)
	if e == nil {
		_, e = sfile.WriteString(fmt.Sprintf("%d", s.Number))
		sfile.Close()
	}
	return
}
func (s *Sequence) Getone() (number int64, e error) {
	//此处后续加锁，保证唯一
	s.Number++
	number = s.Number
	e = s.save2file()
	return
}
