#include <stdio.h>
#include <assert.h>
#include <sys/time.h>
#include <unistd.h>
#include <pthread.h>
#include <stdlib.h>

#include "tsuploaderapi.h"
#include "localkey.h"
#include "devsdk.h"
#include "ipc_test.h"
#include "log2file.h"
#include "dbg.h"
#include "media_cfg.h"
#include "cfg_parse.h"
#include "socket_logging.h"
#include "queue.h"
#include "mymalloc.h"

/* global variable */
MediaStreamConfig gAjMediaStreamConfig;
static DevSdkAudioType gAudioType =  AUDIO_TYPE_AAC;
static int gKodoInitOk = 0;
static char gTestToken[1024] = { 0 };
static char gSubToken[1024] = { 0 };
static Config gIpcConfig;
static unsigned char gMovingDetect = 0;
static LinkTsMuxUploader *pMainUploader;
static LinkTsMuxUploader *pSubUploader;
static struct cfg_struct *cfg;
static char gMainStreamDeviceId[128] = { 0 };
static char gSubStreamDeviceId[128] = { 0 };
static char gLogFile[128] = { 0 };
static Queue *pVideoMainStreamCache = NULL;
static Queue *pVideoSubStreamCache = NULL;
static Queue *pAudioSubStreamCache = NULL;
static Queue *pAudioMainStreamCache = NULL;
static __thread int count = STREAM_CACHE_SIZE;

int GetKodoInitSts()
{
    return gKodoInitOk;
}

Config *GetConfig()
{
    return &gIpcConfig;
}

int GetAudioType()
{
    return gAudioType;
}

int GetMovingDetectSts()
{
    return gMovingDetect;
}

void SetOutputType( int output )
{
    gIpcConfig.logOutput = output;
}

int GetOutputType()
{
    return gIpcConfig.logOutput;
}

void SetMovingDetection( int enable )
{
    gIpcConfig.movingDetection = enable;
}

int GetMovingDetection()
{
    return gIpcConfig.movingDetection;
}

void SetUpdateFrom( int updateFrom )
{
    gIpcConfig.updateFrom = updateFrom;
}

void SetCache( int enable )
{
    if ( gIpcConfig.openCache != enable ) {
        gIpcConfig.openCache = enable;
    }
}

void InitConfig()
{
    gIpcConfig.logOutput = OUTPUT_CONSOLE;
    gIpcConfig.logVerbose = 0;
    gIpcConfig.logPrintTime = 1;
    gIpcConfig.timeStampPrintInterval = TIMESTAMP_REPORT_INTERVAL;
    gIpcConfig.heartBeatInterval = 50;
    gIpcConfig.logFile = "/tmp/oem/tsupload.log";
    gIpcConfig.tokenUploadInterval = 3540;
    gIpcConfig.tokenRetryCount = TOKEN_RETRY_COUNT;
    gIpcConfig.bucketName = "ipcamera";
    gIpcConfig.ak = "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ";
    gIpcConfig.sk = "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS";
    gIpcConfig.movingDetection = 1;
    gIpcConfig.configUpdateInterval = 10;
    gIpcConfig.multiChannel = 1;
    gIpcConfig.openCache = 1;
    gIpcConfig.cacheSize = STREAM_CACHE_SIZE;
    gIpcConfig.updateFrom = UPDATE_FROM_SOCKET;
}

void LoadConfig()
{
    cfg = cfg_init();

    if (cfg_load(cfg,"/tmp/oem/app/ipc.conf") < 0) {
        fprintf(stderr,"Unable to load ipc.conf\n");
    }
}

