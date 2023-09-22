package journal

import (
	"errors"
	"os"
	"strings"
)

type FileJournal struct {
	file string
}

func NewFileJournal(file string) (*FileJournal, error) {
	if len(file) == 0 {
		return nil, errors.New("empty file provided")
	}

	return &FileJournal{file: file}, nil
}

func (a *FileJournal) Journal(action string) error {
	return appendToFile(a.file, action)
}

func appendToFile(filename, content string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(filename)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer file.Close()

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
