package base

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

func AddWeekdays() {
	tm := time.Now()
	curYear := tm.Year()
	if tm.Hour() == 23 && tm.Minute() >= 40 {
		AddYearWorkdays(curYear) //check this year workdays
		if tm.Month() == time.December && tm.Day() == 31 {
			AddYearWorkdays(curYear + 1) //add next year workdays
		}
	}
}

func AddThisYearWorkdays() (e error) {
	return AddYearWorkdays(time.Now().Year())
}

func AddYearWorkdays(year int) (e error) {
	db := DB()
	if db != nil {
		e = addYearWorkdays(db, year)
	} else {
		e = errors.New("DB not ready!")
	}
	return
}

func addYearWorkdays(db *sql.DB, year int) (e error) {
	days := 0
	asql := "select count(1) from workday where year=?"
	row := db.QueryRow(asql, year)
	if row != nil {
		row.Scan(&days)
	}
	tm := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	n := tm.YearDay()
	if days != n {
		tm = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < n; i++ {
			daytype := 1 //1:正常工作日 2:调换工作日 3:法定假日 4:调休假日
			iweekday := int(tm.Weekday())
			if iweekday == 0 || iweekday == 6 {
				daytype = 3
			}
			code := strconv.FormatInt(tm.Unix(), 10)
			name := tm.Format("2006-01-02")
			year := tm.Year()
			month := int(tm.Month())
			day := tm.Day()
			ct := 0
			asql := "select count(1) from workday where code=?"
			row := db.QueryRow(asql, code)
			if row != nil {
				row.Scan(&ct)
			}
			if ct == 0 {
				asql = "insert into workday(code,name,year,month,day,weekday,normaltype_id,daytype_id,time_created,time_updated) values(?,?,?,?,?,?,?,?," + SQL_now() + "," + SQL_now() + ")"
				_, e = db.Exec(asql, code, name, year, month, day, iweekday, daytype, daytype)
			}
			if e != nil {
				break
			}
			tm = tm.AddDate(0, 0, 1)
		}
	}
	return
}
