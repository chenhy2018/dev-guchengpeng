// Last Update:2018-07-04 10:11:43
/**
 * @file stream.c
 * @brief push rtp strem  
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-28
 */
#include "devsdk.h"
#include "main.h"
#include "sdk_interface.h"
#include "common.h"
#include "dbg.h"

int VideoGetFrameCb( char *_pFrame, 
                   int _nLen, int _nIskey, double _dTimeStamp, 
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex, 
                   void *_pContext)
{
    ErrorID ret = 0;
    AccountID accountId = GetAccountId();
    int callId = GetCallId();
    int64_t stamp = (int64_t)_dTimeStamp;

    if ( StreamStatus() ) {
        DbgGetVideoFrame( _pFrame, _nLen );
        ret = SendPacket( accountId, callId, STREAM_VIDEO, _pFrame, _nLen, (int64_t)_dTimeStamp );
        CHECK_SDK_RETURN( SendPacket, ret );
    }

    return 0;
}

int AudioGetFrameCb( char *_pFrame, int _nLen, double _dTimeStamp,
                     unsigned long _nFrameIndex, void *_pContext )
{
    ErrorID ret = 0;
    AccountID accountId = GetAccountId();
    int callId = GetCallId();

    if ( StreamStatus() ) {
        ret = SendPacket( accountId, callId, STREAM_AUDIO, _pFrame, _nLen, (int64_t)_dTimeStamp );
        CHECK_SDK_RETURN( SendPacket, ret );
    }

    return 0;
}

int InitIPC()
{
    int context = 0;
    int s32Ret = 0;
    AudioConfig audioConfig;

    s32Ret = dev_sdk_init( DEV_SDK_PROCESS_APP );
    if ( s32Ret < 0 ) {
        DBG_ERROR("dev_sdk_init error, s32Ret = %d\n", s32Ret );
        return -1;
    }
    dev_sdk_start_video( 0, 0, VideoGetFrameCb, &context );
    dev_sdk_get_AudioConfig( &audioConfig );
    if ( audioConfig.audioEncode.enable ) {
        dev_sdk_start_audio_play( AUDIO_TYPE_G711 );
        dev_sdk_start_audio( 0, 1, AudioGetFrameCb, NULL );
    } else {
        DBG_ERROR("not enabled\n");
    }

    return 0;
}

int DeInitIPC()
{
    dev_sdk_stop_video(0, 1);
    dev_sdk_stop_audio(0, 1);
    dev_sdk_stop_audio_play();
    dev_sdk_release();
}
