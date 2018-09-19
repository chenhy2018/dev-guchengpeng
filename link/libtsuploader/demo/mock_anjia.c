// Last Update:2018-09-19 17:29:18
/**
 * @file mock_anjia.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-09-19
 */

#ifdef MOCK_ANJIA

#include <pthread.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include "devsdk.h"

typedef struct {
    VIDEO_CALLBACK videoCb;
    AUDIO_CALLBACK audioCb;
} IpcChannelInfo;

typedef struct {
    IpcChannelInfo channels[32];
    int index;
} IpcInfo;

static IpcInfo gIpcInfo, *pIpc = &gIpcInfo;

int dev_sdk_init(DevSdkServerType type)
{
    (void)type;

    memset( pIpc, 0, sizeof(IpcInfo) );

    return 0;
}

int GetMediaStreamConfig( MediaStreamConfig  *config)
{
    strcpy( config->rtmpConfig.server, "ipc99" );

    return 0;
}

int dev_sdk_get_AudioConfig(AudioConfig *pAudioCfg)
{
    pAudioCfg->audioEncode.enable = 0;

    return 0;
}

int dev_sdk_start_audio_play(DevSdkAudioType audiotype)
{
    (void) audiotype;

    return 0;
}

int dev_sdk_stop_video(int camera, int stream)
{
    (void)camera;
    (void)stream;

    return 0;
}

int dev_sdk_stop_audio(int camera, int stream)
{
    (void)camera;
    (void)stream;

    return 0;
}

int dev_sdk_stop_audio_play(void)
{
    return 0;
}

int dev_sdk_release(void)
{
    return 0;
}

void *VideoCaptureTask( void *param )
{
    VIDEO_CALLBACK videoCallBack = (VIDEO_CALLBACK)param;
    char buffer[1024];
    int iskey = 0;
    double timestamp = 0;
    unsigned long frame_index = 0;

    memset( buffer, 1, sizeof(buffer) );

    for (;;) {
        videoCallBack( 0, buffer, sizeof(buffer), iskey, timestamp, frame_index, 0, NULL );
        timestamp += 40.0;
        frame_index++;
        if (  frame_index%20 == 0 ) {
            iskey = 1;
        } else {
            iskey = 0;
        }
        usleep( 40 );
    }

    return NULL;
}


int dev_sdk_start_video(int camera, int stream, VIDEO_CALLBACK vcb, void *pcontext)
{
    pthread_t videoTask = 0;

    (void)camera;
    (void)pcontext;

    pthread_create( &videoTask, NULL, VideoCaptureTask, (void *)vcb );

    return 0;
}

void *AudioCaptureTask( void *param )
{
    AUDIO_CALLBACK audioCallBack = (AUDIO_CALLBACK)param;
    char buffer[1024];
    double timestamp = 0;
    unsigned long frame_index = 0;

    memset( buffer, 1, sizeof(buffer) );

    for (;;) {
        audioCallBack( buffer, sizeof(buffer), timestamp, frame_index, NULL );
        timestamp += 40.0;
        frame_index++;
        usleep( 40 );
    }

    return NULL;
}

int dev_sdk_start_audio(int camera, int stream, AUDIO_CALLBACK acb, void *pcontext)
{
    pthread_t audioTask = 0;

    pthread_create( &audioTask, NULL, AudioCaptureTask, (void *)acb );

    return 0;
}

int dev_sdk_register_callback(ALARM_CALLBACK alarmcb, CONTROL_RESPONSE crespcb, DEBUG_CALLBACK debugcb, void *pcontext)
{
    return 0;
}

#endif

