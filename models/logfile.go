package models

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogFile struct {
	Directory string
	FileType  string
}

func (l *LogFile) Exists() bool {
	filename := fmt.Sprintf("%s-%s.log", l.FileType,
		time.Now().Format("2006-01-02"))
	path := filepath.Join(l.Directory, filename)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (l *LogFile) WriteToLog(msg string) {
	filename := fmt.Sprintf("%s-%s.log", l.FileType,
		time.Now().Format("2006-01-02"))
	path := filepath.Join(l.Directory, filename)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	msgOut := fmt.Sprintf("%s - %s\n", time.Now().Format("2006-01-02 15:04"), msg)

	if _, err := f.WriteString(msgOut); err != nil {
		log.Println(err)
	}
}
