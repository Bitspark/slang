package daemon

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/Bitspark/go-version"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type manifest struct {
	SlangVersion string `yaml:"slangVersion"`
	TimeUnix     int64  `yaml:"timeUnix"`
}

/* TODO re-implemet this service
 */

var suffixes = []string{"_visual.yaml"}

var SharingService = &Service{map[string]*Endpoint{
	"/export": {func(st Storage, w http.ResponseWriter, r *http.Request) {
		fail := func(err *Error) {
			sendFailure(w, &responseBad{err})
		}
		/*
		 * GET
		 */
		if r.Method == "GET" {
			opId, err := uuid.Parse(r.FormValue("id"))

			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			buf := new(bytes.Buffer)
			zipWriter := zip.NewWriter(buf)

			fileWriter, _ := zipWriter.Create("manifest.yaml")
			manifestBytes, _ := yaml.Marshal(&manifest{
				SlangVersion: SlangVersion,
				TimeUnix:     time.Now().Unix(),
			})

			fileWriter.Write(manifestBytes)

			/* TODO
			read := make(map[uuid.UUID]bool)
			err = st.PackOperator(zipWriter, opId, read)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}
			*/

			zipWriter.Close()

			w.Header().Set("Pragma", "public")
			w.Header().Set("Expires", "0")
			w.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
			w.Header().Set("Cache-Control", "public")
			w.Header().Set("Content-Description", "File Transfer")
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", opId))
			w.Header().Set("Content-Transfer-Encoding", "binary")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buf.Bytes())))
			w.Write(buf.Bytes())
		}
	}},
	"/import": {func(st Storage, w http.ResponseWriter, r *http.Request) {
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

			fail(&Error{Msg: "not implemented yet", Code: "E000X"})
			return

			baseDir := "" //st.WorkingDir()
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

			writeJSON(w, &struct {
				Success bool `json:"success"`
			}{true})
		}
	}},
}}