void UpdateConfig()
{
    const char *logOutput = NULL;
    const char *logFile = NULL;
    static int last = 0;
    const char *movingDetect = NULL;
    const char *cache = NULL;

    cfg = cfg_init();

    if (cfg_load(cfg,"/tmp/oem/app/ipc.conf") < 0) {
        fprintf(stderr,"Unable to load ipc.conf\n");
    }
    //cfg_dump( cfg );
    logOutput = cfg_get( cfg, "LOG_OUTPUT" );
    if ( logOutput ) {
        if ( strcmp( logOutput, "socket") == 0 ) {
            gIpcConfig.logOutput = OUTPUT_SOCKET;
        } else if ( strcmp(logOutput, "console" ) == 0 ) {
            gIpcConfig.logOutput = OUTPUT_CONSOLE;
        } else if ( strcmp( logOutput, "mqtt") == 0  ) {
            gIpcConfig.logOutput = OUTPUT_MQTT;
        } else if ( strcmp ( logOutput, "file") == 0 ) {
            gIpcConfig.logOutput = OUTPUT_FILE;
        } else {
            gIpcConfig.logOutput = OUTPUT_SOCKET;
        }
    }

    if ( last ) {
     //   printf("last = %d, logOutput = %d\n", last, gIpcConfig.logOutput );
        if ( last != gIpcConfig.logOutput ) {
            last = gIpcConfig.logOutput;
            printf("%s %s %d reinit the logger, logOutput = %s\n", __FILE__, __FUNCTION__, __LINE__, logOutput );
            LoggerInit( gIpcConfig.logPrintTime, gIpcConfig.logOutput, gIpcConfig.logFile, gIpcConfig.logVerbose );
        }
    } else {
        last = gIpcConfig.logOutput;
    }

    logFile = cfg_get( cfg, "LOG_FILE" );
    if ( logFile ) {
        strcpy( gLogFile, logFile );
        gIpcConfig.logFile = gLogFile;
    }

    movingDetect = cfg_get( cfg, "MOUTION_DETECTION" );
    if ( movingDetect ) {
        if ( strcmp( movingDetect, "1" ) == 0 ) {
            if ( gIpcConfig.movingDetection != 1 ) {
                gIpcConfig.movingDetection = 1;
                printf("%s %s %d open moving detection\n", __FILE__, __FUNCTION__, __LINE__ );
                DBG_LOG("open moving detection\n");
            }
        } else {
            if ( gIpcConfig.movingDetection != 0 ) {
                gIpcConfig.movingDetection = 0;
                printf("%s %s %d close moving detection\n", __FILE__, __FUNCTION__, __LINE__ );
                DBG_LOG("close moving detection\n");
            }
        }
    }
    cache = cfg_get( cfg, "OPEN_CACHE");
    if ( cache ) {
        if ( strcmp( cache, "1") == 0 ) {
            if ( gIpcConfig.openCache != 1 ) {
                gIpcConfig.openCache = 1;
            }
        } else {
            if ( gIpcConfig.openCache != 0 ) {
                gIpcConfig.openCache = 0;
            }
        }
    }
    //printf("logFile = %s\n", logFile );
    cfg_free( cfg );
    //printf("read from ipc.conf, logOutput = %s\n", logOutput );
    //printf("read from ipc.conf, logfile = %s\n", gIpcConfig.logFile );
}


/* function */
void TraceTimeStamp( int type, double _dTimeStamp, char *stream )
{
    double duration = 0;
    char *pType = NULL;
    static double lastTimeStamp = 0, interval = 0;
    static struct timeval start = { 0, 0 }, end = { 0, 0 };

    if ( type == TYPE_VIDEO ) {
        pType = "video";
    } else {
        pType = "audio";
    }

    duration = _dTimeStamp - lastTimeStamp;
    gettimeofday( &end, NULL );
    interval = GetTimeDiff( &start, &end );
    if ( interval >= gIpcConfig.timeStampPrintInterval ) {
        DBG_LOG( "[ %s ] [ %s ] [ %s ] [ timestamp interval ] [ %f ]\n", 
                 gAjMediaStreamConfig.rtmpConfig.server,
                 stream,
                 pType,
                 duration );
        start = end;
    }
    lastTimeStamp = _dTimeStamp;
}

void ReportKodoInitError( char *stream, char *reason )
{
    static struct timeval start = { 0, 0 }, end = { 0, 0 };
    double interval = 0;

    gettimeofday( &end, NULL );
    interval = GetTimeDiff( &start, &end );
    if ( interval >= gIpcConfig.timeStampPrintInterval ) {
        DBG_LOG( "[ %s ] [ %s ] [ %s ]\n", 
                 gAjMediaStreamConfig.rtmpConfig.server,
                 stream,
                 reason
                 );
        start = end;
    }
}

