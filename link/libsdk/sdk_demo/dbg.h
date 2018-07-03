// Last Update:2018-06-29 20:26:46
/**
 * @file dbg.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-26
 */

#ifndef DBG_H
#define DBG_H

#include <stdio.h>


enum {
    LOG_OUTPUT_NONE,
    LOG_OUTPUT_FILE,
    LOG_OUTPUT_CONSOLE,
    LOG_OUTPUT_SOCKET,
};

enum {
    LOG_LEVEL_DEBUG,
    LOG_LEVEL_FATAL,
};

#define BUFFER_SIZE 1024
#define BASIC() printf("[ %s %s +%d ] ", __FILE__, __FUNCTION__, __LINE__)
#define LOG(args...) PrintLog( LOG_LEVEL_DEBUG, __FILE__, __FUNCTION__, __LINE__, args ) 
#define DBG_ERROR(args...) PrintLog( LOG_LEVEL_FATAL, __FILE__, __FUNCTION__, __LINE__, args ) 
#define DBG_LOG(args...) LOG(args)

#define NONE                 "\e[0m"
#define BLACK                "\e[0;30m"
#define L_BLACK              "\e[1;30m"
#define RED                  "\e[0;31m"
#define L_RED                "\e[1;31m"
#define GREEN                "\e[0;32m"
#define L_GREEN              "\e[1;32m"
#define BROWN                "\e[0;33m"
#define YELLOW               "\e[1;33m"
#define BLUE                 "\e[0;34m"
#define L_BLUE               "\e[1;34m"
#define PURPLE               "\e[0;35m"
#define L_PURPLE             "\e[1;35m"
#define CYAN                 "\e[0;36m"
#define L_CYAN               "\e[1;36m"
#define GRAY                 "\e[0;37m"
#define WHITE                "\e[1;37m"

#define BOLD                 "\e[1m"
#define UNDERLINE            "\e[4m"
#define BLINK                "\e[5m"
#define REVERSE              "\e[7m"
#define HIDE                 "\e[8m"
#define CLEAR                "\e[2J"
#define CLRLINE              "\r\e[K" //or "\e[1K\r"



typedef struct {
    int output;
    char *logFile;
    unsigned printTime;
    FILE *fp;
    unsigned logLevel;
} Logger;

extern int PrintLog( unsigned logLevel, const char *file, const char *function, int line, const char *format, ... );
extern int LoggerInit( unsigned printTime, int output, char *pLogFile );
extern int LoggerUnInit();

typedef struct {
    long long int videoFrameCount;
    long long int audioFrameCount;
    FILE *fp;
} StreamDebugInfo;
#define STREAM_DEBUG 1
#define DBG_STREAM_WRITE_FILE 1

#if STREAM_DEBUG
//#define FRAME_NUMBER (10*60*30)// 10min * 60s * 30frames
#define FRAME_NUMBER (100)// 10min * 60s * 30frames
void DbgAddAudioFrameCount();
void DbgAddVideoFrameCount();
void DbgStreamFileOpen();
void DbgWriteVideoFrame( void *frame, int size );
#else
#define DbgAddAudioFrameCount()
#define DbgAddVideoFrameCount()
#define DbgStreamFileOpen()
#define DbgWriteVideoFrame( frame, size )
#endif

#endif  /*DBG_H*/

