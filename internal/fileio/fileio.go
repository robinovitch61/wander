package fileio

import (
	"fmt"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type SaveCompleteMessage struct {
	FullPath, SuccessMessage, Err string
}

func SaveToFile(saveDialogValue string, fileContent []string) (string, error) {
	var path, fileName string

	if saveDialogValue == "" {
		path = "."
		fileName = formatter.FormatTime(time.Now())
	} else {
		if strings.Contains(saveDialogValue, "~") {
			currUser, userErr := user.Current()
			if userErr != nil {
				return "", userErr
			}
			saveDialogValue = strings.ReplaceAll(saveDialogValue, "~", currUser.HomeDir)
		}

		if strings.Contains(saveDialogValue, string(os.PathSeparator)) {
			path = filepath.Dir(saveDialogValue)
			fileName = filepath.Base(saveDialogValue)
		} else {
			path = "."
			fileName = saveDialogValue
		}
	}

	cleanPath, cleanPathErr := filepath.Abs(path)
	if cleanPathErr != nil {
		return "", cleanPathErr
	}

	if exists, pathExistsErr := fileOrDirectoryExists(cleanPath); pathExistsErr == nil {
		if !exists {
			if mkdirErr := os.MkdirAll(cleanPath, 0755); mkdirErr != nil {
				return "", mkdirErr
			}
		}
	} else {
		return "", pathExistsErr
	}

	pathWithFileName := fmt.Sprintf("%s/%s", cleanPath, fileName)

	if exists, fileExistsErr := fileOrDirectoryExists(pathWithFileName); fileExistsErr == nil {
		if exists {
			extension := filepath.Ext(pathWithFileName)
			now := formatter.FormatTime(time.Now())
			if extension == "" {
				pathWithFileName += "_" + now
			} else {
				pathWithFileName = strings.ReplaceAll(pathWithFileName, extension, now+extension)
			}
		}
	} else {
		return "", fileExistsErr
	}

	f, createErr := os.Create(pathWithFileName)
	if createErr != nil {
		return "", createErr
	}
	defer f.Close()

	for _, line := range fileContent {
		_, writeErr := f.WriteString(line)
		if writeErr != nil {
			return "", writeErr
		}
	}

	return pathWithFileName, nil
}

func fileOrDirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