int CacheHandle( Queue *pQueue, LinkTsMuxUploader *pUploader,
                 int _nStreamType, char *_pFrame,
                 int _nLen, int _nIskey, double _dTimeStamp
  )
{
    Frame frame;

    if ( !pQueue || !pUploader  ) {
        return -1;
    }

    memset( &frame, 0, sizeof(frame) );
    frame.data = (char *) malloc ( _nLen );
    if ( !frame.data ) {
        printf("%s %s %d malloc error\n", __FILE__, __FUNCTION__, __LINE__);
        return -1;
    }
    memcpy( frame.data, _pFrame, _nLen );
    frame.len = _nLen;
    frame.timeStamp = _dTimeStamp;
    frame.isKey = _nIskey;
    pQueue->enqueue( pQueue, (void *)&frame, sizeof(frame) );

    if (  pQueue->getSize( pQueue ) == gIpcConfig.cacheSize ) {
        memset( &frame, 0, sizeof(frame) );
        pQueue->dequeue( pQueue, (void *)&frame, NULL );
        if ( !frame.data ) {
            printf("%s %s %d data is NULL\n", __FILE__, __FUNCTION__, __LINE__ );
            return -1;
        }

        if (  gMovingDetect == ALARM_CODE_MOTION_DETECT  ) {
            count = STREAM_CACHE_SIZE;
            if ( _nStreamType == TYPE_VIDEO ) {
                LinkPushVideo( pUploader, frame.data, frame.len, (int64_t)frame.timeStamp, frame.isKey, 0 );
            } else {
                LinkPushAudio( pUploader, frame.data, frame.len, (int64_t)frame.timeStamp );
            }
        } else if ( gMovingDetect == ALARM_CODE_MOTION_DETECT_DISAPPEAR ) {
            if ( count-- > 0 ) {
                if ( _nStreamType == TYPE_VIDEO ) {
                    LinkPushVideo( pUploader, frame.data, frame.len, (int64_t)frame.timeStamp, frame.isKey, 0 );
                } else {
                    LinkPushAudio( pUploader, frame.data, frame.len, (int64_t)frame.timeStamp );
                }
            } else {
                ReportKodoInitError( "main stream", "use cache, not detect moving" );
            }
        } else {
            ReportKodoInitError( "main stream", "use cache, not detect moving" );
        }
        free( frame.data );
    }

    return 0;
}

int VideoGetFrameCb( int streamno, char *_pFrame,
                   int _nLen, int _nIskey, double _dTimeStamp,
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex,
                   void *_pContext)
{
    static int first = 1;

    if ( first ) {
        printf("%s thread id = %d\n", __FUNCTION__, (int)pthread_self() );
        first = 0;
    }

    if ( !gKodoInitOk ) {
        ReportKodoInitError( "main stream","kodo not init" );
        return 0;
    }

    TraceTimeStamp( TYPE_VIDEO, _dTimeStamp, "main stream" );

    if ( gIpcConfig.movingDetection ) {
        if ( gIpcConfig.openCache && pVideoMainStreamCache ) {
            CacheHandle( pVideoMainStreamCache, pMainUploader, TYPE_VIDEO, _pFrame, _nLen, _nIskey,  _dTimeStamp );
        } else if ( gMovingDetect == ALARM_CODE_MOTION_DETECT ) {
            LinkPushVideo(pMainUploader, _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );
        } else if (gMovingDetect == ALARM_CODE_MOTION_DETECT_DISAPPEAR )  {
            ReportKodoInitError( "main stream video", "not detect moving" );
        } else {
            /* do nothing */
        }
    } else {
        LinkPushVideo(pMainUploader, _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );
    }

    return 0;
}

