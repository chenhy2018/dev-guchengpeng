#include <stdio.h>
#include <libavformat/avformat.h>
#include <assert.h>
#include <sys/time.h>
#include <unistd.h>
#include <pthread.h>
#include <stdlib.h>
#include "tsuploaderapi.h"
#include "devsdk.h"

#define LINE() printf("%s %s ------ %d \n", __FILE__, __FUNCTION__, __LINE__)

#define DBG_ERROR printf

static MediaStreamConfig gAjMediaStreamConfig = {0};
char gtestToken[1024] = {0};
int VideoGetFrameCb( int streamno, char *_pFrame,
                   int _nLen, int _nIskey, double _dTimeStamp,
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex,
                   void *_pContext)
{
    static int first = 1, count = 0;
    static double last = 0, interval = 0;

    if ( _pFrame == NULL ) {
        printf("_pFrame is null\n");
    }
    if ( first  ) {
        printf("video thread id : %ld\n", pthread_self() );
        first = 0;
    }
    if (count++ < 300) {
        /*int i = 0;*/
        /*int type = _pFrame[3] & 0x7E;*/
        /*type = (type >> 1); */
        /*printf("type:%d|", type);*/
        /*for(  i = 0; i < 60; i++) {*/
            /*printf("%02x", _pFrame[i]);*/
        /*}   */
        /*printf("\n");*/
    } else {
        /*exit(1);*/
    }
    interval = _dTimeStamp - last;
    /*printf("time interval : %f\n", interval );*/
    if ( interval <= 0 ) {
        printf("video time interval : %f\n", interval );
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
#if 0
        FILE *fp = fopen("audio.aac","wb+");
        size_t size = 0;
        if ( fp ) {
            size = fwrite( _pFrame, _nLen, 1, fp );
            if ( size != 1 ) {
                DBG_ERROR("fwrite error\n");
                return 0;
            }
            fclose( fp );
            printf("++++++++++++++ write one frame aac audio sucefully\n"); 
            exit(1);
        } else {
            DBG_ERROR("open file error\n");
        }
        printf("audio thread id : %ld\n", pthread_self() );
#endif
        first = 0;
    }

    interval = _dTimeStamp - last;
    /*printf("audio time interval : %f\n", interval );*/
    if ( interval <= 0 ) {
        printf("audio time interval : %f\n", interval );
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
    char version[64] = { 0 };

    /*ret = dev_sdk_version( version, 64 );*/
    /*printf("ret = %d\n", ret );*/
    /*printf("version = %s\n", version );*/
    printf("before dev_sdk_init\n");
    s32Ret = dev_sdk_init( DEV_SDK_PROCESS_APP );
    printf("after dev_sdk_init\n");
    if ( s32Ret < 0 ) {
        DBG_ERROR("dev_sdk_init error, s32Ret = %d\n", s32Ret );
        LINE();
        return -1;
    }
	GetMediaStreamConfig(&gAjMediaStreamConfig);
    ret = dev_sdk_start_video( 0, 0, VideoGetFrameCb, &context );
    dev_sdk_get_AudioConfig( &audioConfig );
    printf("audioConfig.audioEncode.enable = %d\n", audioConfig.audioEncode.enable );
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

    //memset( avArg, 0, sizeof(avArg) );
    avArg.nAudioFormat = TK_AUDIO_AAC;
    avArg.nChannels = 1;
    avArg.nSamplerate = 16000;
    avArg.nVideoFormat = TK_VIDEO_H264;


    SetLogLevelToDebug();

    SetAk("JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ");
    SetSk("G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS");

    //计算token需要，所以需要先设置
    SetBucketName("ipcamera");

    ret = GetUploadToken(gtestToken, sizeof(gtestToken));
    if (ret != 0)
        return ret;

    //ret = InitUploader("testuid", "testdeviceid", "bucket", gtestToken, &avArg );
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
                        printf("update token file<<<<<<<<<<<<<\n");
                        return NULL;
                }
                printf("token:%s\n", gtestToken);
                ret = UpdateToken(gtestToken);
                if (ret != 0) {
                        printf("update token file<<<<<<<<<<<<<\n");
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

    printf("compile tile : %s %s \n", __DATE__, __TIME__ );

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
