// Last Update:2018-09-19 14:25:00
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
#include <errno.h>
#include "socket_logging.h"
#include "devsdk.h"
#include "Queue.h"
#include "dbg.h"
#include "ipc_test.h"

#define BASIC() printf("[ %s %s() %d ] ", __FILE__, __FUNCTION__, __LINE__ )
#define ARRSZ(arr) sizeof(arr)/sizeof(arr[0])


extern MediaStreamConfig gAjMediaStreamConfig;
static socket_status gStatus;
static Queue *gLogQueue;

void CmdHnadleDump( char *param );
void CmdHnadleLogStop( char *param );

char *host = "47.105.118.51";
int port = 8090;
int gsock = 0;
static DemoCmd gCmds[] =
{
    { "dump", CmdHnadleDump },
    { "logstop", CmdHnadleLogStop },
};

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
    printf("connet to %s:%d sucdefully, gsock = %d\n", host, port, gsock  );
    return 0;
}

void SendFileName( char *logfile )
{
    char message[256] = { 0 };
    int ret = 0;

    sprintf( message, "%s.log", logfile );
    printf("send file name %s\n", message );
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
        if ( !gStatus.logStop ) {
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
                    shutdown( gsock, SHUT_RDWR );
                    close( gsock );
                    gsock = -1;
                }
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
                SendFileName( gAjMediaStreamConfig.rtmpConfig.server );
                sleep(2);
                printf("queue size = %d\n", gLogQueue->getSize( gLogQueue )) ;
            }
        } else {
            sleep( 3 );
        }
    }
    return NULL;
}

void *SimpleSshTask( void *param )
{
    ssize_t ret = 0;
    char buffer[1024] = { 0 };
    int i = 0;

    for (;;) {
        if ( gStatus.connecting ) {
            memset( buffer, 0, sizeof(buffer) );
            //printf("%s %s %d gsock = %d\n", __FILE__, __FUNCTION__, __LINE__, gsock );
            ret = recv( gsock, buffer, 1024, 0 );
            printf("buffer = %s", buffer );
            if ( ret < 0 ) {
                if ( errno != 107 ) {
                    printf("recv error, errno = %d\n", errno );
                    continue;
                }
                sleep(2);
            }
            for ( i=0; i<ARRSZ(gCmds); i++ ) {
                char *res = NULL;
                printf("buffer = %s\n", buffer );
                printf("gCmds[i].cmd = %s\n", gCmds[i].cmd );
                res = strstr( buffer, gCmds[i].cmd );
                if ( res ) {
                    gCmds[i].pCmdHandle( buffer );
                    break;
                }
            }
            if ( i == ARRSZ(gCmds) ) {
                printf("unknow command %s", buffer );
            }

        } else {
            sleep( 3 );
        }
    }
    return NULL;
}

void StartSocketLoggingTask()
{
    static pthread_t log = 0, cmd;

    if ( !log ) {
        printf("%s %s %d start socket logging thread\n", __FILE__, __FUNCTION__, __LINE__);
        pthread_create( &log, NULL, SocketLoggingTask, NULL );
        pthread_create( &cmd, NULL, SimpleSshTask, NULL );
    }
}

void CmdHnadleDump( char *param )
{
    char buffer[1024] = { 0 } ;
    int ret = 0;
    Config *pConfig = GetConfig();

    printf("get command dump\n");
    sprintf( buffer, "%s", "Config :\n" );
    sprintf( buffer+strlen(buffer), "logOutput = %d\n", pConfig->logOutput );
    sprintf( buffer+strlen(buffer), "logFile = %s\n", pConfig->logFile );
    sprintf( buffer+strlen(buffer), "movingDetection = %d\n", pConfig->movingDetection );
    sprintf( buffer+strlen(buffer), "gKodoInitOk = %d\n", GetKodoInitSts() );
    sprintf( buffer+strlen(buffer), "gMovingDetect = %d\n", GetMovingDetectSts() );
    sprintf( buffer+strlen(buffer), "gAudioType = %d\n", GetAudioType() );
    sprintf( buffer+strlen(buffer), "queue = %d\n", gLogQueue->getSize( gLogQueue ) );
    ret = send(gsock , buffer , strlen(buffer) , MSG_NOSIGNAL );// MSG_NOSIGNAL ignore SIGPIPE signal
    if(  ret < 0 ) {
        printf("Send failed, ret = %d, %s\n", ret, strerror(errno) );
    }

    
}

void CmdHnadleLogStop( char *param )
{
    printf("get command log stop\n");
    gStatus.logStop = 1;
}



