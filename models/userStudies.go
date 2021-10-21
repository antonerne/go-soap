package models

import (
	"time"

	"gorm.io/gorm"
)

type UserBibleStudy struct {
	ID           uint64                 `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	UserID       string                 `json:"-" gorm:"column:userid"`
	BibleStudyID uint64                 `json:"bible_study_id" gorm:"column:bible_study_id"`
	StartDate    time.Time              `json:"startdate" gorm:"column:startdate"`
	EndDate      time.Time              `json:"enddate" gorm:"column:enddate"`
	Periods      []UserBibleStudyPeriod `json:"periods" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserBibleStudy) TableName() string {
	return "user_bible_study"
}

func (ubs *UserBibleStudy) SetNew(userid string, bibleStudy BibleStudy, db *gorm.DB) {
	ubs.BibleStudyID = bibleStudy.ID
	ubs.StartDate = time.Now()
	ubs.UserID = userid
	db.Create(&ubs)
	ubs.Periods = make([]UserBibleStudyPeriod, 0)

	for _, per := range bibleStudy.Periods {
		var period UserBibleStudyPeriod
		period.SetNew(ubs.ID, per, db)
		ubs.Periods = append(ubs.Periods, period)
	}
}

// ByUserBibleStudy will contain the list of a User's studies
type ByUserBibleStudy []UserBibleStudy

func (s ByUserBibleStudy) Len() int           { return len(s) }
func (s ByUserBibleStudy) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByUserBibleStudy) Less(i, j int) bool { return s[i].StartDate.Before(s[j].StartDate) }

type UserBibleStudyPeriod struct {
	ID               uint64              `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	UserBibleStudyID uint64              `json:"-" gorm:"column:user_bible_study_id"`
	Period           uint                `json:"period" gorm:"column:period"`
	Title            string              `json:"title" gorm:"column:title"`
	StudyDays        []UserBibleStudyDay `json:"studydays" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserBibleStudyPeriod) TableName() string {
	return "user_bible_study_period"
}

func (p *UserBibleStudyPeriod) SetNew(studyID uint64, period BibleStudyPeriod,
	db *gorm.DB) {
	p.UserBibleStudyID = studyID
	p.Period = period.Period
	p.Title = period.Title
	db.Create(&p)
	p.StudyDays = make([]UserBibleStudyDay, 0)

	for _, day := range period.StudyDays {
		var d UserBibleStudyDay
		d.SetNew(p.ID, day, db)
		p.StudyDays = append(p.StudyDays, d)
	}
}

// ByUserBibleStudy will contain the list of a User's studies
type ByUserBibleStudyPeriod []UserBibleStudyPeriod

func (s ByUserBibleStudyPeriod) Len() int           { return len(s) }
func (s ByUserBibleStudyPeriod) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByUserBibleStudyPeriod) Less(i, j int) bool { return s[i].Period < s[j].Period }

type UserBibleStudyDay struct {
	ID                     uint64                    `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	UserBibleStudyPeriodID uint64                    `json:"-" gorm:"column:bible_study_period_id"`
	Day                    uint                      `json:"day" gorm:"column:day"`
	References             []UserBibleStudyReference `json:"references" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserBibleStudyDay) TableName() string {
	return "user_bible_study_period_days"
}

func (d *UserBibleStudyDay) SetNew(periodID uint64, day BibleStudyDay, db *gorm.DB) {
	d.UserBibleStudyPeriodID = periodID
	d.Day = day.Day
	db.Create(&d)
	d.References = make([]UserBibleStudyReference, 0)

	for _, ref := range day.References {
		var r UserBibleStudyReference
		r.SetNew(d.ID, ref, db)
		d.References = append(d.References, r)
	}
}

// ByUserBibleStudy will contain the list of a User's studies
type ByUserBibleStudyDay []UserBibleStudyDay

func (s ByUserBibleStudyDay) Len() int           { return len(s) }
func (s ByUserBibleStudyDay) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByUserBibleStudyDay) Less(i, j int) bool { return s[i].Day < s[j].Day }

type UserBibleStudyReference struct {
	ID                  uint64 `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	UserBibleStudyDayID uint64 `json:"-" gorm:"column:user_bible_study_day_id"`
	BookID              uint   `json:"_" gorm:"column:book_id"`
	Chapter             uint   `json:"chapter" gorm:"column:chapter"`
	Verses              string `json:"verses,omitempty" gorm:"column:verses"`
	Completed           bool   `json:"completed,omitempty" gorm:"-"`
}

func (UserBibleStudyReference) TableName() string {
	return "user_bible_study_reference"
}

func (r *UserBibleStudyReference) SetNew(dayID uint64,
	ref BibleStudyDayReference, db *gorm.DB) {
	r.UserBibleStudyDayID = dayID
	r.BookID = ref.BookID
	r.Chapter = ref.Chapter
	r.Verses = ref.Verses
	r.Completed = false
	db.Create(&r)
}

// ByUserBibleStudy will contain the list of a User's studies
type ByUserBibleStudyReference []UserBibleStudyReference

func (s ByUserBibleStudyReference) Len() int      { return len(s) }
func (s ByUserBibleStudyReference) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByUserBibleStudyReference) Less(i, j int) bool {
	if s[i].BookID == s[j].BookID {
		return s[i].Chapter < s[j].Chapter
	}
	return s[i].BookID < s[j].BookID
}
