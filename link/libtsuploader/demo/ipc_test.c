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

/* global variable */
static DevSdkAudioType gAudioType =  AUDIO_TYPE_AAC;
static int gKodoInitOk = 0;
static char gTestToken[1024] = { 0 };
static Config gIpcConfig;
static unsigned char gMovingDetect = 0;
MediaStreamConfig gAjMediaStreamConfig = { 0 };
TsMuxUploader *pTsMuxUploader;

/*
 * TODO: config read from config file, ex: ipc.conf
 * */
void InitConfig()
{
    gIpcConfig.logOutput = OUTPUT_SOCKET;
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
}


/* function */
void TraceTimeStamp( int type, double _dTimeStamp )
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
        DBG_LOG( "[ %s ] [ %s ] [ timestamp interval ] [ %f ]\n", 
                 gAjMediaStreamConfig.rtmpConfig.server,
                 pType,
                 duration );
        start = end;
    }
    lastTimeStamp = _dTimeStamp;
}

void ReportKodoInitError( char *reason )
{
    static struct timeval start = { 0, 0 }, end = { 0, 0 };
    double interval = 0;

    gettimeofday( &end, NULL );
    interval = GetTimeDiff( &start, &end );
    if ( interval >= gIpcConfig.timeStampPrintInterval ) {
        DBG_LOG( "[ %s ] [ %s ]\n", 
                 gAjMediaStreamConfig.rtmpConfig.server,
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
    if ( !gKodoInitOk ) {
        ReportKodoInitError( "kodo not init" );
        return 0;
    }

    if ( gIpcConfig.movingDetection && !gMovingDetect ) {
        ReportKodoInitError( "not detect moving" );
        return 0;
    }


    TraceTimeStamp( TYPE_VIDEO, _dTimeStamp );
    PushVideo(pTsMuxUploader, _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );

    return 0;
}

int AudioGetFrameCb( char *_pFrame, int _nLen, double _dTimeStamp,
                     unsigned long _nFrameIndex, void *_pContext )
{
    static int first = 1;
    static double localTimeStamp = 0, timeStamp = 0;
    static double min=0, max=0;
    int ret = 0;

    if ( !gKodoInitOk ) {
        ReportKodoInitError("gKodoInitOk");
        return 0;
    }

    if ( gIpcConfig.movingDetection && !gMovingDetect ) {
        ReportKodoInitError("gMovingDetect");
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

    if ( gAudioType = AUDIO_TYPE_AAC ) {
        timeStamp = _dTimeStamp;
    } else {
        timeStamp = localTimeStamp;
    }

    TraceTimeStamp( TYPE_AUDIO, _dTimeStamp );

    ret = PushAudio(pTsMuxUploader, _pFrame, _nLen, (int64_t)timeStamp );
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
    int ret = 0;

    DBG_LOG("start to init IPC\n");
    s32Ret = dev_sdk_init( DEV_SDK_PROCESS_APP );
    if ( s32Ret < 0 ) {
        DBG_ERROR("dev_sdk_init error, s32Ret = %d\n", s32Ret );
        return -1;
    }
    GetMediaStreamConfig(&gAjMediaStreamConfig);
    SendFileName();
    ret = dev_sdk_start_video( 0, 0, VideoGetFrameCb, &context );
    dev_sdk_get_AudioConfig( &audioConfig );
    DBG_LOG("audioConfig.audioEncode.enable = %d\n", audioConfig.audioEncode.enable );
    if ( audioConfig.audioEncode.enable ) {
        dev_sdk_start_audio_play( gAudioType );
        dev_sdk_start_audio( 0, 1, AudioGetFrameCb, NULL );
        DBG_LOG("channels = %d\n", audioConfig.audioCapture.channels );
        DBG_LOG("bitspersample = %d\n", audioConfig.audioCapture.bitspersample );
        DBG_LOG("samplerate = %d\n", audioConfig.audioCapture.samplerate );
        DBG_LOG("volume_capture = %d\n", audioConfig.audioCapture.volume_capture );
        DBG_LOG("amplify = %d\n", audioConfig.audioCapture.amplify );
        DBG_LOG("ra_answer = %d\n", audioConfig.audioCapture.ra_answer );
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
        ret = GetUploadToken( gTestToken, sizeof(gTestToken) );
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

    UserUploadArg userUploadArg;
    memset(&userUploadArg, 0, sizeof(userUploadArg));
    userUploadArg.pToken_ = gTestToken;
    userUploadArg.nTokenLen_ = strlen(gTestToken);
    userUploadArg.pDeviceId_ = gAjMediaStreamConfig.rtmpConfig.server;
    userUploadArg.nDeviceIdLen_ = strlen(gAjMediaStreamConfig.rtmpConfig.server);
    userUploadArg.nUploaderBufferSize = 512;

    ret = CreateAndStartAVUploader(&pTsMuxUploader, &avArg, &userUploadArg);
    if (ret != 0) {
        DBG_LOG("CreateAndStartAVUploader error, ret = %d\n", ret );
        return ret;
    }

    DBG_LOG("[ %s ] kodo init ok\n", gAjMediaStreamConfig.rtmpConfig.server );
    gKodoInitOk = 1;

}

static void * upadateToken() {
        int ret = 0;

        while( 1 ) {
            sleep( gIpcConfig.tokenUploadInterval );// 59 minutes
            memset(gtestToken, 0, sizeof(gtestToken));
            ret = GetUploadToken(gTestToken, sizeof(gTestToken));
            if ( ret != 0 ) {
                DBG_ERROR("GetUploadToken error, ret = %d\n", ret );
                return NULL;
            }
            DBG_LOG("token:%s\n", gTestToken);
            ret = UpdateToken(pTsMuxUploader, gTestToken, strlen(gTestToken));
            if (ret != 0) {
                DBG_ERROR("UpdateToken error, ret = %d\n", ret );
                return NULL;
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
}

int WaitForNetworkOk()
{
    int i = 0, ret = 0;

    for ( i=0; i<gIpcConfig.tokenRetryCount; i++ ) {
        ret = GetUploadToken( gTestToken, sizeof(gTestToken) );
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
    DBG_LOG("[ %s ] alarm.code = %d\n", gAjMediaStreamConfig.rtmpConfig.server, alarm.code );

    if ( alarm.code == ALARM_CODE_MOTION_DETECT ) {
        gMovingDetect = 1;
    } else if ( alarm.code == ALARM_CODE_MOTION_DETECT_DISAPPEAR ) {
        gMovingDetect = 0;
    }

    return 0;
}

int main()
{
    int ret = 0;

    InitConfig();
    WaitForNetworkOk();
    LoggerInit( gIpcConfig.logPrintTime, gIpcConfig.logOutput, gIpcConfig.logFile, gIpcConfig.logVerbose );
    SetLogCallback( SdkLogCallback );

    DBG_LOG("compile tile : %s %s \n", __DATE__, __TIME__ );

    StartTokenUpdateTask();
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
    DestroyAVUploader(&pTsMuxUploader);
    UninitUploader();

    return 0;
}

