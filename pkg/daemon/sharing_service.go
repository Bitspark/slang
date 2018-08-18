package daemon

import (
	"github.com/Bitspark/slang/pkg/api"
	"net/http"
	"strings"
	"path/filepath"
	"fmt"
	"bytes"
	"archive/zip"
	"io/ioutil"
	"github.com/Bitspark/slang/pkg/elem"
	"gopkg.in/yaml.v2"
	"time"
	"io"
	"github.com/Bitspark/go-version"
	"os"
)

type manifest struct {
	SlangVersion string `yaml:slangVersion`
	TimeUnix     int64  `yaml:timeUnix`
}

var suffixes = []string{"_visual.yaml"}

func packOperator(e *api.Environ, zipWriter *zip.Writer, fqop string, read map[string]bool) error {
	if r, ok := read[fqop]; ok && r {
		return nil
	}

	relPath := strings.Replace(fqop, ".", string(filepath.Separator), -1)
	absPath, p, err := e.GetFilePathWithFileEnding(relPath, "")
	if err != nil {
		return err
	}

	fileWriter, _ := zipWriter.Create(filepath.ToSlash(absPath[len(p):]))
	fileContents, err := ioutil.ReadFile(absPath)
	if err != nil {
		return err
	}
	fileWriter.Write(fileContents)

	read[fqop] = true

	var absBasePath string
	if strings.HasSuffix(absPath, ".yaml") {
		absBasePath = absPath[:len(absPath)-len(".yaml")]
	} else if strings.HasSuffix(absPath, ".json") {
		absBasePath = absPath[:len(absPath)-len(".json")]
	}

	for _, suffix := range suffixes {
		fileContents, err := ioutil.ReadFile(absBasePath + suffix)
		if err != nil {
			continue
		}
		fileWriter, _ := zipWriter.Create(filepath.ToSlash(absBasePath[len(p):] + suffix))
		fileWriter.Write(fileContents)
	}

	def, err := e.ReadOperatorDef(absPath, nil)
	if err != nil {
		return err
	}

	var baseFqop string
	dotIdx := strings.LastIndex(fqop, ".")
	if dotIdx != -1 {
		baseFqop = fqop[:dotIdx+1]
	}
	for _, ins := range def.InstanceDefs {
		if strings.HasPrefix(ins.Operator, "slang.") {
			continue
		}
		if elem.IsRegistered(ins.Operator) {
			continue
		}
		if !strings.HasPrefix(ins.Operator, ".") {
			packOperator(e, zipWriter, ins.Operator, read)
		} else {
			packOperator(e, zipWriter, baseFqop+ins.Operator[1:], read)
		}
	}

	return nil
}

var SharingService = &Service{map[string]*Endpoint{
	"/export": {func(e *api.Environ, w http.ResponseWriter, r *http.Request) {
		fail := func(err *Error) {
			sendFailure(w, &responseBad{err})
		}
		/*
		 * GET
		 */
		if r.Method == "GET" {
			opFQName := r.FormValue("fqop")

			buf := new(bytes.Buffer)
			zipWriter := zip.NewWriter(buf)

			fileWriter, _ := zipWriter.Create("manifest.yaml")
			manifestBytes, _ := yaml.Marshal(&manifest{
				SlangVersion: SlangVersion,
				TimeUnix:     time.Now().Unix(),
			})
			fileWriter.Write(manifestBytes)

			read := make(map[string]bool)
			err := packOperator(e, zipWriter, opFQName, read)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			zipWriter.Close()

			w.Header().Set("Pragma", "public")
			w.Header().Set("Expires", "0")
			w.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
			w.Header().Set("Cache-Control", "public")
			w.Header().Set("Content-Description", "File Transfer")
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", opFQName))
			w.Header().Set("Content-Transfer-Encoding", "binary")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buf.Bytes())))
			w.Write(buf.Bytes())
		}
	}},
	"/import": {func(e *api.Environ, w http.ResponseWriter, r *http.Request) {
		fail := func(err *Error) {
			sendFailure(w, &responseBad{err})
		}
		/*
		 * GET
		 */
		if r.Method == "POST" {
			var buf bytes.Buffer
			file, header, err := r.FormFile("file")
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}
			defer file.Close()

			io.Copy(&buf, file)

			zipReader, err := zip.NewReader(file, header.Size)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			manifest := manifest{}
			for _, file := range zipReader.File {
				if file.Name == "manifest.yaml" {
					fileReader, _ := file.Open()
					buf := new(bytes.Buffer)
					buf.ReadFrom(fileReader)
					yaml.Unmarshal(buf.Bytes(), &manifest)
					fileReader.Close()
				}
			}

			myVersion, err := version.NewVersion(SlangVersion)
			if err == nil {
				manifestVersion, err := version.NewVersion(manifest.SlangVersion)
				if err == nil {
					if myVersion.LessThan(manifestVersion) {
						fail(&Error{Msg: "Please upgrade your slang version", Code: "E000X"})
						return
					}
				}
			}

			baseDir := e.WorkingDir()
			for _, file := range zipReader.File {
				if file.Name == "manifest.yaml" {
					continue
				}

				fileReader, _ := file.Open()
				buf := new(bytes.Buffer)
				buf.ReadFrom(fileReader)
				fpath := filepath.Join(baseDir, filepath.FromSlash(file.Name))
				os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
				ioutil.WriteFile(fpath, buf.Bytes(), os.ModePerm)
				fileReader.Close()
			}

			writeJSON(w, &struct{ Success bool `json:"success"` }{true})
		}
	}},
}}
