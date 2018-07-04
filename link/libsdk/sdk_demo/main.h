// Last Update:2018-07-03 10:46:30
/**
 * @file main.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-05-23
 */

#ifndef MAIN_H
#define MAIN_H

#include "sdk_interface.h"

//#define DBG_BASIC() printf("[%s %s() +%d] ", __FILE__, __FUNCTION__, __LINE__)
//#define DBG_LOG(args...) DBG_BASIC();printf(args)
//#define LINE() DBG_LOG("++++++++++++++++++++++++\n")
extern unsigned StreamStatus();
extern int GetCallId();
extern AccountID GetAccountId();
void StartStream();
void StopStream();

#endif  /*MAIN_H*/
