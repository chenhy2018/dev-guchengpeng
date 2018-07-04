// Last Update:2018-07-04 11:43:24
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
#include <sys/time.h>
#include <unistd.h>
#include "dbg.h"

static Logger gLogger;
static StreamDebugInfo gDebugInfo;

void LoggerSetPrintTime( int enable )
{
    gLogger.printTime = enable;
}

void DbgSetVideoFrameThreshod( int threshold )
{
    gDebugInfo.videoThreshold = threshold;
}

int LoggerInit( unsigned printTime, int output, char *pLogFile )
{
    struct timeval time;

    memset( &gDebugInfo, 0, sizeof(gDebugInfo) );
    memset( &gLogger, 0, sizeof(gLogger) );

    gettimeofday( &time, NULL );

    gLogger.output = output;
    gLogger.logFile = pLogFile;
    gLogger.printTime = printTime;

    gDebugInfo.videoThreshold = FRAME_NUMBER;
    gDebugInfo.startTime = time.tv_sec*1000 + time.tv_usec/1000;
    gDebugInfo.endTime = gDebugInfo.startTime;

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


void DbgAddVideoFrameCount()
{
    gDebugInfo.videoFrameCount ++;
    if ( !(gDebugInfo.videoFrameCount%FRAME_NUMBER) ) {
        DBG_LOG("already send %lld video frames, total size %lld bytes\n", gDebugInfo.videoFrameCount, gDebugInfo.videoStreamTotalBytes );
    }
}

void DbgAddAudioFrameCount()
{
    gDebugInfo.audioFrameCount ++;
    if ( !(gDebugInfo.audioFrameCount%FRAME_NUMBER) ) {
        DBG_LOG("already send %lld audio frames\n", gDebugInfo.audioFrameCount );
    }
}

void DbgGetFrameAmount( unsigned long long int *pVideoCount, unsigned long long int *pAudioCount )
{
    *pVideoCount = gDebugInfo.videoFrameCount;
    *pAudioCount = gDebugInfo.audioFrameCount;
}

void DbgFrameAmountReset()
{
    gDebugInfo.videoFrameCount = 0;
    gDebugInfo.audioFrameCount = 0;
    gDebugInfo.videoStreamTotalBytes = 0;
    gDebugInfo.audioStreamTotalBytes = 0;
}

char *gStreamFile = "./400c.h264";
void DbgStreamFileOpen()
{
    gDebugInfo.fp = fopen( gStreamFile, "w+" );
    if ( NULL == gDebugInfo.fp ) {
        DBG_ERROR("fopen error, errno = %d\n", errno );
    }
}

void DbgGetVideoFrame( void *frame, int size )
{
    size_t ret = 0;
    int duration = 0;
    struct timeval time;

    gettimeofday( &time, NULL );

    gDebugInfo.videoFrameCount ++;
    gDebugInfo.videoStreamTotalBytes += size;
    gDebugInfo.startTime = gDebugInfo.endTime;
    gDebugInfo.endTime = time.tv_sec*1000 + time.tv_usec/1000;
    duration = (gDebugInfo.endTime - gDebugInfo.startTime)/1000;
    if ( !(gDebugInfo.videoFrameCount%gDebugInfo.videoThreshold) ) {
        if ( duration ) {
            DBG_LOG("already send %lld video frames, total size %lld bytes\n", 
                    gDebugInfo.videoFrameCount, gDebugInfo.videoStreamTotalBytes );
        } else {
            DBG_LOG("already send %lld video frames, total size %lld bytes, bit rate %d bytes/s\n",
                    gDebugInfo.videoFrameCount, gDebugInfo.videoStreamTotalBytes );
        }
    }

    if ( gDebugInfo.videoFrameCount <= 1000 ) {
//        DBG_LOG("size = %d\n", size );
        ret = fwrite( frame, size, 1, gDebugInfo.fp );
        if ( ret != 1 ) {
            DBG_ERROR("fwrite error, errno = %d\n", errno );
        }
        fflush( gDebugInfo.fp );
    }
}

void DbgDumpStream()
{
    DBG_LOG("total send %lld video frames, total size %lld bytes\n", gDebugInfo.videoFrameCount, gDebugInfo.videoStreamTotalBytes );
    DBG_LOG("total send %lld audio frames, total size %lld bytes\n", gDebugInfo.audioFrameCount, gDebugInfo.audioStreamTotalBytes );
}



