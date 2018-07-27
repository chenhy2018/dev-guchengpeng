#ifndef __BASE_H__
#define __BASE_H__

#include <string.h>
#include <stdlib.h>
#include <assert.h>
#include <sys/time.h>
#include <errno.h>
#include "log.h"
#ifndef __APPLE__
#include <stdint.h>
#endif

typedef enum {
        TK_VIDEO_H264,
        TK_VIDEO_H265
}TkVideoFormat;
typedef enum {
        TK_AUDIO_PCMU,
        TK_AUDIO_PCMA,
        TK_AUDIO_AAC
}TkAudioFormat;

typedef enum {
        TK_UPLOAD_INIT,
        TK_UPLOAD_FAIL,
        TK_UPLOAD_OK
}UploadState;

typedef struct _AvArg{
        int nAudioFormat;
        int nChannels;
        int nSamplerate;
        int nVideoFormat;
} AvArg;

#define TK_STREAM_UPLOAD 1

#define MKTAG(a,b,c,d) ((a) | ((b) << 8) | ((c) << 16) | ((unsigned)(d) << 24))
#define MKERRTAG(a, b, c, d) (-(int)MKTAG(a, b, c, d))

#define TK_NO_MEMORY MKERRTAG('N','M','E','M')
#define TK_TIMEOUT MKERRTAG('T','M','O','T')
#define TK_NO_PUSH MKERRTAG('N','P','S','H')
#define TK_MUTEX_ERROR MKERRTAG('M','U','T','X')
#define TK_COND_ERROR MKERRTAG('C','O','N','D')
#define TK_THREAD_ERROR MKERRTAG('X','C','S','B')
#define TK_ARG_ERROR MKERRTAG('A','R','G','E')

#endif
