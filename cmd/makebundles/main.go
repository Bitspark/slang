package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
)

var libDir string
var outDir string

func main() {
	var bundleLib bool
	var bundleElems bool

	flag.StringVar(&libDir, "libdir", "./", "Input location of the standard library files")
	flag.StringVar(&outDir, "outdir", "./", "Output location of the bundle files")
	flag.BoolVar(&bundleLib, "bundlelib", true, "Bundle standard library")
	flag.BoolVar(&bundleElems, "bundleelems", true, "Bundle elementaries")
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	store := storage.NewStorage().AddBackend(storage.NewReadOnlyFileSystem(libDir))

	var uuids []uuid.UUID

	if bundleElems {
		elemUUIDs := elem.GetBuiltinIds()
		for _, id := range elemUUIDs {
			uuids = append(uuids, id)
		}
	}

	if bundleLib {
		libraryUUIDs, err := store.List()
		if err != nil {
			panic(err)
		}
		for _, id := range libraryUUIDs {
			uuids = append(uuids, id)
		}
	}

	for _, u := range uuids {
		def, err := store.Load(u)
		if err != nil {
			panic(err)
		}
		err = makeBundle(def, store)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("%d blueprints have been bundled\n", len(uuids))
}

func makeBundle(def *core.Blueprint, store *storage.Storage) error {
	b, err := api.CreateBundle(def, store)

	if err != nil {
		return err
	}

	opDefJson, err := json.Marshal(&b)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(outDir, def.Id.String()+".slang.json"), opDefJson, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
