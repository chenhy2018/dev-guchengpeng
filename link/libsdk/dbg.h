// Last Update:2018-06-21 15:24:37
/**
 * @file dbg.h
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-27
 */

#ifndef DBG_H
#define DBG_H

#ifdef WITH_P2P
#include "sdk_interface_p2p.h"
#else
#include "sdk_interface.h"
#endif

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

typedef void LogFunc(const char *data);
void SetDebugLogLevel(int level);
#if SDK_DBG
void DumpUAList();
void DbgBacktrace();
char * DbgCallStatusGetStr(CallStatus status);
char *DbgSdkRetGetStr( ErrorID id );
#else
#define DumpUAList()
#define DbgBacktrace()
#define DbgCallStatusGetStr( CallStatus status )
#define DbgSdkRetGetStr( ErrorID id )
#endif

#endif  /*DBG_H*/
