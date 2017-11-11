Slightly Less Brief Theory
==========================
The way this attack works is presented below.  Some details have been omitted
in the interest of brevity.

Linking
-------
When a dynamically-linked ELF binary is loaded, the linker reads the ELF header
to work out which libraries need to be loaded, and for which exported symbols.
The libraries can be seen with `ldd(1)`:
```bash
$ ldd /usr/bin/uptime
        linux-vdso.so.1 =>  (0x00007ffcb69fc000)
        libprocps.so.4 => /lib/x86_64-linux-gnu/libprocps.so.4 (0x00007f3c041a3000)
        libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f3c03dd9000)
        libsystemd.so.0 => /lib/x86_64-linux-gnu/libsystemd.so.0 (0x00007f3c04561000)
        /lib64/ld-linux-x86-64.so.2 (0x00007f3c043ca000)
        libselinux.so.1 => /lib/x86_64-linux-gnu/libselinux.so.1 (0x00007f3c03bb7000)
        librt.so.1 => /lib/x86_64-linux-gnu/librt.so.1 (0x00007f3c039af000)
        liblzma.so.5 => /lib/x86_64-linux-gnu/liblzma.so.5 (0x00007f3c0378d000)
        libgcrypt.so.20 => /lib/x86_64-linux-gnu/libgcrypt.so.20 (0x00007f3c034ac000)
        libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f3c0328f000)
        libpcre.so.3 => /lib/x86_64-linux-gnu/libpcre.so.3 (0x00007f3c0301f000)
        libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007f3c02e1b000)
        libgpg-error.so.0 => /lib/x86_64-linux-gnu/libgpg-error.so.0 (0x00007f3c02c07000)
```

In this case, `uptime(1)` is loading several libraries.  In reality,
`uptime(1)` itself is only loading a couple of them, and those load others,
which may load others, and so on.  Here we can see the name of the libraries
(e.g.  `libprocps.so.4`) and the file on disk where each library is found (e.g.
`/lib/x86_64-linux-gnu/libprocps.so.4`).

Linking, Maliciously
--------------------
If we edit the binary (specifically the ELF headers in the binary), we can
change the name of the which the linker will try to load at runtime.

From:
```plaintext
...snip...
^@^@^@^@^@^@^@^@^@^@^@^@^@libprocps.so.4^@_ITM_deregisterTMCloneTable^@__gmon_s
tart__^@_Jv_RegisterClasses^@_ITM_registerTMCloneTable^@print_uptime^@libc.so.6
^@__printf_chk^@setlocale^@dcgettext^@__stack_chk_fail^@_exit^@program_invocati
on_name^@__errno_location^@__fprintf_chk^@stdout^@fputs^@fclose^@stderr^@getopt
...snip...
```

To:
```plaintext
...snip...
^@^@^@^@^@^@^@^@^@^@^@^@^@libkitten.so.4^@_ITM_deregisterTMCloneTable^@__gmon_s
tart__^@_Jv_RegisterClasses^@_ITM_registerTMCloneTable^@print_uptime^@libc.so.6
^@__printf_chk^@setlocale^@dcgettext^@__stack_chk_fail^@_exit^@program_invocati
on_name^@__errno_location^@__fprintf_chk^@stdout^@fputs^@fclose^@stderr^@getopt
...snip...
```

Fortunately, the string "`libprocps.so.4`" only appeared once in the entire
file.  In general, though, it'll be near the top of the file and near a bunch
of function names.  In this example, "`libprocps.so.4`" was changed to
"`libkitten.so.4`".  `ldd(1)` now looks for `libkitten` instead of `libprocps`.
```bash
$ ldd /usr/bin/uptime
        linux-vdso.so.1 =>  (0x00007fff17f53000)
        libkitten.so.4 => not found
        libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f8dfaa17000)
        /lib64/ld-linux-x86-64.so.2 (0x00007f8dfade1000)
```

The linker is now looking for `libkitten`, and, not unreasonably, can't find
it.  Also of note, the other libraries are now absent, due to them having
previously been required by `libprocps`.

Malicious libraries
-------------------
The linker has a set of path through which to search for libraries.  If we
watch the still-modified `/usr/bin/uptime` looking for `libkitten.so.4`, we can
get a good idea of the search path.
```bash
$ strace /usr/bin/uptime 2>&1 | grep libkitten                                                                                                
open("/lib/x86_64-linux-gnu/tls/x86_64/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/lib/x86_64-linux-gnu/tls/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/lib/x86_64-linux-gnu/x86_64/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/lib/x86_64-linux-gnu/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/usr/lib/x86_64-linux-gnu/tls/x86_64/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/usr/lib/x86_64-linux-gnu/tls/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/usr/lib/x86_64-linux-gnu/x86_64/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/usr/lib/x86_64-linux-gnu/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/lib/tls/x86_64/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/lib/tls/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/lib/x86_64/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/lib/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/usr/lib/tls/x86_64/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/usr/lib/tls/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/usr/lib/x86_64/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
open("/usr/lib/libkitten.so.4", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
```
If we make a file named, say, `/lib/libkitten.so.4`, the linker will attempt to
load it instead of the original `libprocps.so.4`.  The linker will give an
error as it still expects to find the symbols that would have been imported
from `libprocps`, but that's easy enough to work out.

Those functions (and required associated library versions) can be found with
`objdump(1)` or with the program in this repository.  A malicious
`libkitten.so.4` can then be crafted to load `libprocps.so.4` (via
`dlopen(3)`), export proxy functions which call libprocps' functions and
additionally perform other malicious tasks.  The end result is a reverse shell
every time someone calls `uptime(1)`.

Do it!
------
Please see [QUICKSTART](./QUICKSTART.md) for a step-by-step guide.
