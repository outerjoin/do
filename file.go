package do

import (
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/bugsnag/osext"
)

func FileContent(filepath string) ([]byte, error) {

	var absPath string
	var exists bool

	// If absolute path, try to find the file directly
	if strings.HasPrefix(filepath, "/") {
		absPath = filepath
		_, err := os.Stat(absPath)
		exists = err == nil
	}

	// Try to locate it in working directory
	if !exists {
		wdir, ferr := os.Getwd()
		if ferr == nil {
			absPath = CleanFilePath(wdir + "/" + filepath)
			_, err := os.Stat(absPath)
			exists = err == nil
		}
	}

	// Try to locate it in executable directory
	if !exists {
		edir, ferr := osext.ExecutableFolder()
		if ferr == nil {
			absPath = CleanFilePath(edir + "/" + filepath)
			_, err := os.Stat(absPath)
			exists = err == nil
		}
	}

	if !exists {
		return []byte{}, errors.New("file could not be located in working or executable path: " + filepath)
	}

	b, err := ioutil.ReadFile(absPath)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

var doubleSlash *regexp.Regexp

func CleanFilePath(path string) string {
	if doubleSlash == nil {
		doubleSlash, _ = regexp.Compile(`[\\/]+`)
	}
	return doubleSlash.ReplaceAllString(path, "/")
}
