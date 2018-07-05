#include <unistd.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <getopt.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <pthread.h>
#include "main.h"
#include "sdk_interface.h"
#include "dbg.h"
#include "stream.h"
#include "common.h"
#include "command.h"
#include "dev_core.h"

#define POLL_EVENT_TIMEOUT 1000

static char *gLogFile = "/tmp/ipc-stream.log";
static char *gAccountId = "1023";
static char *gPasswd = "6epJRKvx";
static char *gHost = "39.107.247.14";

static struct app {
    char *logfile;
    char *account;
    char *passwd;
    char *host;
    unsigned socketServer;
    unsigned printTime;
    unsigned startStream;
    AccountID accountId;
    int callId;
    unsigned running;
} app;


static int init_options( int argc, char *argv[] )
{
    int c = 0;

    for (;;) {
        static struct option long_options[] =
        {
            { "account", required_argument, 0, 'a' },
            { "passwd", required_argument, 0, 'p' },
            { "host", required_argument, 0, 'h' },
            { "logfile", required_argument, 0, 'l' }, 
            { "start-socket-server", no_argument, 0, 's' },
            { "print-time", no_argument, 0, 't' },
            { 0, 0, 0, 0 }
        };
        int option_index = 0;

        c = getopt_long (argc, argv, "a:p:h:l:st", long_options, &option_index);
        if ( c == -1 ) {
            break;
        }
        switch( c ) {
        case 0:
            if (long_options[option_index].flag != 0)
                break;
            LOG ("option %s", long_options[option_index].name);
            if ( optarg ) {
                LOG (" with arg %s", optarg);
            }
            LOG ("\n");
            break;
        case 'l':
            app.logfile = optarg;
            break;
        case 's':
            app.socketServer = 1;
            break;
        case 'a':
            app.account = optarg;
            break;
        case 'p':
            app.passwd = optarg;
            break;
        case 'h':
            app.host = optarg;
            break;
        case 't':
            app.printTime = 1;
            break;
        case '?':
            break;
        default:
            DBG_ERROR("parse options error\n");
            return -1;
        }
    }

    if (optind < argc) {
        LOG ("non-option ARGV-elements: ");
        while (optind < argc) {
            LOG ("%s ", argv[optind++]);
        }
        LOG ("\n");
    }

    return 0;
}

void StartStream()
{
    app.startStream = 1;
}

void StopStream()
{
    app.startStream = 0;
}

unsigned StreamStatus()
{
    return ( app.startStream );
}

void EnableSocketServer()
{
    app.socketServer = 1;
}

void AppQuit()
{
    app.running = 0;
}

int GetCallId()
{
    return ( app.callId) ;
}

int AppStatus()
{
    return app.running;
}

AccountID GetAccountId()
{
    return ( app.accountId) ;
}

