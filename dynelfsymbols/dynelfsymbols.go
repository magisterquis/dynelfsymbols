// dynelfsymbols finds the symbols needed for shared object monkey business
package main

/*
 * dynelfsymbols.go
 * Find exported symbols and nice files
 * By J. Stuart McMurray
 * Created 20171110
 * Last Modified 20171111
 */

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		csrc = flag.String(
			"c",
			"",
			"Print C stub code to backdoor the `library`",
		)
	)
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: %v [options] directory [directory...]
       %v [-c library] ElfFile

When a directory or directories are given, a list of ELF files is printed,
sorted by the smallest number of imported symbols from a single library.  This
is meant to make it easy to find a binary which can be edited to point at a
malicious library which MitMs a minimal number of functions.  Each file name is
listed, along with the libraries from which the smallest number of symbols are
imported and how many symbols are imported from the libraries.

When a single file is given, a list of imported symbols is printed, sorted by
the number of symbols imported from each symbol's library.  This is meant for
validation that an ELF file is suitable for modification, and that none of the
libraries to be intercepted are unsuitable for the purpose.

With -c <library>, C stub code is printed which aims to make implementing 
library to MitM the library passed to -c much easier.  The functions' arguments
and return types will need to be set in the source, and malicious code will
need to be added to the contructor.

Options:
`,
			os.Args[0],
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	/* Make sure we have a target */
	if 0 == flag.NArg() {
		flag.Usage()
		os.Exit(1)
	}

	/* If we're to make a C stub, do it */
	if "" != *csrc {
		if 1 != flag.NArg() {
			fmt.Fprintf(
				os.Stderr,
				"Unable to print C stub for more than "+
					"one file\n",
			)
			os.Exit(5)
		}
		PrintStub(flag.Arg(0), *csrc)
		return
	}

	/* If we have a single file, print the libraries and symbols */
	if 1 == flag.NArg() {
		/* Get file info */
		fi, err := os.Stat(flag.Arg(0))
		if nil != err {
			fmt.Fprintf(os.Stderr, "Unable to stat %v\n", err)
			os.Exit(2)
		}
		/* If it's a regular file, print libraries and symbols */
		if fi.Mode().IsRegular() {
			if err := PrintFile(flag.Arg(0)); nil != err {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(3)
			}
			os.Exit(0)
		}
	}

	/* If we have a directory or list of files, find the library with the
	smallest number of imported symbols for each, and print. */
	HandleDirs(flag.Args())
}
