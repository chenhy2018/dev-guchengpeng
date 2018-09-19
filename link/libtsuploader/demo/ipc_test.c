#include <stdio.h>
#include <assert.h>
#include <sys/time.h>
#include <unistd.h>
#include <pthread.h>
#include <stdlib.h>

#include "tsuploaderapi.h"
#include "devsdk.h"
#include "ipc_test.h"
#include "log2file.h"
#include "dbg.h"
#include "media_cfg.h"
#include "cfg_parse.h"
#include "socket_logging.h"
//#include "mymalloc.h"

/* global variable */
MediaStreamConfig gAjMediaStreamConfig;
static DevSdkAudioType gAudioType =  AUDIO_TYPE_AAC;
static int gKodoInitOk = 0;
static char gTestToken[1024] = { 0 };
static Config gIpcConfig;
static unsigned char gMovingDetect = 0;
static TsMuxUploader *pMainUploader;
static TsMuxUploader *pSubUploader;
static struct cfg_struct *cfg;
static char gMainStreamDeviceId[128] = { 0 };
static char gSubStreamDeviceId[128] = { 0 };
static char gLogFile[128] = { 0 };

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
/*
 * TODO: config read from config file, ex: ipc.conf
 * */
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

    cfg = cfg_init();

    if (cfg_load(cfg,"/tmp/oem/app/ipc.conf") < 0) {
        fprintf(stderr,"Unable to load ipc.conf\n");
    }
    //cfg_dump( cfg );
    logOutput = cfg_get( cfg, "LOG_OUTPUT" );
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

    if ( last ) {
     //   printf("last = %d, logOutput = %d\n", last, gIpcConfig.logOutput );
        if ( last != gIpcConfig.logOutput ) {
            last = gIpcConfig.logOutput;
            printf("%s %s %d reinit the logger, logOutput = %s\n", __FILE__, __FUNCTION__, __LINE__, logOutput );
            LoggerInit( gIpcConfig.logPrintTime, gIpcConfig.logOutput, gIpcConfig.logFile, gIpcConfig.logVerbose );
        }
    } else {
        last = logOutput;
    }

    logFile = cfg_get( cfg, "LOG_FILE" );
    strcpy( gLogFile, logFile );
    gIpcConfig.logFile = gLogFile;
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

int VideoGetFrameCb( int streamno, char *_pFrame,
                   int _nLen, int _nIskey, double _dTimeStamp,
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex,
                   void *_pContext)
{
    static int first = 1;

    if ( first ) {
        printf("%s thread id = %d\n", __FUNCTION__, pthread_self() );
        first = 0;
    }

    if ( !gKodoInitOk ) {
        ReportKodoInitError( "main stream","kodo not init" );
        return 0;
    }

    if ( gIpcConfig.movingDetection && !gMovingDetect ) {
        ReportKodoInitError( "main stream", "not detect moving" );
        return 0;
    }


    TraceTimeStamp( TYPE_VIDEO, _dTimeStamp, "main stream" );
    PushVideo(pMainUploader, _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );

    return 0;
}