void StartIPC()
{
    ErrorID ret = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    MediaEvent *pMediaEvent = NULL;

    DBG_LOG("start IPC...\n");
    app.accountId = Register( app.account, app.passwd, app.host, app.host, app.host );
    if ( app.accountId >= RET_MEM_ERROR ) {
        DBG_ERROR("Register() error, ret = %d\n", app.accountId );
        return;
    }

    DBG_LOG("accountId = %d for %s\n", app.accountId, app.account );
    while ( app.running ) {
        ret = PollEvent( app.accountId, &type, &pEvent, POLL_EVENT_TIMEOUT );
        if ( ret >= RET_MEM_ERROR ) {
            DBG_ERROR("PollEvent() error, ret = %d\n", ret );
            return;
        }

        if ( ret == RET_RETRY) {
            continue;
        }

        switch( type ) {
        case EVENT_CALL:
            LOG("EVENT_CALL\n");
            if ( pEvent ) {
                pCallEvent = &pEvent->body.callEvent;
                switch( pCallEvent->status ) {
                case CALL_STATUS_INCOMING:
                    DBG_LOG("CALL_STATUS_INCOMING\n");
                    if ( pCallEvent->pFromAccount ) {
                        DBG_LOG("incoming call from account : %s\n", pCallEvent->pFromAccount );
                    }
                    ret = AnswerCall( app.accountId, pCallEvent->callID );
                    if ( ret >= RET_MEM_ERROR ) {
                        DBG_ERROR("AnswerCall() error, ret = %d\n", ret );
                    }
                    break;
                case CALL_STATUS_REGISTERED:
                    DBG_LOG("CALL_STATUS_REGISTERED\n");
                    break;
                case CALL_STATUS_REGISTER_FAIL:
                    DBG_LOG("CALL_STATUS_REGISTER_FAIL\n");
                    break;
                case CALL_STATUS_ESTABLISHED:
                    DBG_LOG("CALL_STATUS_ESTABLISHED\n");
                    app.callId = pCallEvent->callID;
                    break;
                case CALL_STATUS_HANGUP:
                    DBG_LOG("CALL_STATUS_HANGUP\n");
                    StopStream();
                    DbgDumpStream();
                    break;
                case CALL_STATUS_ERROR:
                    DBG_LOG("CALL_STATUS_ERROR\n");
                    break;
                default:
                    DBG_ERROR("wrong status, pCallEvent->status = %d\n", pCallEvent->status );
                    break;
                }
            } else {
                DBG_ERROR("get one event, but event data is NULL\n");
            }
            break;
        case EVENT_DATA:
            LOG("EVENT_DATA\n");
            break;
        case EVENT_MEDIA:
            LOG("EVENT_MEDIA\n");
            if ( pEvent ) {
                pMediaEvent = &pEvent->body.mediaEvent;
                DBG_LOG("midia channel count : %d \n", pMediaEvent->nCount );
                DBG_LOG("callId = %d\n", pMediaEvent->callID );
                DBG_LOG("media 0 stream type : %d\n", pMediaEvent->media[0].streamType );
                DBG_LOG("media 0 codec type : %d\n", pMediaEvent->media[0].codecType );
                DBG_LOG("media 0 sample reate : %d\n", pMediaEvent->media[0].sampleRate );
                DBG_LOG("media 0 channels : %d\n", pMediaEvent->media[0].channels );
                DBG_LOG("media 1 stream type : %d\n", pMediaEvent->media[1].streamType );
                DBG_LOG("media 1 codec type : %d\n", pMediaEvent->media[1].codecType );
                DBG_LOG("media 1 sample reate : %d\n", pMediaEvent->media[1].sampleRate );
                DBG_LOG("media 1 channels : %d\n", pMediaEvent->media[1].channels );
                DbgFrameAmountReset();
                if ( pMediaEvent->nCount ) { // count = 0, sdp negotiation fail 
                    StartStream();
                }
            }
            break;
        case EVENT_MESSAGE:
            LOG("EVENT_MESSAGE\n");
            break;
        default:
            DBG_ERROR("unknow event, type = %d\n", type);
            break;
        }
    }
}

void InitMedia( Media *media )
{
    media[0].streamType = STREAM_VIDEO;
    media[0].codecType = CODEC_H264;
    media[0].sampleRate = 90000;
    media[0].channels = 0;
    media[1].streamType = STREAM_AUDIO;
    media[1].codecType = CODEC_G711A;
    media[1].sampleRate = 8000;
    media[1].channels = 1;
}



int main( int argc, char *argv[] )
{
    ErrorID ret = 0;
    Media media[2];
    int sts = 0;
    pthread_t tid = 0;
    CoreDevice *pDev = NULL;

    DBG_LOG("enter main %s %s ...\n", __DATE__, __TIME__ );
    memset( &app, 0, sizeof(app) );
    app.running = 1;
    if ( init_options( argc, argv ) < 0 ) {
        DBG_LOG("init_options() error\n");
        return 0;
    }

    /* if user not specify the account/passwd/host,
     * use the default
     *  */
    if ( !app.account ) {
        app.account = gAccountId;
    }

    if ( !app.passwd ) {
        app.passwd = gPasswd;
    }

    if ( !app.host ) {
        app.host = gHost;
    }


    sts = LoggerInit( app.printTime, 0, app.logfile );
    if ( sts == 0 ) {
        DBG_LOG("Logger init OK...\n");
    } else {
        DBG_LOG("Logger init FAIL...\n");
        return 0;
    }

    InitMedia( media );
    ret = InitSDK( media, 2 );
    if ( ret >= RET_MEM_ERROR ) {
        DBG_ERROR("initsdk error, ret = %d\n", ret );
        return 0;
    } else {
        DBG_LOG("SDK init OK...\n");
    }

    SetLogLevel( LOG_ERROR );
    setPjLogLevel( 2 );
    pDev = NewCoreDevice();
    if ( !pDev ) {
        DBG_ERROR("get core device error\n");
        return 0;
    }

    if ( pDev->init() < 0 ) {
        DBG_ERROR("init capture device error\n");
        return 0;
    }

    DbgStreamFileOpen();
    pthread_create( &tid, NULL, UserInputHandleThread, NULL );

    StartIPC();

    LoggerUnInit();
    pDev->deInit();

    return 0;
}



