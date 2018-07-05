// Last Update:2018-07-05 11:40:03
/**
 * @file command.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-07-05
 */
#include <string.h>
#include "main.h"
#include "dbg.h"
#include "common.h"

typedef struct {
    char *cmd;
    void (*pCmdHandle)(char *param);
} DemoCmd;

void CmdHandleQuit( char *param );
void CmdHanleHelp( char *param );
void CmdHandleServer( char *param );
void CmdHandleStartStream( char *param );
void CmdHandleDump( char *param );
void CmdHandleStopStream( char *param );
void CmdHandleLogCloseTime( char *param );
void CmdHandleLogOpenTime( char *param );
void CmdHandleDbgReset( char *param );
void CmdHandleVideoThreshold( char *param );

static DemoCmd gCmds[] =
{
    { "quit", CmdHandleQuit },
    { "help", CmdHanleHelp },
    { "start-socket-server", CmdHandleServer },
    { "start-stream", CmdHandleStartStream },
    { "dump", CmdHandleDump },
    { "stop-stream", CmdHandleStopStream },
    { "log-open-time", CmdHandleLogOpenTime },
    { "log-close_time", CmdHandleLogCloseTime },
    { "dbg-reset", CmdHandleDbgReset },
    { "video-threshold", CmdHandleVideoThreshold },
};

void CmdHandleVideoThreshold( char *param )
{
}

void CmdHandleDbgReset( char *param )
{
    DBG_LOG("frame amount reset\n");
    DbgFrameAmountReset();
}

void CmdHandleLogCloseTime( char *param )
{
    LoggerSetPrintTime( 0 );
    DBG_LOG("disable print time\n");
}

void CmdHandleLogOpenTime( char *param )
{
    LoggerSetPrintTime( 1 );
    DBG_LOG("enable print time\n");
}

void CmdHandleStopStream( char *param )
{
    DBG_LOG("stop stream\n");
    StopStream();
    DbgFrameAmountReset();
}

void CmdHandleQuit( char *param )
{
    AppQuit();
}

void CmdHandleDump( char *param )
{
    DbgDumpStream();
}

void CmdHanleHelp( char *param )
{
    printf("\thelp - this usage\n"
           "\tquit - quit the app\n"
           "\tstart-stream - start the rtp stream\n"
           "\tstop-stream - stop the rtp stream\n"
           "\tdump - show how much frames has been sent\n"
           "\tlog-open-time - log print time\n"
           "\tlog-close-time - log disable time\n"
           "\tdbg-reset - clear the frame total size and amount\n"
           "\tstart-socket-server - start socket server\n");
}

void CmdHandleServer( char *param )
{
    EnableSocketServer();
}

void CmdHandleStartStream( char *param )
{
    DBG_LOG("start the rtp stream\n");
    StartStream();
}

void *UserInputHandleThread( void *arg )
{
    char buffer[1024] = { 0 };
    char *ret = NULL;
    int i = 0;

    while( AppStatus() ) {
        printf("sdk_demo >");
        ret = fgets( buffer, sizeof(buffer), stdin );
        if ( NULL == ret ) {
            DBG_ERROR("fgets error, errno = %d\n", errno);
            continue;
        }

        if ( strcmp( buffer, "\n") == 0 ||
             strcmp( buffer, "\r") == 0 ) {
            continue;
        }
        for ( i=0; i<ARRSZ(gCmds); i++ ) {
            ret = strstr( buffer, gCmds[i].cmd );
            if ( ret ) {
                gCmds[i].pCmdHandle( buffer );
                break;
            } 
        }
        if ( i == ARRSZ(gCmds) ) {
            DBG_ERROR("unknow command %s", buffer );
        }
    }

    return NULL;
}

