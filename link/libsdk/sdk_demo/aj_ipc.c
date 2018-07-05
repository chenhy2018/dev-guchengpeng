// Last Update:2018-07-05 11:44:49
/**
 * @file aj_ipc.c
 * @brief anjia ip camera 
 * @author liyq
 * @version 0.1.00
 * @date 2018-07-05
 */

#include "devsdk.h"
#include "dbg.h"
#include "dev_core.h"

static int InitIPC( VideoFrameCb videoCb, AudioFrameCb audioCb )
{
    int context = 0;
    int s32Ret = 0;
    AudioConfig audioConfig;

    s32Ret = dev_sdk_init( DEV_SDK_PROCESS_APP );
    if ( s32Ret < 0 ) {
        DBG_ERROR("dev_sdk_init error, s32Ret = %d\n", s32Ret );
        return -1;
    }
    dev_sdk_start_video( 0, 0, videoCb, &context );
    dev_sdk_get_AudioConfig( &audioConfig );
    if ( audioConfig.audioEncode.enable ) {
        dev_sdk_start_audio_play( AUDIO_TYPE_G711 );
        dev_sdk_start_audio( 0, 1, audioCb, NULL );
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

CaptureDevice gAJCaptureDev =
{
    InitIPC,
    DeInitIPC,
};

