// Last Update:2018-06-29 20:30:47
/**
 * @file dbg.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-26
 */
#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include <time.h>
#include <errno.h>
#include <unistd.h>
#include "dbg.h"

static Logger gLogger;

int LoggerInit( unsigned printTime, int output, char *pLogFile )
{
    gLogger.output = output;
    gLogger.logFile = pLogFile;
    gLogger.printTime = printTime;

    if ( gLogger.logFile ) {
        gLogger.fp = fopen( gLogger.logFile, "w+" );
        if ( gLogger.fp == NULL ) {
            printf("open log file %s error, errno = %d\n", gLogger.logFile, errno );
            return -1;
        } else {
            printf("gLogger.fp = %p\n", gLogger.fp );
        }
    }

    return 0;
}

int LoggerUnInit()
{
    if ( gLogger.fp )
        fclose( gLogger.fp );

    return 0;
}

int PrintLog( unsigned logLevel, const char *file, const char *function, int line, const char *format, ... )
{
    char buffer[BUFFER_SIZE] = { 0 };
    va_list arg;
    int len = 0;
    char *pTime = NULL;

    if ( gLogger.printTime ) {
        time_t result;

        result = time( NULL );
        pTime = asctime( localtime(&result));
        pTime[strlen(pTime) -1 ] = 0;
        len = sprintf( buffer, "[ %s ] ", pTime );
    }
    len = sprintf( buffer+len, "[ %s %s +%d ] ", file, function, line );

    va_start( arg, format );
    vsprintf( buffer+strlen(buffer), format, arg );
    va_end( arg );

    if ( gLogger.fp ) {
        size_t written = 0;

        if ( logLevel == LOG_LEVEL_FATAL ) {
            sprintf( buffer+strlen(buffer), " [ *** ERROR *** ] ");
        }
        written = fwrite( buffer, strlen(buffer), 1, gLogger.fp );
        if ( written != 1 ) {
            printf("fwrite error, errno = %d, written = %d\n", errno, written );
            return -1;
        }
        fflush( gLogger.fp );
    } else {
        switch ( logLevel ) {
        case LOG_LEVEL_DEBUG:
            printf( "%s", buffer );
            break;
        case LOG_LEVEL_FATAL:
            printf( RED"%s"NONE, buffer );
            break;
        default:
            break;
        }
    }

    return 0;
}


StreamDebugInfo gDebugInfo;
void DbgAddVideoFrameCount()
{
    gDebugInfo.videoFrameCount ++;
    if ( !(gDebugInfo.videoFrameCount%FRAME_NUMBER) ) {
        DBG_LOG("already send %lld vido frames\n", gDebugInfo.videoFrameCount );
    }
}

void DbgAddAudioFrameCount()
{
    gDebugInfo.audioFrameCount ++;
    if ( !(gDebugInfo.audioFrameCount%FRAME_NUMBER) ) {
        DBG_LOG("already send %lld audio frames\n", gDebugInfo.audioFrameCount );
    }
}

char *gStreamFile = "./400c.264";
void DbgStreamFileOpen()
{
    gDebugInfo.fp = fopen( gStreamFile, "w+" );
    if ( NULL == gDebugInfo.fp ) {
        DBG_ERROR("fopen error, errno = %d\n", errno );
    }
}

void DbgWriteVideoFrame( void *frame, int size )
{
    size_t ret = 0;

    if ( gDebugInfo.videoFrameCount <= 1000 ) {
        DBG_LOG("size = %d\n", size );
        ret = fwrite( frame, size, 1, gDebugInfo.fp );
        if ( ret != 1 ) {
            DBG_ERROR("fwrite error, errno = %d\n", errno );
        }
        fflush( gDebugInfo.fp );
    }
}



