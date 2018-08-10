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

static MediaStreamConfig gAjMediaStreamConfig = {0};
char gtestToken[1024] = {0};
int VideoGetFrameCb( int streamno, char *_pFrame,
                   int _nLen, int _nIskey, double _dTimeStamp,
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex,
                   void *_pContext)
{
    static int first = 1;
    static double last = 0, interval = 0;

    if ( first  ) {
        printf("video thread id : %ld\n", pthread_self() );
        first = 0;
    }
    interval = _dTimeStamp - last;
    if ( interval <= 0 ) {
        DBG_ERROR("video time interval : %f\n", interval );
    }
    last = _dTimeStamp;
    PushVideo( _pFrame, _nLen, (int64_t)_dTimeStamp, _nIskey, 0 );

    return 0;
}

int AudioGetFrameCb( char *_pFrame, int _nLen, double _dTimeStamp,
                     unsigned long _nFrameIndex, void *_pContext )
{
    static int first = 1;
    static double last = 0, interval = 0;
    int i = 0;
    static int count = 0;

    if ( first == 1 ) {
        printf("++++++++ audio thread id %ld\n", pthread_self() );
        first = 0;
    }

    DBG_LOG("_dTimeStamp = %f, _nLen = %d\n", _dTimeStamp, _nLen );
    interval = _dTimeStamp - last;
    if ( interval <= 0 ) {
        DBG_ERROR("audio time interval : %f\n", interval );
    }
    last = _dTimeStamp;
    PushAudio( _pFrame, _nLen, (int64_t)_dTimeStamp );
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
        dev_sdk_start_audio_play( AUDIO_TYPE_AAC );
        dev_sdk_start_audio( 0, 1, AudioGetFrameCb, NULL );
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

    avArg.nAudioFormat = TK_AUDIO_PCMU;
    avArg.nChannels = 1;
    avArg.nSamplerate = 8000;
    avArg.nVideoFormat = TK_VIDEO_H264;

    SetLogLevelToDebug();
    SetAk("JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ");
    SetSk("G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS");

    //计算token需要，所以需要先设置
    SetBucketName("ipcamera");

    ret = GetUploadToken(gtestToken, sizeof(gtestToken));
    if (ret != 0)
        return ret;

    ret = InitUploader("testuid5", "testdeviceid5", gtestToken, &avArg);
    if (ret != 0) {
        return ret;
    }

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

    ret = InitKodo();
    if ( ret < 0 ) {
        DBG_ERROR("ret = %d\n",ret );
    } else {
        DBG_ERROR("ret is 0\n");
    }
    InitIPC();

    for (;; ) {
        sleep(1);
    }

    DeInitIPC();
    UninitUploader();

    return 0;
}

