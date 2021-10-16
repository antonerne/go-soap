package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type BibleBook struct {
	ID        uint   `json:"id" bson:"_id"`
	Code      string `json:"code" bson:"code"`
	Title     string `json:"title" bson:"title"`
	Chapters  uint   `json:"chapters" bson:"chapters"`
	Apocrapha bool   `json:"apocrapha,omitempty" bson:"apocrapha,omitempty"`
}

// ByBibleBooks will containt the list of books and allow for sorting
type ByBibleBooks []BibleBook

func (s ByBibleBooks) Len() int           { return len(s) }
func (s ByBibleBooks) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByBibleBooks) Less(i, j int) bool { return s[i].ID < s[j].ID }

type BibleStudyDayReference struct {
	BookID    uint   `json:"bookid" bson:"bookid"`
	Code      string `json:"code,omitempty" bson:"code,omitempty"`
	Title     string `json:"title,omitempty" bson:"title,omitempty"`
	Chapter   uint   `json:"chapter" bson:"chapter"`
	Verses    string `json:"verses,omitempty" bson:"verses,omitempty"`
	Completed bool   `json:"completed,omitempty" bson:"completed"`
}

func (bsr *BibleStudyDayReference) AssignBook(book BibleBook) {
	bsr.Code = book.Code
	bsr.Title = book.Title
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
	Day        uint                     `json:"day" bson:"day"`
	References []BibleStudyDayReference `json:"references" bson:"references"`
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
	Period    uint            `json:"period" bson:"period"`
	Title     string          `json:"title" bson:"title"`
	StudyDays []BibleStudyDay `json:"studydays" bson:"studydays"`
}

// ByBibleStudyPeriod will contain the list of study periods defining the study
type ByBibleStudyPeriod []BibleStudyPeriod

func (s ByBibleStudyPeriod) Len() int           { return len(s) }
func (s ByBibleStudyPeriod) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByBibleStudyPeriod) Less(i, j int) bool { return s[i].Period < s[j].Period }

type BibleStudy struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	Title   string             `json:"title" bson:"title"`
	Days    uint               `json:"days" bson:"days"`
	Periods []BibleStudyPeriod `json:"periods" bson:"periods"`
}

// ByBibleStudyPeriod will contain the list of study periods defining the study
type ByBibleStudy []BibleStudy

func (s ByBibleStudy) Len() int           { return len(s) }
func (s ByBibleStudy) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByBibleStudy) Less(i, j int) bool { return s[i].Title < s[j].Title }