int SubStreamVideoGetFrameCb( int streamno, char *_pFrame,
                   int _nLen, int _nIskey, double _dTimeStamp,
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex,
                   void *_pContext)
{
    static int first = 1;

    if ( first ) {
        printf("%s thread id = %d\n", __FUNCTION__, (int)pthread_self() );
        first = 0;
    }
    if ( !gKodoInitOk ) {
        ReportKodoInitError( "sub stream","kodo not init" );
        return 0;
    }

    TraceTimeStamp( TYPE_VIDEO, _dTimeStamp, "sub stream" );
    if ( gIpcConfig.movingDetection ) {
        if ( gIpcConfig.openCache && pVideoSubStreamCache ) {
            CacheHandle( pVideoSubStreamCache, pSubUploader, TYPE_VIDEO, _pFrame, _nLen, _nIskey,  _dTimeStamp );
        } else if ( gMovingDetect == ALARM_CODE_MOTION_DETECT ) {
            LinkPushVideo(pSubUploader, _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );
        } else if (gMovingDetect == ALARM_CODE_MOTION_DETECT_DISAPPEAR )  {
            ReportKodoInitError( "sub stream video", "not detect moving" );
        } else {
            /* do nothing */
        }
    } else {
        LinkPushVideo(pSubUploader, _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );
    }

    return 0;
}

int AudioGetFrameCb( char *_pFrame, int _nLen, double _dTimeStamp,
                     unsigned long _nFrameIndex, void *_pContext )
{
    static int first = 1;
    static double localTimeStamp = 0, timeStamp = 0;
    int ret = 0;
    static int isfirst = 1;

    if ( isfirst ) {
        printf("%s thread id = %d\n", __FUNCTION__, (int)pthread_self() );
        isfirst = 0;
    }

    if ( !gKodoInitOk ) {
        ReportKodoInitError("main stream audio", "gKodoInitOk");
        return 0;
    }

/**
 * the bug of an-jia
 * timestamp is not conrrect
 * so, we use the timestamp of ourself
 */
    if ( first ) {
        localTimeStamp = _dTimeStamp;
        first = 0;
    } else {
        localTimeStamp += G711_TIMESTAMP_INTERVAL;
    }

    if ( gAudioType == AUDIO_TYPE_AAC ) {
        timeStamp = _dTimeStamp;
    } else {
        timeStamp = localTimeStamp;
    }

    TraceTimeStamp( TYPE_AUDIO, _dTimeStamp, "main stream" );

    if ( gIpcConfig.movingDetection ) {
        if ( gIpcConfig.openCache && pAudioMainStreamCache ) {
            CacheHandle( pAudioMainStreamCache, pMainUploader, TYPE_AUDIO, _pFrame, _nLen, 0,  _dTimeStamp );
        } else if ( gMovingDetect == ALARM_CODE_MOTION_DETECT ) {
            ret = LinkPushAudio( pMainUploader, _pFrame, _nLen, (int64_t)timeStamp );
            if ( ret != 0 ) {
                DBG_ERROR("ret = %d\n", ret );
            }
        } else if (gMovingDetect == ALARM_CODE_MOTION_DETECT_DISAPPEAR )  {
            ReportKodoInitError( "main stream audio", "not detect moving" );
        } else {
            /* do nothing */
        }
    } else {
        ret = LinkPushAudio( pMainUploader, _pFrame, _nLen, (int64_t)timeStamp );
        if ( ret != 0 ) {
            DBG_ERROR("ret = %d\n", ret );
        }
    }

    return 0;
}

int SubStreamAudioGetFrameCb( char *_pFrame, int _nLen, double _dTimeStamp,
                     unsigned long _nFrameIndex, void *_pContext )
{
    static int first = 1;
    static double localTimeStamp = 0, timeStamp = 0;
    int ret = 0;
    static int isfirst = 1;

    if ( isfirst ) {
        printf("%s thread id = %d\n", __FUNCTION__, (int)pthread_self() );
        isfirst = 0;
    }

    if ( !gKodoInitOk ) {
        ReportKodoInitError("sub stream", "gKodoInitOk");
        return 0;
    }

    if ( gIpcConfig.movingDetection && !gMovingDetect ) {
        ReportKodoInitError("sub stream", "gMovingDetect");
        return 0;
    }

/**
 * the bug of an-jia
 * timestamp is not conrrect
 * so, we use the timestamp of ourself
 */
    if ( first ) {
        localTimeStamp = _dTimeStamp;
        first = 0;
    } else {
        localTimeStamp += G711_TIMESTAMP_INTERVAL;
    }

    if ( gAudioType == AUDIO_TYPE_AAC ) {
        timeStamp = _dTimeStamp;
    } else {
        timeStamp = localTimeStamp;
    }

    TraceTimeStamp( TYPE_AUDIO, _dTimeStamp, "sub stream" );

    if ( gIpcConfig.movingDetection ) {
        if ( gIpcConfig.openCache && pAudioSubStreamCache ) {
            CacheHandle( pAudioSubStreamCache, pSubUploader, TYPE_AUDIO, _pFrame, _nLen, 0,  _dTimeStamp );
        } else if ( gMovingDetect == ALARM_CODE_MOTION_DETECT ) {
            ret = LinkPushAudio( pSubUploader, _pFrame, _nLen, (int64_t)timeStamp );
            if ( ret != 0 ) {
                DBG_ERROR("ret = %d\n", ret );
            }
        } else if (gMovingDetect == ALARM_CODE_MOTION_DETECT_DISAPPEAR )  {
            ReportKodoInitError( "sub stream", "not detect moving" );
        } else {
            /* do nothing */
        }
    } else {
        ret = LinkPushAudio( pSubUploader, _pFrame, _nLen, (int64_t)timeStamp );
        if ( ret != 0 ) {
            DBG_ERROR("ret = %d\n", ret );
        }
    }

    return 0;
}

