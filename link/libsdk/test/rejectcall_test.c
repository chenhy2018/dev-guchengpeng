// Last Update:2018-06-14 15:39:10
/**
 * @file rejectcall_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-14
 */
#include "dbg.h"
#include "unit_test.h"
#ifdef WITH_P2P
#include "sdk_interface_p2p.h"
#else
#include "sdk_interface.h"
#endif

typedef struct {
    char *data;
    unsigned char invalidCallId;
    unsigned char invalidAccountId;
} RejectCallData;

typedef struct {
    TestCase father;
    RejectCallData data;
} RejectCallTestCase;


int RejectCallTestSuitInit( TestSuit *this, TestSuitManager *_pManager );
int RejectCallTestSuitCallback( TestSuit *this );
int RejectCallTestSuitGetTestCase( TestSuit *this, TestCase **testCase );

RejectCallTestCase gRejectCallTestCases[] =
{
    {
        { "normal reject", CALL_STATUS_REJECT },
        { "1234567890abc" }
    },
    {
        { "invalid call id", RET_CALL_NOT_EXIST },
        { "1234567890abc", 1 }
    },
    {
        { "invalid account id", RET_ACCOUNT_NOT_EXIST },
        { "1234567890abc", 0, 1 }
    },
};

TestSuit gRejectCallTestSuit = 
{
    "RejectCall()",
    RejectCallTestSuitCallback,
    RejectCallTestSuitInit,
    RejectCallTestSuitGetTestCase,
    (void*)&gRejectCallTestCases,
    0,
};

int RejectCallTestSuitInit( TestSuit *this, TestSuitManager *_pManager )
{
    pthread_t tid = 0;
    ErrorID sts = 0;
    Media media[2];

    this->total = ARRSZ(gRejectCallTestCases);
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

int RejectCallTestSuitCallback( TestSuit *this )
{
    RejectCallTestCase *pTestCases = NULL;
    RejectCallData *pData = NULL;
    RejectCallTestCase *pTestCase = NULL;
    int i = 0;
    int ret = 0, callId = 0;
    static ErrorID sts = 0;
    AccountID id = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    int retry = 0;
    int incoming = 0;

    UT_LINE();
    if ( !this ) {
        return TEST_FAIL;
    }

    pTestCases = (RejectCallTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;
#ifdef WITH_P2P
    id = Register( "1003", "1003", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
#else
    id = Register( "1003", "1003", "123.59.204.198", "123.59.204.198" );
#endif
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
        if ( incoming && !pData->invalidCallId && !pData->invalidAccountId ) {
            retry ++;
            if ( retry == 3 ) {
                UT_ERROR("waiting for event error\n");
                return TEST_FAIL;
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
                char *pFromAccount = (char*)pCallEvent->context;
                if ( pFromAccount )
                    UT_STR( pFromAccount );
                if ( CALL_STATUS_INCOMING == pCallEvent->status ) {
                    incoming = 1;
                    UT_VAL( pCallEvent->callID );
                    if ( pData->invalidCallId ) {
                        ret = RejectCall( id, 0 );
                        return ( ret == pTestCase->father.expact ? TEST_PASS : TEST_FAIL );
                    } else if ( pData->invalidAccountId ) {
                        ret = RejectCall( 0, pCallEvent->callID );
                        return ( ret == pTestCase->father.expact ? TEST_PASS : TEST_FAIL );
                    } else {
                        ret = RejectCall( id, pCallEvent->callID );
                        if ( ret != RET_OK ) {
                            return TEST_FAIL;
                        }
                    }
                } else if ( pTestCase->father.expact == pCallEvent->status ) {
                    if ( incoming ) {
                        return TEST_PASS;
                    } else {
                        DBG_ERROR("incomint error\n");
                        return TEST_FAIL;
                    }
                } else {
                    UT_ERROR("wrong event, pCallEvent->status = %d\n", pCallEvent->status );
                    return TEST_FAIL;
                }
            }
            break;
#ifdef WITH_P2P
        case EVENT_DATA:
            UT_LOG("get event EVENT_DATA\n");
            break;
#endif
        case EVENT_MESSAGE:
            UT_LOG("get event EVENT_MESSAGE\n");
            break;
        default:
            UT_LOG("unknow event, type = %d\n", type );
            break;
        }
    }

    return TEST_FAIL;
}

int RejectCallTestSuitGetTestCase( TestSuit *this, TestCase **testCase )
{
    RejectCallTestCase *pTestCases = NULL;

    if ( !testCase || !this ) {
        return -1;
    }

    pTestCases = (RejectCallTestCase *)this->testCases;
    *testCase = (TestCase *)&pTestCases[this->index];

    return 0;
}

static void cleanup( void *arg )
{
    pthread_t thread = (pthread_t )arg;

    printf("+++++++++ thread %d exit\n", (int)thread );
}

