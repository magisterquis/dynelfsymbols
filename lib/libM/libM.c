#define _GNU_SOURCE
#include <dlfcn.h>
#include <err.h>
#include <stdlib.h>

/* Begin version script:
------------------------
VERSIONSCRIPT GLIBC_2.2.5 {
VERSIONSCRIPT 	global:
VERSIONSCRIPT 		modf;
VERSIONSCRIPT };
---------------------
End version script */

/* Constructor prototype */
static void con(void) __attribute__((constructor));

/* Exported function prototypes */
double modf(double x, double *iptr);

/* Pointers to real functions */
static double (*modf_real)(double x, double *iptr);

/* Exported functions which call real functions */
double modf(double x, double *iptr) {return modf_real(x, iptr);}

/* con is called when the library is loaded */
void
con(void)
{
	char *e;

	/* Load the real libm.so.6 */
	if (NULL == dlopen("libm.so.6", RTLD_NOW|RTLD_GLOBAL))
		errx(1, "dlopen: %s", dlerror());

	/* Get hold of the real functions */

	/* modf */
	dlerror();
	modf_real = dlsym(RTLD_NEXT, "modf");
	if (NULL != (e = dlerror()))
		errx(2, "dlsym modf: %s", e);

	/************************************
	 * Further malicious code goes here *
	 ************************************/
}

/*
Slice off the version script and save it to a file named "vs".  After modifying
the ARGS_CHANGEMEs and TYPE_CHANGEMEs to reflect the real function prototypes,
he rest of this C source code can be compiled with a command similar to:

cc -O2 -shared -fPIC -Wl,--version-script=vs -o foo.so foo.c -lc -ldl
*/
