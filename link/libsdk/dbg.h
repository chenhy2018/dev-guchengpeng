// Last Update:2018-05-31 15:49:35
/**
 * @file dbg.h
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-27
 */

#ifndef DBG_H
#define DBG_H

#define SDK_DBG 1

#if SDK_DBG
#define DBG_BASIC() printf("[%s %s() +%d] ", __FILE__, __FUNCTION__, __LINE__)
#define DBG_LOG(args...) DBG_BASIC();printf(args)
#define DBG_LINE(args...) DBG_LOG("++++++++++\n")
#define DBG_ERROR(args...) DBG_BASIC();printf("###error#### ");printf(args)
#else
#define DBG_BASIC() 
#define DBG_LOG(args...) 
#define DBG_LINE(args...) 
#define DBG_ERROR(args...) 
#endif

#if SDK_DBG
void DumpUAList();
void DbgBacktrace();
#else
#define DumpUAList()
#define DbgBacktrace()
#endif

#endif  /*DBG_H*/
