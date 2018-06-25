// Last Update:2018-06-21 15:24:24
/**
 * @file dbg.c
 * @brief 
 * @author
 * @version 0.1.00
 * @date 2018-05-31
 */

#include "sdk_interface.h"
#include "list.h"
#include "sdk_local.h"
#include "dbg.h"
#include "dbg.h"
#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include "log.h"
#if SDK_DBG

DbgStr callStatusStr[] = 
{
    DBG_STRING( CALL_STATUS_IDLE ),
    DBG_STRING( CALL_STATUS_REGISTERED ),
    DBG_STRING( CALL_STATUS_REGISTER_FAIL ),
    DBG_STRING( CALL_STATUS_INCOMING ),
    DBG_STRING( CALL_STATUS_TIMEOUT ),
    DBG_STRING( CALL_STATUS_ESTABLISHED ),
    DBG_STRING( CALL_STATUS_RING ),
    DBG_STRING( CALL_STATUS_REJECT ),
    DBG_STRING( CALL_STATUS_HANGUP ),
    DBG_STRING( CALL_STATUS_ERROR ),
};

DbgStr sdkRetStr[] = 
{
    DBG_STRING( RET_OK ),
    DBG_STRING( RET_MEM_ERROR ),
    DBG_STRING( RET_ACCOUNT_NOT_EXIST ),
    DBG_STRING( RET_FAIL ),
    DBG_STRING( RET_REGISTERING ),
    DBG_STRING( RET_INIT_ERROR ),
    DBG_STRING( RET_CALL_NOT_EXIST ),
    DBG_STRING( RET_PARAM_ERROR ),
    DBG_STRING( RET_USER_UNAUTHORIZED ),
    DBG_STRING( RET_CALL_INVAILD_CONNECTION ),
    DBG_STRING( RET_TIMEOUT_FROM_SERVER ),
    DBG_STRING( RET_CALL_INVAILD_SDP ),
    DBG_STRING( RET_INTERAL_FAIL ),
    DBG_STRING( RET_REGISTER_TIMEOUT ),
    DBG_STRING( RET_SDK_ALREADY_INITED ),
};

#if 0
void DumpUAList()
{
    UA *pUA = NULL;

    list_for_each_entry( pUA, &UaList.list, list ){
        if ( pUA ) {
            printf("[%02d] account id : \n", pUA->id );
        }
    }
    
}

void DbgBacktrace()
{
    void *array[20];
    size_t size;
    char **strings;
    size_t i;

    size = backtrace (array, 20);
    strings = backtrace_symbols (array, size);

    DBG_LOG ("Obtained %zd stack frames.\n", size);

    for (i = 0; i < size; i++)
        printf ("%s\n", strings[i]);

    free (strings);
}
#endif

char * DbgCallStatusGetStr( CallStatus status )
{
    int i = 0;

    for ( i=0; i<ARRSZ(callStatusStr); i++ ) {
        if ( status == callStatusStr[i].val ) {
            return callStatusStr[i].str;
        }
    }

    return "NULL";
}

char *DbgSdkRetGetStr( ErrorID id )
{
    int i = 0;

    for ( i=0; i<ARRSZ(sdkRetStr); i++ ) {
        if ( id == sdkRetStr[i].val ) {
            return sdkRetStr[i].str;
        }
    }
    return "NULL";
}

#endif

static int  dbgLevel = LOG_DEBUG;    // Default Logging level

static char* getDateString() {
    // Initialize and get current time
    time_t t = time( NULL );

    // Allocate space for date string
    char* date = (char*)malloc( 100 );

    // Format the time correctly
    strftime(date, 100, "[%F %T]", localtime(&t));
    return date;
}

void printData(const char* data)
{
        while (*data) {
                putchar(*data);
                ++data;
        }
}

static LogFunc* debugFunc = printData;
struct  timeval start;
struct  timeval end;

void writeLog(int loglvl, const char* file, const char* function, const int line, const char* format, ... )
{
        if (loglvl > dbgLevel) {
                return;
        }
        //output date
        va_list arg;
        char* date = getDateString();
        char printf_buf[1024];
        //debug level
        char debug[20] = {0};
        switch (loglvl) {
              case  LOG_VERBOSE:
                      strcpy(debug, " [LOG_VERBOSE] ");
                      break;
              case LOG_DEBUG:
                      strcpy(debug, " [LOG_DEBUG] ");
                      break;
              case LOG_WARN:
                      strcpy(debug, " [LOG_WARN] ");
                      break;
              case LOG_INFO:
                      strcpy(debug, " [LOG_INFO] ");
                      break;
              case LOG_ERROR:
                      strcpy(debug, " [LOG_ERROR] ");
                      break;
              case LOG_FATAL:
                      strcpy(debug, " [LOG_FATAL] ");
                      break;
              default:
                      strcpy(debug, " [LOG_INFO] ");
                      break;
        }
        char fun_buf[100];
        sprintf(fun_buf, "%s %s [line +%d] ", file, function, line);
        va_start( arg, format );
        vsprintf(printf_buf, format, arg);
        va_end(arg);
        char *output = (char*)malloc(1024);
        unsigned  long diff;
        gettimeofday(&end,NULL);
        diff = 1000000 * (end.tv_sec-start.tv_sec)+ end.tv_usec-start.tv_usec;
        snprintf(output, 1024, "%s DIFF %ld %s %s [line +%d] %s [pid %ld]  %s", date, diff, file, function, line, debug, pthread_self(), printf_buf);
        debugFunc(output);
        free(date);
        free(output);
        gettimeofday(&start,NULL);
}

void pjLogFunc(int level, const char *data, int len)
{
         if (level > dbgLevel) {
                 return;
         }
         debugFunc(data);
}

void SetLogLevel(int level) {
        dbgLevel = level;
        pj_log_set_log_func(pjLogFunc);
        pj_log_set_level(2);
}
