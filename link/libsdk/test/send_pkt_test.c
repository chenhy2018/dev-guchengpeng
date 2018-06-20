// Last Update:2018-06-14 16:57:37
/**
 * @file send_pkt_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-14
 */

#include <string.h>
#include "dbg.h"
#include "unit_test.h"
#include "sdk_interface.h"

typedef struct {
    char *pData;
    unsigned char video;
    unsigned char audio;
    unsigned char data;
} SendPacketData;

typedef struct {
    TestCase father;
    SendPacketData data;
} SendPacketTestCase;


int SendPacketTestSuitInit( TestSuit *this, TestSuitManager *_pManager );
int SendPacketTestSuitCallback( TestSuit *this );
int SendPacketTestSuitGetTestCase( TestSuit *this, TestCase **testCase );

SendPacketTestCase gSendPacketTestCases[] =
{
    {
        { "normal video packet", RET_OK },
        { "1234567890abc", 1 }
    },
    {
        { "normal audio packet", RET_OK },
        { "1234567890abc", 0, 1 }
    },
    {
        { "normal data packet", RET_OK },
        { "1234567890abc", 0, 1 }
    },
};

TestSuit gSendPacketTestSuit = 
{
    "SendPacket()",
    SendPacketTestSuitCallback,
    SendPacketTestSuitInit,
    SendPacketTestSuitGetTestCase,
    (void*)&gSendPacketTestCases,
    0,
};

int SendPacketTestSuitInit( TestSuit *this, TestSuitManager *_pManager )
{
    pthread_t tid = 0;
    ErrorID sts = 0;
    Media media[2];

    this->total = ARRSZ(gSendPacketTestCases);
    this->index = 0;
    this->pManager = _pManager;

    media[0].streamType = STREAM_VIDEO;
    media[0].codecType = CODEC_H264;
    media[0].sampleRate = 90000;
    media[0].channels = 0;
    media[1].streamType = STREAM_AUDIO;
    media[1].codecType = CODEC_G711A;
    media[1].sampleRate = 8000;
    media[1].channels = 1;
    sts = InitSDK( media, 2 );
    if ( RET_OK != sts ) {
        UT_ERROR("sdk init error\n");
        return -1;
    }

    return 0;
}

int SendPacketTestSuitCallback( TestSuit *this )
{
    SendPacketTestCase *pTestCases = NULL;
    SendPacketData *pData = NULL;
    SendPacketTestCase *pTestCase = NULL;
    int i = 0;
    int ret = 0, callId = 0;
    static ErrorID sts = 0;
    AccountID id = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    DataEvent *pDataEvent = NULL;
    unsigned char established = 0;
    unsigned char retry = 0;

    UT_LINE();
    if ( !this ) {
        return TEST_FAIL;
    }

    pTestCases = (SendPacketTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;

    id = Register( "1003", "1003", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        DBG_ERROR("sts = %d\n", sts );
        return TEST_FAIL;
    }

    UT_VAL( id );
    for (;;) {
        sts = PollEvent( id, &type, &pEvent, 5 );
        if ( sts >= RET_MEM_ERROR ) {
            UT_ERROR("PollEvent error, sts = %d\n", sts );
            return TEST_FAIL;
        }

        if ( RET_RETRY == sts ) {
            if ( established ) {
                retry++;
                if ( retry == 5 ) {
                    UT_ERROR("waiting data time out\n");
                    return TEST_FAIL;
                }
            } else {
                retry ++;
                if ( retry == 5 ) {
                    UT_ERROR("waiting call established time out\n");
                    return TEST_FAIL;
                }
            }
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
                if ( pCallEvent->pFromAccount )
                    UT_STR( pCallEvent->pFromAccount );
                if ( CALL_STATUS_INCOMING == pCallEvent->status ) {
                    UT_VAL( pCallEvent->callID );
                    ret = AnswerCall( id, pCallEvent->callID );
                    if ( ret != RET_OK ) {
                        return TEST_FAIL;
                    }
                } else if ( CALL_STATUS_ESTABLISHED == pCallEvent->status ) {
                    established = 1;
                    retry = 0;
                    UT_LOG("CALL_STATUS_ESTABLISHED\n");
                } else {
                    UT_ERROR("wrong event, pCallEvent->status = %d\n", pCallEvent->status );
                    return TEST_FAIL;
                }
            }
            break;
        case EVENT_DATA:
            UT_LOG("get event EVENT_DATA\n");
            if ( pEvent ) {
                pDataEvent = &pEvent->body.dataEvent;
                UT_VAL( pDataEvent->stream );
                UT_VAL( pDataEvent->size );
                switch ( pDataEvent->stream ) {
                case STREAM_VIDEO:
                    if ( pData->video ) {
                        if ( memcmp( pData->pData, pDataEvent->data, pDataEvent->size) == 0 ) {
                            return TEST_PASS;
                        } else {
                            UT_STR( (char *) pDataEvent->data );
                            return TEST_FAIL;
                        }
                    } else {
                        UT_ERROR("fail\n");
                        return TEST_FAIL;
                    }
                    break;
                case STREAM_AUDIO:
                    if ( pData->audio ) {
                        if ( memcmp( pData->pData, pDataEvent->data, pDataEvent->size) == 0 ) {
                            return TEST_PASS;
                        } else {
                            UT_STR( (char *) pDataEvent->data );
                            return TEST_FAIL;
                        }
                    } else {
                        UT_ERROR("fail\n");
                        return TEST_FAIL;
                    }
                    break;
                case STREAM_DATA:
                    if ( pData->data ) {
                        if ( memcmp( pData->pData, pDataEvent->data, pDataEvent->size) == 0 ) {
                            return TEST_PASS;
                        } else {
                            UT_STR( (char *) pDataEvent->data );
                            return TEST_FAIL;
                        }
                    } else {
                        UT_ERROR("fail\n");
                        return TEST_FAIL;
                    }
                    break;
                default:
                    UT_ERROR("unknow stream type, stream = %d\n", pDataEvent->stream );
                    break;
                }
            }
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

    return TEST_FAIL;
}

int SendPacketTestSuitGetTestCase( TestSuit *this, TestCase **testCase )
{
    SendPacketTestCase *pTestCases = NULL;

    if ( !testCase || !this ) {
        return -1;
    }

    pTestCases = (SendPacketTestCase *)this->testCases;
    *testCase = (TestCase *)&pTestCases[this->index];

    return 0;
}

static void cleanup( void *arg )
{
    pthread_t thread = (pthread_t )arg;

    printf("+++++++++ thread %d exit\n", (int)thread );
}



