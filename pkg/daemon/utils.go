package daemon

import (
	"os"
	"strings"
	"net/http"
	"io"
	"archive/zip"
	"path/filepath"
)

func EnsureEnvironVar(key string, dfltVal string) string {
	if val := os.Getenv(key); strings.Trim(val, " ") != "" {
		return val
	}
	os.Setenv(key, dfltVal)
	return dfltVal
}

func download(srcUrl string, dstFile *os.File) error {
	response, err := http.Get(srcUrl)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(dstFile, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func unzip(srcPath string, dstPath string) ([]string, error) {
	var filePaths []string

	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return filePaths, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filePaths, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		filePath := filepath.Join(dstPath, f.Name)
		filePaths = append(filePaths, filePath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(filePath, os.ModePerm)
		} else {
			// Make File
			if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				return filePaths, err
			}

			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filePaths, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filePaths, err
			}

		}
	}
	return filePaths, nil
}

func moveAll(srcDir string, dstDir string, skipFirstLevel bool) error {
	var outerErr error
	filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			outerErr = err
			return err
		}

		if info.IsDir() {
			return nil
		}

		var dstFilePath string
		// skip directory level: /root/dir0/file1  ==> /root/file1
		relFilePath, err := filepath.Rel(srcDir, path)
		if err != nil {
			outerErr = err
			return err
		}
		if skipFirstLevel {
			if dir, _ := filepath.Split(path); dir == "/" {
				// skip all children of root
				return nil
			}
			// omit string till *idx*
			idx := strings.IndexRune(relFilePath[1:], filepath.Separator) + 1
			relFilePath = relFilePath[idx:]
		}
		dstFilePath = filepath.Join(dstDir, relFilePath)

		if err = os.MkdirAll(filepath.Dir(dstFilePath), os.ModePerm); err != nil {
			outerErr = err
			return err
		}

		if err = os.Rename(path, dstFilePath); err != nil {
			outerErr = err
			return err
		}

		return nil
	})
	return outerErr
}
