// Last Update:2018-08-16 15:30:56
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

#define BASIC() printf("[ %s %s() %d ] ", __FILE__, __FUNCTION__, __LINE__ )
#define DBG_ERROR(args...) BASIC();printf(args)
#define DBG_LOG(args...) BASIC();printf(args)

char *host = "39.107.247.14";
int port = 8089;
int gsock = 0;

int socket_init()
{
    struct sockaddr_in server;
    char message[1000] , server_reply[2000];

    gsock = socket(AF_INET , SOCK_STREAM , 0);
    if (gsock == -1) {
        DBG_LOG("Could not create socket\b");
    }
    DBG_LOG("Socket created\n");

    server.sin_addr.s_addr = inet_addr( host );
    server.sin_family = AF_INET;
    server.sin_port = htons( port );

    if (connect(gsock , (struct sockaddr *)&server , sizeof(server)) < 0) {
        DBG_ERROR("connect failed. Error\n");
        gsock = -1;
        return -1;
    }

    DBG_LOG("connet to %s:%d sucdefully\n", host, port  );
    return 0;
}

int log_send( char *message )
{
    if ( gsock > 0 ) {
        if( send(gsock , message , strlen(message) , 0) < 0) {
            DBG_LOG("Send failed");
            return -1;
        }
    }

    return 0;
}

int report_status( int code )
{
    static int total = 0, error = 0;
    char message[512] = { 0 };

    memset( message, 0, sizeof(message) );
    if ( code != 200 ) {
        error ++;
    }
    total++;
    sprintf( message, "upload ts total : %d error : %d percent %%%d\n", total, error, error/total*100 ); 
    log_send( message );

    return 0;
}

