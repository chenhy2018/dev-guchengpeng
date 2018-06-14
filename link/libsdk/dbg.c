// Last Update:2018-05-31 15:50:49
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
#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include "log.h"
#if SDK_DBG
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
        sprintf(output, "%s %s %s [line +%d] %s %s", date, file, function, line, debug, printf_buf);
        debugFunc(output);
        free(date);
        free(output);
}

void SetLogFunc(LogFunc *func)
{
        debugFunc = func;
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
        pj_log_set_level(level);
}