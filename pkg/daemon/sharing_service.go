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
)

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

	fileWriter, _ := zipWriter.Create(absPath[len(p):])
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
		fileWriter, _ := zipWriter.Create(absBasePath[len(p):] + suffix)
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
			w.Header().Set("Content-Description", "File Transfe")
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", opFQName))
			w.Header().Set("Content-Transfer-Encoding", "binary")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buf.Bytes())))
			w.Write(buf.Bytes())
		}
	}},
	"/import": {func(e *api.Environ, w http.ResponseWriter, r *http.Request) {

	}},
}}