static int InitIPC( )
{
    static int context = 1;
    int s32Ret = 0;
    AudioConfig audioConfig;

    DBG_LOG("start to init IPC\n");
    s32Ret = dev_sdk_init( DEV_SDK_PROCESS_APP );
    if ( s32Ret < 0 ) {
        DBG_ERROR("dev_sdk_init error, s32Ret = %d\n", s32Ret );
        printf("dev_sdk_init error, s32Ret = %d\n", s32Ret );
        return -1;
    }
    GetMediaStreamConfig(&gAjMediaStreamConfig);
    sleep( 2 );
    SendFileName( gAjMediaStreamConfig.rtmpConfig.server );
    pVideoMainStreamCache = NewQueue();
    dev_sdk_start_video( 0, 0, VideoGetFrameCb, &context );
    if ( gIpcConfig.multiChannel ) {
        pVideoSubStreamCache = NewQueue();
        dev_sdk_start_video( 0, 1, SubStreamVideoGetFrameCb, &context );
    }
    dev_sdk_get_AudioConfig( &audioConfig );
    DBG_LOG("audioConfig.audioEncode.enable = %d\n", audioConfig.audioEncode.enable );
    if ( audioConfig.audioEncode.enable ) {
        pAudioMainStreamCache = NewQueue();
        dev_sdk_start_audio_play( gAudioType );
        dev_sdk_start_audio( 0, 0, AudioGetFrameCb, NULL );
        if ( gIpcConfig.multiChannel ) {
            pAudioSubStreamCache = NewQueue();
            dev_sdk_start_audio( 0, 1, SubStreamAudioGetFrameCb, NULL );
        }
    } else {
        DBG_ERROR("not enabled\n");
    }

    return 0;
}

static int DeInitIPC()
{
    dev_sdk_stop_video( 0, 1 );
    dev_sdk_stop_audio( 0, 1 );
    dev_sdk_stop_audio_play();
    dev_sdk_release();
    return 0;
}

