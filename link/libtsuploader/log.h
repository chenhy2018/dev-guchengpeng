#ifndef __LOG_H__
#define __LOG_H__

#include <stdarg.h>
#include <stdio.h>


#define _s_l_(x) #x
#define _str_line_(x) _s_l_(x)
#define __STR_LINE__ _str_line_(__LINE__)

#define LOG_LEVEL_TRACE 1
#define LOG_LEVEL_DEBUG 2
#define LOG_LEVEL_INFO 3
#define LOG_LEVEL_WARN 4
#define LOG_LEVEL_ERROR 5

extern int nLogLevel;

#define logtrace(fmt,...) \
        if (LOG_LEVEL_TRACE >= nLogLevel) printf(__FILE__ ":" __STR_LINE__ "[T]: " fmt "\n", ##__VA_ARGS__)
#define logdebug(fmt,...) \
        if (LOG_LEVEL_DEBUG >= nLogLevel) printf(__FILE__ ":" __STR_LINE__ "[D]: " fmt "\n", ##__VA_ARGS__)
#define loginfo(fmt,...) \
        if (LOG_LEVEL_INFO >= nLogLevel) printf( __FILE__ ":" __STR_LINE__ "[I]: " fmt "\n", ##__VA_ARGS__)
#define logwarn(fmt,...) \
        if (LOG_LEVEL_WARN >= nLogLevel) printf( __FILE__ ":" __STR_LINE__ "[W]: " fmt "\n", ##__VA_ARGS__)
#define logerror(fmt,...) \
        if (LOG_LEVEL_ERROR >= nLogLevel) printf(__FILE__ ":" __STR_LINE__ "[E]: " fmt "\n", ##__VA_ARGS__)

void SetLogLevelToTrace();
void SetLogLevelToDebug();
void SetLogLevelToInfo();
void SetLogLevelToWarn();
void SetLogLevelToError();

void Log(int nLevel, char * pFmt, ...);


#endif
