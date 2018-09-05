// Last Update:2018-09-04 15:05:07
/**
 * @file dbg.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-09-03
 */
#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include <time.h>
#include <errno.h>
#include <sys/time.h>
#include <unistd.h>

#include "dbg.h"
#include "socket_logging.h"
#include "log2file.h"

static Logger gLogger;

int LoggerInit( unsigned printTime, int output, char *pLogFile, int logVerbose )
{
    struct timeval time;

    memset( &gLogger, 0, sizeof(gLogger) );

    gLogger.output = output;
    gLogger.logFile = pLogFile;
    gLogger.printTime = printTime;
    gLogger.logVerbose = logVerbose;

    switch( output ) {
    case OUTPUT_FILE:
        fileOpen( gLogger.logFile );
        break;
    case OUTPUT_SOCKET:
        socket_init();
        break;
    case OUTPUT_MQTT:
        break;
    case OUTPUT_CONSOLE:
    default:
        break;
    }
    

    return 0;
}


int dbg( unsigned logLevel, const char *file, const char *function, int line, const char *format, ...  )
{
    char buffer[BUFFER_SIZE] = { 0 };
    va_list arg;
    int len = 0;
    char *pTime = NULL;
    char now[200] = { 0 };

    if ( gLogger.printTime ) {
        memset( now, 0, sizeof(now) );
        get_current_time( now );
        len = sprintf( buffer, "[ %s ] ", now );
    }

    if ( gLogger.logVerbose ) {
        len = sprintf( buffer+len, "[ %s %s +%d ] ", file, function, line );
    }

    va_start( arg, format );
    vsprintf( buffer+strlen(buffer), format, arg );
    va_end( arg );

    switch( gLogger.output ) {
    case OUTPUT_FILE:
        writeLog( buffer ); 
        break;
    case OUTPUT_SOCKET:
        log_send( buffer );
        break;
    case OUTPUT_MQTT:
        break;
    case OUTPUT_CONSOLE:
        if ( logLevel == DBG_LEVEL_FATAL ) {
            printf( RED"%s"NONE, buffer );
        } else {
            printf("%s", buffer );
        }
        break;
    default:
        break;
    }

    return 0;

}

