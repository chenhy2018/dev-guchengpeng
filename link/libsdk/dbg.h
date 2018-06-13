// Last Update:2018-06-13 14:42:48
/**
 * @file dbg.h
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-27
 */

#ifndef DBG_H
#define DBG_H

#include "sdk_interface.h"
#define SDK_DBG 1

#define DBG_STRING( v ) { v, #v }
#define ARRSZ( arr ) (sizeof(arr)/sizeof(arr[0]))
#if SDK_DBG
#define DBG_DEBUG(args...) 
#define DBG_LOG(args...) writeLog(LOG_INFO, __FILE__, __FUNCTION__, __LINE__, args)
#define DBG_LINE(args...) DBG_LOG("++++++++++\n")
#define DBG_ERROR(args...) writeLog(LOG_ERROR, __FILE__, __FUNCTION__, __LINE__, args)
#define DBG_VAL(val) DBG_LOG(#val" = %d\n", val)
#define DBG_STR(str) DBG_LOG(#str" = %s\n", str)
void writeLog (int loglvl, const char* file, const char* function, const int line, const char* str, ... );

#else
#define DBG_LOG(args...) 
#define DBG_LINE(args...) 
#define DBG_ERROR(args...) 
#define DBG_VAL(val) 
#define DBG_STR(str)
#endif

typedef struct {
    int val;
    char *str;
} DbgStr;

#if SDK_DBG
void DumpUAList();
void DbgBacktrace();
char * DbgCallStatusGetStr(CallStatus status);
#else
#define DumpUAList()
#define DbgBacktrace()
#define DbgCallStatusGetStr( CallStatus status )
#endif

#endif  /*DBG_H*/
