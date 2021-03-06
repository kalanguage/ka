#define _GNU_SOURCE

#ifdef _WIN32
//because winsock2 must be included before windows.h
#include <winsock2.h>
#include <windows.h>
#endif

//include all of the syscalls
#include "syscalls/read.h"
#include "syscalls/write.h"
#include "syscalls/open.h"
#include "syscalls/close.h"
#include "syscalls/fstat.h"
#include "syscalls/lseek.h"
#include "syscalls/read_writev.h"
#include "syscalls/pipe.h"
#include "syscalls/mem.h"
#include "syscalls/select.h"
#include "syscalls/sched_yield.h"
#include "syscalls/dup.h"
#include "syscalls/pause.h"
#include "syscalls/pid.h"
#include "syscalls/socket.h"
#include "syscalls/execv.h"
#include "syscalls/exit.h"
#include "syscalls/fsync.h"
#include "syscalls/ftrunc.h"
#include "syscalls/dir.h"
#include "syscalls/file.h"
#include "syscalls/systime.h"
#include "syscalls/chroot.h"
#include "syscalls/host.h"
/////////////////////////////
