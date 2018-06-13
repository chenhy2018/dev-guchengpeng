// Last Update:2018-06-13 19:11:37
/**
 * @file callee.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-13
 */

#include "../../dbg.h"
#include "../unit_test.h"
#include "../../sdk_interface.h"

void callee()
{
    ErrorID sts = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    Media media[2];
    AccountID id = 0;

    UT_LOG("CalleeThread() entry...\n");

    media[0].streamType = STREAM_VIDEO;
    media[0].codecType = CODEC_H264;
    media[0].sampleRate = 90000;
    media[0].channels = 0;
    media[1].streamType = STREAM_AUDIO;
    media[1].codecType = CODEC_G711A;
    media[1].sampleRate = 8000;
    media[1].channels = 1;
    sts = InitSDK( media, 2 );
    if ( sts >= RET_MEM_ERROR ) {
        UT_ERROR("InitSDK error\n");
        return;
    }

    id = Register( "1002", "1002", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        UT_ERROR("Register error, sts = %d\n", sts );
        return;
    }

    UT_VAL( sts );

    for (;;) {
        sts = PollEvent( id, &type, &pEvent, 5 );
        if ( sts >= RET_MEM_ERROR ) {
            UT_ERROR("PollEvent error, sts = %d\n", sts );
            return;
        }
        UT_VAL( type );
        switch( type ) {
        case EVENT_CALL:
            UT_LOG("get event EVENT_CALL\n");
            UT_LOG("pEvent = %p\n", pEvent );
            if ( pEvent ) {
                pCallEvent = &pEvent->body.callEvent;
                UT_VAL( pCallEvent->status );
                char *callSts = DbgCallStatusGetStr( pCallEvent->status );
                UT_LOG("status : %s\n", callSts );
                UT_STR( pCallEvent->pFromAccount );
            }
            break;
        case EVENT_DATA:
            UT_LOG("get event EVENT_DATA\n");
            break;
        case EVENT_MESSAGE:
            UT_LOG("get event EVENT_MESSAGE\n");
            break;
        case EVENT_MEDIA:
            UT_LOG("get event EVENT_MEDIA\n");
            break;
        default:
            UT_LOG("unknow event, type = %d\n", type );
            break;
        }
    }
}

int main()
{
    callee();
    return 0;
}

