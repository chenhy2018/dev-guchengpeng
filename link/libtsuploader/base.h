#ifndef __BASE_H__
#define __BASE_H__

#include <string.h>
#include <stdlib.h>
#include <assert.h>
#include <sys/time.h>
#include <errno.h>
#include "log.h"

#define MKTAG(a,b,c,d) ((a) | ((b) << 8) | ((c) << 16) | ((unsigned)(d) << 24))
#define MKERRTAG(a, b, c, d) (-(int)MKTAG(a, b, c, d))

#define TK_NO_MEMORY MKERRTAG('N','M','E','M')
#define TK_TIMEOUT MKERRTAG('T','M','O','T')
#define TK_NO_PUSH MKERRTAG('N','P','S','H')
#define TK_MUTEX_ERROR MKERRTAG('M','U','T','X')
#define TK_COND_ERROR MKERRTAG('C','O','N','D')
#define TK_THREAD_ERROR MKERRTAG('X','C','S','B')

#endif
