#include <stdio.h>
#include <libavformat/avformat.h>
#include <assert.h>
#include <sys/time.h>
#include <unistd.h>
#include <pthread.h>
#include <stdlib.h>
#include "tsuploaderapi.h"
#include "devsdk.h"

#define BASIC() printf("[ %s %s() +%d ] ", __FILE__, __FUNCTION__, __LINE__ )
#define LINE() printf("%s %s ------ %d \n", __FILE__, __FUNCTION__, __LINE__)
#define DBG_LOG( args... ) BASIC();printf(args)
#define DBG_ERROR( args... ) BASIC();printf("[ ERROR ] ");printf(args)
DevSdkAudioType audio_type =  AUDIO_TYPE_AAC;
static int kodo_init_ok = 0;

MediaStreamConfig gAjMediaStreamConfig = {0};
char gtestToken[1024] = {0};
int VideoGetFrameCb( int streamno, char *_pFrame,
                   int _nLen, int _nIskey, double _dTimeStamp,
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex,
                   void *_pContext)
{
    static int first = 1;
    static double last = 0, interval = 0, duration = 0, lastTimeStamp = 0;
    static struct timeval start = { 0, 0 }, end = { 0, 0 };
    char now[200] = { 0 };

    if ( !kodo_init_ok ) {
        return 0;
    }

    duration = _dTimeStamp - lastTimeStamp;
    gettimeofday( &end, NULL );
    interval = GetTimeDiff( &start, &end );
    if ( interval >= 5 ) {
        char message[256] = { 0 };

        memset( now, 0, sizeof(now) );
        get_current_time( now );
        sprintf( message, "[ %s ] [ %s ] [ %s ] [ video ] [ timestamp interval ] [ %f ]\n", 
                 now,
                 gAjMediaStreamConfig.rtmpConfig.streamid,
                 gAjMediaStreamConfig.rtmpConfig.server,
                 duration );
        log_send( message );
        start = end;
    }

    if ( first  ) {
        printf("video thread id : %ld\n", pthread_self() );
        first = 0;
    }
    interval = _dTimeStamp - last;
    /*DBG_LOG("interval = %f\n", interval );*/
    if ( interval <= 0 ) {
        DBG_ERROR("video time interval : %f\n", interval );
    }
    last = _dTimeStamp;
    PushVideo( _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );

    lastTimeStamp = _dTimeStamp;
    return 0;
}

int AudioGetFrameCb( char *_pFrame, int _nLen, double _dTimeStamp,
                     unsigned long _nFrameIndex, void *_pContext )
{
    static int first = 1;
    static double localTimeStamp = 0, timeStamp, lastTimeStamp=0;
    double diff = 0, duration = 0;
    static double min=0, max=0;
    int ret = 0;
    static int total = 0, error = 0;
    static int count = 0;
    static struct timeval start = { 0, 0}, end = { 0, 0 };
    int interval = 0;

    if ( !kodo_init_ok ) {
        return 0;
    }

    if ( first == 1 ) {
        printf("++++++++ audio thread id %ld\n", pthread_self() );
        localTimeStamp = _dTimeStamp;
        first = 0;
    } else {
        localTimeStamp += 40;
    }

    duration = _dTimeStamp - lastTimeStamp;
    gettimeofday( &end, NULL );
    interval = GetTimeDiff( &start, &end );
    if ( interval >= 5 ) {
        char message[256] = { 0 };
        char now[200] = { 0 };

        get_current_time( now );
        DBG_LOG("strlen(now) = %d\n", strlen(now) );
        sprintf( message, "[ %s ] [ %s ] [ %s ] [ auido ] [ timestamp interval ] [ %f ]\n", 
                 now,
                 gAjMediaStreamConfig.rtmpConfig.streamid,
                 gAjMediaStreamConfig.rtmpConfig.server,
                 duration );
        log_send( message );
        start = end;
    }

    /*DBG_LOG("localTimeStamp = %f\n", localTimeStamp );*/
    diff = localTimeStamp - _dTimeStamp;
    if ( min > diff ) {
        min = diff;
    }

    if ( max < diff ) {
        max = diff;
    }

    /*DBG_LOG("diff = %f, min = %f, max = %f\n", diff, min, max );*/
    if ( audio_type = AUDIO_TYPE_AAC ) {
        timeStamp = _dTimeStamp;
    } else {
        timeStamp = localTimeStamp;
    }

    ret = PushAudio( _pFrame, _nLen, (int64_t)timeStamp );
    if ( ret != 0 ) {
        DBG_ERROR("ret = %d\n", ret );
        error++;
    }
    lastTimeStamp = _dTimeStamp;

    return 0;
}