int InitKodo()
{
    int ret = 0, i=0;
    LinkMediaArg mediaArg;
    char url[1024] = "http://47.105.118.51:8086/qiniu/upload/token/"; 

    DBG_LOG("start to init kodo\n");
    if ( gAudioType == AUDIO_TYPE_AAC ) {
        mediaArg.nAudioFormat = LINK_AUDIO_AAC;
        mediaArg.nChannels = 1;
        mediaArg.nSamplerate = 16000;
    } else {
        mediaArg.nAudioFormat = LINK_AUDIO_PCMU;
        mediaArg.nChannels = 1;
        mediaArg.nSamplerate = 8000;
    }
    mediaArg.nVideoFormat = LINK_VIDEO_H264;

    LinkSetLogLevel(LINK_LOG_LEVEL_DEBUG);
    LinkSetAk( gIpcConfig.ak );
    LinkSetSk( gIpcConfig.sk );

    //计算token需要，所以需要先设置
    LinkSetBucketName( gIpcConfig.bucketName );

    strncat( url, gAjMediaStreamConfig.rtmpConfig.server, strlen(gAjMediaStreamConfig.rtmpConfig.server) );
    strncat( url, "a", 1 );
    DBG_LOG("url = %s\n", url );

    for ( i=0; i<gIpcConfig.tokenRetryCount; i++ ) {
        ret = LinkGetUploadToken( gTestToken, sizeof(gTestToken), url );
        if ( ret != 0 ) {
            DBG_ERROR("GetUploadToken error, ret = %d, retry = %d\n", ret, i );
            continue;
        } else {
            break;
        }
    }

    if ( i == gIpcConfig.tokenRetryCount ) {
        DBG_LOG( "GetUploadToken error, ret = %d\n", ret );
    }

    DBG_LOG("gAjMediaStreamConfig.rtmpConfig.streamid = %s\n", gAjMediaStreamConfig.rtmpConfig.streamid);
    DBG_LOG("gAjMediaStreamConfig.rtmpConfig.server = %s\n", gAjMediaStreamConfig.rtmpConfig.server );

    ret = LinkInitUploader();
    if (ret != 0) {
        DBG_LOG("InitUploader error, ret = %d\n", ret );
        return ret;
    }

    sprintf( gMainStreamDeviceId, "%s%s", gAjMediaStreamConfig.rtmpConfig.server, "a" );
    sprintf( gSubStreamDeviceId, "%s%s", gAjMediaStreamConfig.rtmpConfig.server, "b" );

    LinkUserUploadArg userUploadArg;
    memset(&userUploadArg, 0, sizeof(userUploadArg));
    userUploadArg.pToken_ = gTestToken;
    userUploadArg.nTokenLen_ = strlen(gTestToken);
    userUploadArg.pDeviceId_ = gMainStreamDeviceId;
    userUploadArg.nDeviceIdLen_ = strlen(gMainStreamDeviceId);
    userUploadArg.nUploaderBufferSize = 512;

    ret = LinkCreateAndStartAVUploader(&pMainUploader, &mediaArg, &userUploadArg);
    if (ret != 0) {
        DBG_LOG("CreateAndStartAVUploader error, ret = %d\n", ret );
        return ret;
    }

    /* sub stream */
    if ( gIpcConfig.multiChannel ) {
        userUploadArg.pDeviceId_ = gSubStreamDeviceId;
        ret = LinkCreateAndStartAVUploader(&pSubUploader, &mediaArg, &userUploadArg);
        if (ret != 0) {
            DBG_LOG("CreateAndStartAVUploader error, ret = %d\n", ret );
            return ret;
        }
    }
    DBG_LOG("[ %s ] kodo init ok\n", gAjMediaStreamConfig.rtmpConfig.server );
    gKodoInitOk = 1;
    return 0;
}

static void * upadateToken() {
    int ret = 0;
    char url[1024] = "http://47.105.118.51:8086/qiniu/upload/token/";
    char subUrl[1024] = "http://47.105.118.51:8086/qiniu/upload/token";

    strncat( url, gAjMediaStreamConfig.rtmpConfig.server, strlen(gAjMediaStreamConfig.rtmpConfig.server) );
    strncat( url, "a", 1 );
    DBG_LOG("url = %s\n", url );
    strncat( subUrl, gAjMediaStreamConfig.rtmpConfig.server, strlen(gAjMediaStreamConfig.rtmpConfig.server) );
    strncat( subUrl, "a", 1 );

    while( 1 ) {
        sleep( gIpcConfig.tokenUploadInterval );// 59 minutes
        memset(gTestToken, 0, sizeof(gTestToken));
        ret = LinkGetUploadToken(gTestToken, sizeof(gTestToken), url );
        if ( ret != 0 ) {
            DBG_ERROR("GetUploadToken error, ret = %d\n", ret );
            return NULL;
        }
        DBG_LOG("token:%s\n", gTestToken);
        ret = LinkUpdateToken(pMainUploader, gTestToken, strlen(gTestToken));
        if (ret != 0) {
            DBG_ERROR("UpdateToken error, ret = %d\n", ret );
            return NULL;
        }

        if ( gIpcConfig.multiChannel ) {
            memset( gSubToken, 0, sizeof(gSubToken));
            ret = LinkGetUploadToken( gSubToken, sizeof(gSubToken), url );
            if ( ret != 0 ) {
                DBG_ERROR("GetUploadToken error, ret = %d\n", ret );
                return NULL;
            }
            DBG_LOG("token:%s\n", gSubToken);
            ret = LinkUpdateToken(pSubUploader, gSubToken, strlen(gSubToken));
            if (ret != 0) {
                DBG_ERROR("UpdateToken error, ret = %d\n", ret );
                return NULL;
            }
        }
    }
    return NULL;
}

