// Last Update:2018-09-14 18:05:32
/**
 * @file socket_logging.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-08-16
 */

#ifndef SOCKET_LOGGING_H
#define SOCKET_LOGGING_H

typedef struct {
    int connecting;
    int retry_count;
} socket_status;

typedef struct {
    char *cmd;
    void (*pCmdHandle)(char *param);
} DemoCmd;

extern int log_send( char *message );
extern int log_send( char *message );
extern int report_status( int code, char *_pFileNmae );
extern int GetTimeDiff( struct timeval *_pStartTime, struct timeval *_pEndTime );
extern int get_current_time( char *now_time );
extern void SendFileName();
extern void StartSocketLoggingTask();

#endif  /*SOCKET_LOGGING_H*/