int SubStreamVideoGetFrameCb( int streamno, char *_pFrame,
                   int _nLen, int _nIskey, double _dTimeStamp,
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex,
                   void *_pContext)
{
    static int first = 1;

    if ( first ) {
        printf("%s thread id = %d\n", __FUNCTION__, pthread_self() );
        first = 0;
    }
    if ( !gKodoInitOk ) {
        ReportKodoInitError( "sub stream","kodo not init" );
        return 0;
    }

    if ( gIpcConfig.movingDetection && !gMovingDetect ) {
        ReportKodoInitError( "sub stream", "not detect moving" );
        return 0;
    }


    TraceTimeStamp( TYPE_VIDEO, _dTimeStamp, "sub stream" );
    PushVideo(pSubUploader, _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );

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
        printf("%s thread id = %d\n", __FUNCTION__, pthread_self() );
        isfirst = 0;
    }

    if ( !gKodoInitOk ) {
        ReportKodoInitError("main stream", "gKodoInitOk");
        return 0;
    }

    if ( gIpcConfig.movingDetection && !gMovingDetect ) {
        ReportKodoInitError("main stream", "gMovingDetect");
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

    ret = PushAudio( pMainUploader, _pFrame, _nLen, (int64_t)timeStamp );
    if ( ret != 0 ) {
        DBG_ERROR("ret = %d\n", ret );
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
        printf("%s thread id = %d\n", __FUNCTION__, pthread_self() );
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

    ret = PushAudio( pSubUploader, _pFrame, _nLen, (int64_t)timeStamp );
    if ( ret != 0 ) {
        DBG_ERROR("ret = %d\n", ret );
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
    dev_sdk_start_video( 0, 0, VideoGetFrameCb, &context );
    if ( gIpcConfig.multiChannel ) {
        dev_sdk_start_video( 0, 1, SubStreamVideoGetFrameCb, &context );
    }
    dev_sdk_get_AudioConfig( &audioConfig );
    DBG_LOG("audioConfig.audioEncode.enable = %d\n", audioConfig.audioEncode.enable );
    if ( audioConfig.audioEncode.enable ) {
        dev_sdk_start_audio_play( gAudioType );
        dev_sdk_start_audio( 0, 0, AudioGetFrameCb, NULL );
        if ( gIpcConfig.multiChannel ) {
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
    AvArg avArg;

    DBG_LOG("start to init kodo\n");
    if ( gAudioType == AUDIO_TYPE_AAC ) {
        avArg.nAudioFormat = TK_AUDIO_AAC;
        avArg.nChannels = 1;
        avArg.nSamplerate = 16000;
    } else {
        avArg.nAudioFormat = TK_AUDIO_PCMU;
        avArg.nChannels = 1;
        avArg.nSamplerate = 8000;
    }
    avArg.nVideoFormat = TK_VIDEO_H264;

    SetLogLevelToDebug();
    SetAk( gIpcConfig.ak );
    SetSk( gIpcConfig.sk );

    //计算token需要，所以需要先设置
    SetBucketName( gIpcConfig.bucketName );

    for ( i=0; i<gIpcConfig.tokenRetryCount; i++ ) {
        ret = GetUploadToken( gTestToken, sizeof(gTestToken), NULL );
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

    ret = InitUploader();
    if (ret != 0) {
        DBG_LOG("InitUploader error, ret = %d\n", ret );
        return ret;
    }

    sprintf( gMainStreamDeviceId, "%s%s", gAjMediaStreamConfig.rtmpConfig.server, "a" );
    sprintf( gSubStreamDeviceId, "%s%s", gAjMediaStreamConfig.rtmpConfig.server, "b" );

    UserUploadArg userUploadArg;
    memset(&userUploadArg, 0, sizeof(userUploadArg));
    userUploadArg.pToken_ = gTestToken;
    userUploadArg.nTokenLen_ = strlen(gTestToken);
    userUploadArg.pDeviceId_ = gMainStreamDeviceId;
    userUploadArg.nDeviceIdLen_ = strlen(gMainStreamDeviceId);
    userUploadArg.nUploaderBufferSize = 512;

    ret = CreateAndStartAVUploader(&pMainUploader, &avArg, &userUploadArg);
    if (ret != 0) {
        DBG_LOG("CreateAndStartAVUploader error, ret = %d\n", ret );
        return ret;
    }

    /* sub stream */
    if ( gIpcConfig.multiChannel ) {
        userUploadArg.pDeviceId_ = gSubStreamDeviceId;
        ret = CreateAndStartAVUploader(&pSubUploader, &avArg, &userUploadArg);
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

    while( 1 ) {
        sleep( gIpcConfig.tokenUploadInterval );// 59 minutes
        memset(gTestToken, 0, sizeof(gTestToken));
        ret = GetUploadToken(gTestToken, sizeof(gTestToken), NULL );
        if ( ret != 0 ) {
            DBG_ERROR("GetUploadToken error, ret = %d\n", ret );
            return NULL;
        }
        DBG_LOG("token:%s\n", gTestToken);
        ret = UpdateToken(pMainUploader, gTestToken, strlen(gTestToken));
        if (ret != 0) {
            DBG_ERROR("UpdateToken error, ret = %d\n", ret );
            return NULL;
        }

        if ( gIpcConfig.multiChannel ) {
            ret = UpdateToken(pSubUploader, gTestToken, strlen(gTestToken));
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
        ret = GetUploadToken(  gTestToken, sizeof(gTestToken), NULL );
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

void SdkLogCallback( char *log )
{
    DBG_LOG( log );
}

int AlarmCallback(ALARM_ENTRY alarm, void *pcontext)
{
    if ( alarm.code == ALARM_CODE_MOTION_DETECT ) {
        DBG_LOG("get event ALARM_CODE_MOTION_DETECT\n");
        gMovingDetect = 1;
    } else if ( alarm.code == ALARM_CODE_MOTION_DETECT_DISAPPEAR ) {
        DBG_LOG("get event ALARM_CODE_MOTION_DETECT_DISAPPEAR\n");
        gMovingDetect = 0;
    }

    return 0;
}

void *ConfigUpdateTask( void *param )
{
    for (;;) {
        UpdateConfig();
        sleep( gIpcConfig.configUpdateInterval );
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
    MyMallocInit();
    UpdateConfig();
    WaitForNetworkOk();
    printf("gIpcConfig.logFile = %s\n", gIpcConfig.logFile );
    LoggerInit( gIpcConfig.logPrintTime, gIpcConfig.logOutput, gIpcConfig.logFile, gIpcConfig.logVerbose );
    /* sdk log callback */
    SetLogCallback( SdkLogCallback );

    DBG_LOG("compile tile : %s %s \n", __DATE__, __TIME__ );

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

    for (;; ) {
        sleep( gIpcConfig.heartBeatInterval );
        DBG_LOG("[ %s ] [ HEART BEAT] main thread is running\n", gAjMediaStreamConfig.rtmpConfig.server );
    }

    DeInitIPC();
    DestroyAVUploader(&pMainUploader);
    DestroyAVUploader(&pSubUploader);
    UninitUploader();

    return 0;
}