int StartTokenUpdateTask()
{
    pthread_t updateTokenThread;
    pthread_attr_t attr;
    int ret = 0;

    pthread_attr_init ( &attr );
    pthread_attr_setdetachstate ( &attr, PTHREAD_CREATE_DETACHED );
    ret = pthread_create( &updateTokenThread, &attr, upadateToken, NULL );
    if (ret != 0 ) {
        DBG_ERROR("create update token thread fail\n");
        return ret;
    }
    pthread_attr_destroy (&attr);

    return 0;
}

int WaitForNetworkOk()
{
    int i = 0, ret = 0;

    for ( i=0; i<gIpcConfig.tokenRetryCount; i++ ) {
        ret = LinkGetUploadToken(  gTestToken, sizeof(gTestToken), NULL );
        if ( ret != 0 ) {
            DBG_ERROR("GetUploadToken error, ret = %d, retry = %d\n", ret, i );
            continue;
        } else {
            break;
        }
    }

    if ( i == gIpcConfig.tokenRetryCount ) {
        DBG_LOG("GetUploadToken error, ret = %d\n", ret );
        return -1;
    }

    return 0;
}

void SdkLogCallback(int nLogLevel, char *log )
{
    DBG_LOG( log );
}

int AlarmCallback(ALARM_ENTRY alarm, void *pcontext)
{
    if ( alarm.code == ALARM_CODE_MOTION_DETECT ) {
        DBG_LOG("get event ALARM_CODE_MOTION_DETECT\n");
        gMovingDetect = alarm.code;
        /*gMovingDetect = 1;*/
    } else if ( alarm.code == ALARM_CODE_MOTION_DETECT_DISAPPEAR ) {
        DBG_LOG("get event ALARM_CODE_MOTION_DETECT_DISAPPEAR\n");
        gMovingDetect = alarm.code;
        /*gMovingDetect = 0;*/
    }

    return 0;
}

void *ConfigUpdateTask( void *param )
{
    for (;;) {
        if ( gIpcConfig.updateFrom == UPDATE_FROM_FILE ) {
            UpdateConfig();
            sleep( gIpcConfig.configUpdateInterval );
        }
    }
}

void StartConfigUpdateTask()
{
    pthread_t thread;

    pthread_create( &thread, NULL, ConfigUpdateTask, NULL );
}

int main()
{
    int ret = 0;

    InitConfig();
    UpdateConfig();
    WaitForNetworkOk();
    printf("gIpcConfig.logFile = %s\n", gIpcConfig.logFile );
    LoggerInit( gIpcConfig.logPrintTime, gIpcConfig.logOutput, gIpcConfig.logFile, gIpcConfig.logVerbose );
    /* sdk log callback */
    LinkSetLogCallback( SdkLogCallback );

    StartTokenUpdateTask();
    StartConfigUpdateTask();
    /* 
     * ipc need to receive server command
     * so socket logging task must been started
     *
     * */
    StartSocketLoggingTask();

    ret = InitIPC();
    if ( 0 != ret ) {
        DBG_ERROR("InitIPC() fail\n");
    }

    if ( gIpcConfig.movingDetection ) {
        dev_sdk_register_callback( AlarmCallback, NULL, NULL, NULL );
    }

    ret = InitKodo();
    if ( ret < 0 ) {
        DBG_ERROR("ret = %d\n",ret );
    } else {
        DBG_ERROR("ret is 0\n");
    }

    DBG_LOG("compile time : %s %s \n", __DATE__, __TIME__ );
    for (;; ) {
        sleep( gIpcConfig.heartBeatInterval );
        DBG_LOG("[ %s ] [ HEART BEAT] main thread is running\n", gAjMediaStreamConfig.rtmpConfig.server );
    }

    DeInitIPC();
    LinkDestroyAVUploader(&pMainUploader);
    LinkDestroyAVUploader(&pSubUploader);
    LinkUninitUploader();

    return 0;
}

