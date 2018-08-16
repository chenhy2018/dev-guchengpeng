// Last Update:2018-08-16 17:17:01
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
    int retrying;
    int retry_count;
} socket_status;

extern int log_send( char *message );
extern int log_send( char *message );
extern int report_status( int code );

#endif  /*SOCKET_LOGGING_H*/
