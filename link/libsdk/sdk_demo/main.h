// Last Update:2018-07-05 11:38:35
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

extern unsigned StreamStatus();
extern int GetCallId();
extern AccountID GetAccountId();
extern void StartStream();
extern void StopStream();
extern int AppStatus();
extern void AppQuit();
extern void EnableSocketServer();

#endif  /*MAIN_H*/