static int InitIPC( )
{
    static int context = 1;
    int s32Ret = 0;
    AudioConfig audioConfig;
    int ret = 0;

    s32Ret = dev_sdk_init( DEV_SDK_PROCESS_APP );
    if ( s32Ret < 0 ) {
        DBG_ERROR("dev_sdk_init error, s32Ret = %d\n", s32Ret );
        return -1;
    }
    GetMediaStreamConfig(&gAjMediaStreamConfig);
    ret = dev_sdk_start_video( 0, 0, VideoGetFrameCb, &context );
    dev_sdk_get_AudioConfig( &audioConfig );
    DBG_LOG("audioConfig.audioEncode.enable = %d\n", audioConfig.audioEncode.enable );
    if ( audioConfig.audioEncode.enable ) {
        dev_sdk_start_audio_play( audio_type );
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
    dev_sdk_stop_video(0, 1);
    dev_sdk_stop_audio(0, 1);
    dev_sdk_stop_audio_play();
    dev_sdk_release();
}

int InitKodo()
{
    int ret = 0;
    AvArg avArg;

    if ( audio_type == AUDIO_TYPE_AAC ) {
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
    SetAk("JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ");
    SetSk("G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS");

    //计算token需要，所以需要先设置
    SetBucketName("ipcamera");

    ret = GetUploadToken(gtestToken, sizeof(gtestToken));
    if (ret != 0) {
        return ret;
    }

    DBG_LOG("gAjMediaStreamConfig.rtmpConfig.streamid = %s\n", gAjMediaStreamConfig.rtmpConfig.streamid);
    DBG_LOG("gAjMediaStreamConfig.rtmpConfig.server = %s\n", gAjMediaStreamConfig.rtmpConfig.server );
    ret = InitUploader( gAjMediaStreamConfig.rtmpConfig.streamid, gAjMediaStreamConfig.rtmpConfig.server, gtestToken, &avArg);
    if (ret != 0) {
        DBG_ERROR("InitUploader error\n");
        return ret;
    }

    kodo_init_ok = 1;

}

static void * upadateToken() {
        int ret = 0;

        while(1) {
            sleep(3540);// 59 minutes
            ret = GetUploadToken(gtestToken, sizeof(gtestToken));
            if (ret != 0) {
                DBG_ERROR("update token file<<<<<<<<<<<<<\n");
                return NULL;
            }
            DBG_LOG("token:%s\n", gtestToken);
            ret = UpdateToken(gtestToken);
            if (ret != 0) {
                DBG_ERROR("update token file<<<<<<<<<<<<<\n");
                return NULL;
            }
        }
        return NULL;
}


int main()
{
    pthread_t updateTokenThread;
    pthread_attr_t attr;
    int ret = 0;

    pthread_attr_init (&attr);
    pthread_attr_setdetachstate (&attr, PTHREAD_CREATE_DETACHED);
    ret = pthread_create(&updateTokenThread, &attr, upadateToken, NULL);
    if (ret != 0) {
        printf("create update token thread fail\n");
        return ret;
    }
    pthread_attr_destroy (&attr);

    DBG_LOG("compile tile : %s %s \n", __DATE__, __TIME__ );

    InitIPC();
    ret = InitKodo();
    if ( ret < 0 ) {
        DBG_ERROR("ret = %d\n",ret );
    } else {
        DBG_ERROR("ret is 0\n");
    }
    socket_init();

    for (;; ) {
        sleep(1);
    }

    DeInitIPC();
    UninitUploader();

    return 0;
}

