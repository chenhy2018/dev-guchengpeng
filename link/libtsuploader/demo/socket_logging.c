// Last Update:2018-09-13 19:29:53
/**
 * @file socket_logging.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-08-16
 */
#include<stdio.h> //DBG_LOG
#include<string.h>    //strlen
#include<sys/socket.h>    //socket
#include<arpa/inet.h> //inet_addr
#include <pthread.h>
#include <errno.h>
#include <time.h>
#include <unistd.h>
#include "socket_logging.h"
#include "devsdk.h"
#include "Queue.h"
#include "dbg.h"

#define BASIC() printf("[ %s %s() %d ] ", __FILE__, __FUNCTION__, __LINE__ )
#define DBG_ERROR(args...) BASIC();printf(args)


extern MediaStreamConfig gAjMediaStreamConfig;
static socket_status gStatus;
static Queue *gLogQueue;

char *host = "47.105.118.51";
int port = 8090;
int gsock = 0;

int socket_init()
{
    struct sockaddr_in server;
    int ret = 0;
    static int first = 1;

    if ( first ) {
        gLogQueue = NewQueue();
        if ( !gLogQueue ) {
            printf("new queue error\n");
            return -1;
        }
        first = 0;
    }

    if ( gsock != -1 ) {
        DBG_LOG("close last sock\n");
        close( gsock );
    }

    gsock = socket(AF_INET , SOCK_STREAM , 0);
    if (gsock == -1) {
        DBG_LOG("Could not create socket\b");
        gStatus.connecting = 0;
        return -1;
    }
    DBG_LOG("Socket created\n");

    server.sin_addr.s_addr = inet_addr( host );
    server.sin_family = AF_INET;
    server.sin_port = htons( port );

    ret = connect(gsock , (struct sockaddr *)&server , sizeof(server));
    if ( ret < 0) {
        DBG_ERROR("connect failed. Error, ret = %d, %s \n", ret , strerror(errno) );
        gStatus.connecting = 0;
        return -1;
    }

    gStatus.connecting = 1;
    printf("connet to %s:%d sucdefully\n", host, port  );
    return 0;
}

void SendFileName()
{
    char message[256] = { 0 };
    int ret = 0;

    sprintf( message, "tsupload_%s.log", gAjMediaStreamConfig.rtmpConfig.server );
    //log_send( message );
    ret = send(gsock , message , strlen(message) , MSG_NOSIGNAL );// MSG_NOSIGNAL ignore SIGPIPE signal
    if(  ret < 0 ) {
        printf("Send failed, ret = %d, %s\n", ret, strerror(errno) );
    }

}

int log_send( char *message )
{
    if ( gLogQueue && gStatus.connecting ) {
    //    printf("message = %s", message );
        gLogQueue->enqueue( gLogQueue, message, strlen(message) );
    }

    return 0;
}

int report_status( int code, char *_pFileName )
{
    static int total = 0, error = 0;
    char message[512] = { 0 };
    char now[200] = { 0 };

    memset( message, 0, sizeof(message) );
    if ( code != 200 ) {
        error ++;
    }
    total++;
    memset( now, 0, sizeof(now) );
    get_current_time( now );
    sprintf( message, "[ %s ] [ %s ] [ cur : %s] [ total : %d ] [ error : %d ] [ percent : %%%f ]\n", 
             now,
             gAjMediaStreamConfig.rtmpConfig.server,
             _pFileName,
             total, error, error*1.0/total*100 ); 
    DBG_LOG("%s", message );

    return 0;
}

int GetTimeDiff( struct timeval *_pStartTime, struct timeval *_pEndTime )
{
    int time = 0;

    if ( _pEndTime->tv_sec < _pStartTime->tv_sec ) {
        return -1;
    }

    if ( _pEndTime->tv_usec < _pStartTime->tv_usec ) {
        time = (_pEndTime->tv_sec - 1 - _pStartTime->tv_sec) +
            ((1000000-_pStartTime->tv_usec) + _pEndTime->tv_usec)/1000000;
    } else {
        time = (_pEndTime->tv_sec - _pStartTime->tv_sec) +
            (_pEndTime->tv_usec - _pStartTime->tv_usec)/1000000;
    }

    return ( time );

}

int get_current_time( char *now_time )
{
    time_t now;
    struct tm *tm_now = NULL;

    time(&now);
    tm_now = localtime(&now);
    strftime( now_time, 200, "%Y-%m-%d %H:%M:%S", tm_now);

    return(0);
}

void *SocketLoggingTask( void *param )
{
    char log[1024] = { 0 };
    int ret = 0;

    for (;;) {
        if ( gStatus.connecting ) {
                if ( gLogQueue ) {
                    memset( log, 0, sizeof(log) );
                    gLogQueue->dequeue( gLogQueue, log, NULL );
                } else {
                    printf("error, gLogQueue is NULL\n");
                    return NULL;
                }
                //printf("log = %s", log);
                ret = send(gsock , log , strlen(log) , MSG_NOSIGNAL );// MSG_NOSIGNAL ignore SIGPIPE signal
                if(  ret < 0 ) {
                    printf("Send failed, ret = %d, %s\n", ret, strerror(errno) );
                    gStatus.connecting = 0;
                }
                sleep(1);
        } else {
            ret = socket_init();
            if ( ret < 0 ) {
                sleep(5);
                gStatus.retry_count ++;
                printf("reconnect retry count %d\n", gStatus.retry_count );
                continue;
            }
            printf("reconnect to %s ok\n", host );
            gStatus.connecting = 1;
            SendFileName();
            sleep(1);
            printf("queue size = %d\n", gLogQueue->getSize( gLogQueue )) ;
        }
    }
    return NULL;
}

void *SimpleSshTask( void *param )
{
    for (;;) {
    }
    return NULL;
}

void StartSocketLoggingTask()
{
    pthread_t log, cmd;

    pthread_create( &log, NULL, SocketLoggingTask, NULL );
    //pthread_create( &cmd, NULL, SimpleSshTask, NULL );
}



