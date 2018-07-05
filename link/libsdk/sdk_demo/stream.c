// Last Update:2018-07-05 10:15:23
/**
 * @file stream.c
 * @brief push rtp strem  
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-28
 */
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


