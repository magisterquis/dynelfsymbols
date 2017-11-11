package main

/*
 * file.go
 * Print info about a single file
 * By J. Stuart McMurray
 * Created 20171110
 * Last Modified 20171110
 */

import (
	"debug/elf"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
)

// File holds the printable data for an ELF file
type File struct {
	Name  string   /* File name */
	Libs  []string /* Libraries with the smallest number of symbols */
	Count int      /* Number of symbols imported from MinName */
}

// Lib holds the symbols imported from a library
type Lib struct {
	Name string
	Syms []elf.ImportedSymbol
}

// LibsInFile returns a map containing the imported symbols from each library
// in the given file.
func LibsInFile(fn string) (map[string][]elf.ImportedSymbol, error) {
	/* Open as ELF file */
	f, err := elf.Open(fn)
	if nil != err {
		return nil, err
	}
	defer f.Close()

	/* Get imported symbols */
	iss, err := f.ImportedSymbols()
	if nil != err {
		return nil, err
	}

	/* Collect imported symbols by library name */
	libm := make(map[string][]elf.ImportedSymbol)
	for _, is := range iss {
		if _, ok := libm[is.Library]; !ok {
			libm[is.Library] = []elf.ImportedSymbol{is}
			continue
		}
		libm[is.Library] = append(libm[is.Library], is)
	}

	/* Sort symbol names */
	for k := range libm {
		sort.Slice(libm[k], func(i, j int) bool {
			return libm[k][i].Name < libm[k][j].Name
		})
	}

	return libm, nil
}

// PrintFile prints the libraries used by a file and the symbols imported from
// each library
func PrintFile(fn string) error {
	/* Get the per-library symbols */
	libm, err := LibsInFile(fn)
	if nil != err {
		return err
	}

	/* Sort imported symbols, per-library */
	libs := []Lib{}
	for l, ss := range libm {
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Name < ss[j].Name
		})
		libs = append(libs, Lib{l, ss})
	}
	/* Sort libraries by number of imports */
	sort.Slice(libs, func(i, j int) bool {
		if len(libs[i].Syms) < len(libs[j].Syms) {
			return true
		} else if len(libs[i].Syms) == len(libs[j].Syms) {
			return libs[i].Name < libs[j].Name
		}
		return false
	})

	/* Make a nice table */
	tw := tabwriter.NewWriter(os.Stdout, 2, 8, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "Library\tSymbol\tVersion\n")
	fmt.Fprintf(tw, "-------\t------\t-------\n")
	for _, lib := range libs {
		for _, sym := range lib.Syms {
			fmt.Fprintf(
				tw,
				"%v\t%v\t%v\n",
				sym.Library,
				sym.Name,
				sym.Version,
			)
		}
	}

	return nil
}
