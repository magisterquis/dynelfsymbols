package main

/*
 * dir.go
 * Hunt through directories for nice elfs
 * By J. Stuart McMurray
 * Created 20171110
 * Last Modified 20171110
 */

import (
	"debug/elf"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
)

// FILES holds all the files parsed and the outcome of finding shared libraries
var FILES = []File{}

// HandleDirs recurses through the directories to find which elf files have
// libraries with small numbers of imported symbols.
func HandleDirs(dirs []string) {
	for _, n := range dirs {
		if err := filepath.Walk(n, findFuncs); nil != err {
			fmt.Fprintf(
				os.Stderr,
				"Unable to find imported symbols in %q: %v",
				n,
				err,
			)
			os.Exit(3)
		}
	}

	/* Sort the files by minimum import count */
	sort.Slice(FILES, func(i, j int) bool {
		if FILES[i].Count < FILES[j].Count {
			return true
		} else if FILES[i].Count == FILES[j].Count {
			return FILES[i].Name < FILES[j].Name
		}
		return false
	})

	/* Make a nice table */
	tw := tabwriter.NewWriter(os.Stdout, 2, 8, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "File\tCount\tLibraries\n")
	fmt.Fprintf(tw, "----\t-----\t---------\n")
	for _, v := range FILES {
		fmt.Fprintf(
			tw,
			"%v\t%v\t%v\n",
			v.Name,
			v.Count,
			strings.Join(v.Libs, " "),
		)
	}
}

/* findFuncs finds the imported symbols per-library in path, if it's an ELF
file. */
func findFuncs(path string, info os.FileInfo, err error) error {
	/* Ignore non-regular files */
	if !info.Mode().IsRegular() {
		return nil
	}

	/* Try to open the file */
	f, err := elf.Open(path)
	/* Ignore format errors, these are generally scripts */
	if _, ok := err.(*elf.FormatError); ok {
		return nil
	}
	if nil != err {
		log.Printf("Unable to parse %q: %T", path, err)
		return nil
	}
	defer f.Close()

	/* Get imported symbols */
	iss, err := f.ImportedSymbols()
	if nil != err {
		log.Printf(
			"Unable to get symbols imported by %q: %v",
			path,
			err,
		)
	}

	/* Work out how many symbols are imported from each library */
	libs := make(map[string]int)
	for _, is := range iss {
		if _, ok := libs[is.Library]; !ok {
			libs[is.Library] = 0
		}
		libs[is.Library]++
	}

	/* Struct to represent this file in the output */
	o := File{Name: path}

	/* Find the lowest number of imports */
	o.Count = 0
	mins := map[string]struct{}{}
	for l, c := range libs {
		/* First iteration is the baseline */
		if 0 == o.Count {
			o.Count = c
			mins[l] = struct{}{}
			continue
		}
		if c < o.Count {
			/* If we've found a library with fewer imports, clear
			the list of smallest-import libraries, set this as the
			only one */
			o.Count = c
			mins = map[string]struct{}{l: struct{}{}}
		} else if c == o.Count {
			/* If we find a library with as few imports as the min,
			note it as well. */
			mins[l] = struct{}{}
		}
	}

	/* Sort the list of low-import libraries */
	for n := range mins {
		o.Libs = append(o.Libs, n)
	}
	sort.Strings(o.Libs)

	FILES = append(FILES, o)

	return nil
}
