package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"go/importer"
	"go/types"
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

func lowerFirst(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

type item struct {
	pkg  string
	recv string
	name string
}

func (i item) RenameCommand() string {
	if i.recv == "" {
		return "gorename -from \\\"" + i.pkg + "\\\"." + i.name + " -to " + lowerFirst(i.name)
	} else {
		return "gorename -from \\\"" + i.pkg + "\\\"." + i.recv + "." + i.name + " -to " + lowerFirst(i.name)
	}
}

func (i item) String() string {
	path := make([]string, 0)
	pkgSplit := strings.Split(i.pkg, "/")
	if len(pkgSplit) > 3 {
		path = append(path, pkgSplit[len(pkgSplit)-1])
	}
	if i.recv != "" {
		path = append(path, i.recv)
	}
	path = append(path, i.name)
	return strings.Join(path, ".")
}

func (i item) ToCSV() string {
	return fmt.Sprintf("%s,%s,%s", i.pkg, i.recv, i.name)
}

func getItems(pkgName string) ([]item, error) {
	items := make([]item, 0)
	pkg, err := importer.Default().Import(pkgName)

	if err != nil {
		return nil, errors.Wrap(err, "importing pkg")
	}

	scope := pkg.Scope()

	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj.Exported() {
			switch typ := obj.Type().Underlying().(type) {
			case *types.Struct:
				items = append(items, item{pkgName, "", name})

				ms := types.NewMethodSet(types.NewPointer(obj.Type()))
				for i := 0; i < ms.Len(); i++ {
					method := ms.At(i)
					if method.Obj().Exported() {
						items = append(items, item{pkgName, name, method.Obj().Name()})
					}
				}

				for i := 0; i < typ.NumFields(); i++ {
					field := typ.Field(i)
					if field.Exported() {
						items = append(items, item{pkgName, name, field.Name()})
					}
				}
			case *types.Interface:
				// ignore interfaces
			default:
				items = append(items, item{pkgName, "", name})
			}
		}
	}
	return items, nil
}

func list(pkgs []string) {
	for _, pkg := range pkgs {
		items, err := getItems(pkg)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, item := range items {
			fmt.Println(item.ToCSV())
		}
	}
}

func run(pkgs []string, blacklist string) {
	bl := make([]item, 0)
	if blacklist != "" {
		f, err := os.Open(blacklist)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		reader := csv.NewReader(f)

		records, err := reader.ReadAll()
		if err != nil {
			log.Fatal(err)
		}
		for _, record := range records {
			if len(record) != 3 {
				log.Fatal("CSV: each line should have 3 fields")
			}
			i := item{record[0], record[1], record[2]}
			bl = append(bl, i)
		}
	}

	wl := make([]item, 0)

	for _, pkg := range pkgs {
		items, err := getItems(pkg)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, i := range items {
			found := false
			for _, bi := range bl {
				// check blacklist
				if i == bi {
					found = true
				}
			}
			if !found {
				wl = append(wl, i)
			}
		}

	}
	for _, i := range wl {
		fmt.Println(i.RenameCommand())
	}
}

func main() {
	flagList := flag.Bool("l", false, "List names (CSV)")
	flagRun := flag.Bool("r", false, "Run unexport")
	flagPkg := flag.String("p", "github.com/pilosa/unexport", "Package to operate on")
	flagBlacklist := flag.String("b", "", "CSV of items to blacklist")
	flag.Parse()

	log.Println("Package:", *flagPkg)

	if *flagBlacklist != "" {
		log.Println("Blacklist:", *flagBlacklist)
	}

	cmd := exec.Command("go", "list", *flagPkg)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	pkgSlice := strings.Split(string(out), "\n")
	pkgs := pkgSlice[:len(pkgSlice)-1]

	switch {
	case *flagList:
		list(pkgs)
	case *flagRun:
		run(pkgs, *flagBlacklist)
	default:
		log.Fatal("Must specify -l (list) or -r (run)")
	}

}
