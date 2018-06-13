// Last Update:2018-06-05 17:53:14
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

#if SDK_DBG
void DumpUAList();
void DbgBacktrace();
#else
#define DumpUAList()
#define DbgBacktrace()
#endif

#endif  /*DBG_H*/
