package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"path"
)

var libDir string
var outDir string

type SlangBundle struct {
	Main       string              `json:"main"`
	Blueprints map[string]*core.OperatorDef `json:"blueprints"`
}

func main() {
	var showHelp bool
	var bundleLib bool
	var bundleElems bool

	flag.BoolVar(&showHelp, "help", false, "Show this dialog")
	flag.StringVar(&libDir, "libdir", "./", "Input location of the standard library files")
	flag.StringVar(&outDir, "outdir", "./", "Output location of the bundle files")
	flag.BoolVar(&bundleLib, "bundlelib", true, "Bundle standard library")
	flag.BoolVar(&bundleElems, "bundleelems", true, "Bundle elementaries")
	flag.Parse()

	if showHelp {
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

func makeBundle(def *core.OperatorDef, store *storage.Storage) error {
	b := SlangBundle{
		Main:       def.Id,
	}
	b.Blueprints = make(map[string]*core.OperatorDef)

	gatherDependencies(def, &b, store)

	opDefJson, err := json.Marshal(&b)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(outDir, def.Id+".bundle.json"), opDefJson, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func gatherDependencies(def *core.OperatorDef, bundle *SlangBundle, store *storage.Storage) error {
	for _, dep := range def.InstanceDefs {
		id, err := uuid.Parse(dep.Operator)
		if err != nil {
			return err
		}
		if _, ok := bundle.Blueprints[id.String()]; !ok {
			depDef, err := store.Load(id)
			if err != nil {
				return err
			}
			bundle.Blueprints[id.String()] = depDef
			err = gatherDependencies(depDef, bundle, store)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
