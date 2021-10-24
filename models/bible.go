package models

type BibleBook struct {
	ID        uint   `json:"id" gorm:"primaryKey;column:id"`
	Code      string `json:"code" gorm:"column:code"`
	Title     string `json:"title" gorm:"column:title"`
	Chapters  uint   `json:"chapters" gorm:"column:chapters"`
	Apocrapha bool   `json:"apocrapha,omitempty" gorm:"column:apocrapha"`
}

func (BibleBook) TableName() string {
	return "bible_books"
}

// ByBibleBooks will containt the list of books and allow for sorting
type ByBibleBooks []BibleBook

func (s ByBibleBooks) Len() int           { return len(s) }
func (s ByBibleBooks) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByBibleBooks) Less(i, j int) bool { return s[i].ID < s[j].ID }

type BibleStudyDayReference struct {
	ID              uint64 `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	BibleStudyDayID uint64 `json:"-" gorm:"column:bible_study_day_id"`
	BookID          uint   `json:"_" gorm:"column:book_id"`
	Chapter         uint   `json:"chapter" gorm:"column:chapter"`
	Verses          string `json:"verses,omitempty" gorm:"column:verses"`
	Completed       bool   `json:"completed,omitempty" gorm:"-"`
}

func (BibleStudyDayReference) TableName() string {
	return "bible_study_day_reference"
}

// ByBibleBooks will containt the list of books and allow for sorting
type ByBibleStudyDayReference []BibleStudyDayReference

func (s ByBibleStudyDayReference) Len() int      { return len(s) }
func (s ByBibleStudyDayReference) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByBibleStudyDayReference) Less(i, j int) bool {
	if s[i].BookID == s[j].BookID {
		return s[i].Chapter < s[j].Chapter
	}
	return s[i].BookID < s[j].BookID
}

type BibleStudyDay struct {
	ID                 uint64                   `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	BibleStudyPeriodID uint64                   `json:"-" gorm:"column:bible_study_period_id"`
	Day                uint                     `json:"day" gorm:"column:day"`
	References         []BibleStudyDayReference `json:"references" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (BibleStudyDay) TableName() string {
	return "bible_study_period_days"
}

// ByBibleStudyDay will contain the list of study days during a period
type ByBibleStudyDay []BibleStudyDay

func (s ByBibleStudyDay) Len() int           { return len(s) }
func (s ByBibleStudyDay) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByBibleStudyDay) Less(i, j int) bool { return s[i].Day < s[j].Day }

func (d *BibleStudyDay) IsComplete() bool {
	for _, ref := range d.References {
		if !ref.Completed {
			return false
		}
	}
	return true
}

type BibleStudyPeriod struct {
	ID           uint64          `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	BibleStudyID uint64          `json:"-" gorm:"column:bible_study_id"`
	Period       uint            `json:"period" gorm:"column:period"`
	Title        string          `json:"title" gorm:"column:title"`
	StudyDays    []BibleStudyDay `json:"studydays" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (BibleStudyPeriod) TableName() string {
	return "bible_study_periods"
}

// ByBibleStudyPeriod will contain the list of study periods defining the study
type ByBibleStudyPeriod []BibleStudyPeriod

func (s ByBibleStudyPeriod) Len() int           { return len(s) }
func (s ByBibleStudyPeriod) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByBibleStudyPeriod) Less(i, j int) bool { return s[i].Period < s[j].Period }

type BibleStudy struct {
	ID               uint64             `json:"id,omitempty" gorm:"primaryKey;column:id;autoIncrement"`
	Title            string             `json:"title" gorm:"column:title"`
	Days             uint               `json:"days" gorm:"column:days"`
	BeginImmediately bool               `json:"begin,omitempty" gorm:"column:begin"`
	Periods          []BibleStudyPeriod `json:"periods" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (BibleStudy) TableName() string {
	return "bible_studies"
}

// ByBibleStudyPeriod will contain the list of study periods defining the study
type ByBibleStudy []BibleStudy

func (s ByBibleStudy) Len() int           { return len(s) }
func (s ByBibleStudy) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByBibleStudy) Less(i, j int) bool { return s[i].Title < s[j].Title }
