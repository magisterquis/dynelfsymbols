package main

/*
 * stub.go
 * Print a C library stub
 * By J. Stuart McMurray
 * Created 20171110
 * Last Modified 20171111
 */

import (
	"debug/elf"
	"fmt"
	"os"
)

// PrintStub takes the name of the binary as well as the library to intercept
// and prints a stub C file.
func PrintStub(bin, lib string) {
	/* Get the libraries and symbols for the file */
	libm, err := LibsInFile(bin)
	if nil != err {
		fmt.Fprintf(
			os.Stderr,
			"Unable to get libraries and symbols from %q: %v\n",
			bin,
			err,
		)
		os.Exit(4)
	}

	/* Get the symbols for the library */
	syms, ok := libm[lib]
	if !ok {
		fmt.Fprintf(os.Stderr, "No symbols from %q in %q\n", lib, bin)
	}

	/* Header */
	fmt.Printf(
		`#define _GNU_SOURCE
#include <dlfcn.h>
#include <err.h>
#include <stdlib.h>

`)

	/* Version Script */
	fmt.Printf("/* Begin version script:\n")
	fmt.Printf("------------------------\n")
	printVersionScript(syms)
	fmt.Printf("---------------------\n")
	fmt.Printf("End version script */\n")

	/* Constructor prototype */
	fmt.Printf(`
/* Constructor prototype */
static void con(void) __attribute__((constructor));

`)

	/* Function prototypes */
	fmt.Printf("/* Exported function prototypes */\n")
	for _, s := range syms {
		fmt.Printf("TYPE_CHANGEME %v(ARGS_CHANGEME);\n", s.Name)
	}
	fmt.Printf("\n")
	fmt.Printf("/* Pointers to real functions */\n")
	for _, s := range syms {
		fmt.Printf("static TYPE_CHANGEME (*%v_real)(ARGS_CHANGEME) = NULL;\n", s.Name)
	}

	/* Exported functions which call real functions */
	fmt.Printf(`
/* Exported functions which call real functions */
`)
	for _, s := range syms {
		fmt.Printf(
			"TYPE_CHANGEME %v(ARGS_CHANGEME) {return %v_real(ARGS_CHANGEME);}\n",
			s.Name,
			s.Name,
		)
	}

	/* Constructor, which loads the real library and makes function
	pointers, and may fire off something bad */
	fmt.Printf(`
/* con is called when the library is loaded */
void
con(void)
{
	char *e;

	/* Load the real %v */
	if (NULL == dlopen("%v", RTLD_NOW|RTLD_GLOBAL))
		errx(1, "dlopen: %%s", dlerror());

`,
		lib,
		lib,
	)

	/* Grab the real functions */
	fmt.Printf("\t/* Get hold of the real functions */\n\n")
	for _, s := range syms {
		fmt.Printf("\t/* %v */\n\t"+`dlerror();
	%v_real = dlsym(RTLD_NEXT, "%v");
	if (NULL != (e = dlerror()))
		errx(2, "dlsym %v: %%s", e);

`,
			s.Name,
			s.Name,
			s.Name,
			s.Name,
		)
	}
	fmt.Printf("\t" + `/************************************
	 * Further malicious code goes here *
	 ************************************/
}
`)
	/* A bit of compilation help */
	fmt.Printf(`
/*
Slice off the version script and save it to a file named "vs".  After modifying
the ARGS_CHANGEMEs and TYPE_CHANGEMEs to reflect the real function prototypes,
he rest of this C source code can be compiled with a command similar to:

cc -O2 -shared -fPIC -Wl,--version-script=vs -o foo.so foo.c -ldl
*/
`)
}

/* printVersionScript prints a version script for all the symbols in syms */
func printVersionScript(syms []elf.ImportedSymbol) {
	/* vers contains the symbols for each version */
	vers := map[string][]string{}

	/* Populate vers */
	for _, sym := range syms {
		if _, ok := vers[sym.Version]; !ok {
			vers[sym.Version] = []string{sym.Name}
			continue
		}
		vers[sym.Version] = append(vers[sym.Version], sym.Name)
	}

	/* Write config for each version */
	for ver, ns := range vers {
		/* Version */
		fmt.Printf("VERSIONSCRIPT %v {\n", ver)
		fmt.Printf("VERSIONSCRIPT \tglobal:\n")
		/* Symbols with this version */
		for _, n := range ns {
			fmt.Printf("VERSIONSCRIPT \t\t%v;\n", n)
		}
		fmt.Printf("VERSIONSCRIPT };\n")
	}
}
