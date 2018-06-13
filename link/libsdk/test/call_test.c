// Last Update:2018-06-13 16:50:38
/**
 * @file call_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-11
 */

#include "dbg.h"
#include "unit_test.h"
#include "sdk_interface.h"

typedef struct {
    char *id;
    char *host;
    unsigned char init;
    int timeOut;
    unsigned char media;
} CallData;

typedef struct {
    TestCase father;
    CallData data;
} MakeCallTestCase;


void *MakeCallEventLoopThread( void *arg );
int MakeCallTestSuitCallback( TestSuit *this );
int MakeCallTestSuitGetTestCase( TestSuit *this, TestCase **testCase );
int MakeCallTestSuitInit( TestSuit *this, TestSuitManager *_pManager );
void *CalleeThread( void *arg );

MakeCallTestCase gMakeCallTestCases[] =
{
    {
        { "valid_account1", CALL_STATUS_ESTABLISHED },
        { "1015", "123.59.204.198", 1, 10, 1 }
    },
};

TestSuit gMakeCallTestSuit =
{
    "MakeCall",
    MakeCallTestSuitCallback,
    MakeCallTestSuitInit,
    MakeCallTestSuitGetTestCase,
    (void*)&gMakeCallTestCases,
    1,
    MakeCallEventLoopThread
};

static void cleanup( void *arg )
{
    pthread_t thread = (pthread_t )arg;

    printf("+++++++++ thread %d exit\n", (int)thread );
}

void *MakeCallEventLoopThread( void *arg )
{
    TestSuit *pTestSuit = (TestSuit *)arg;
    TestSuitManager *pManager = pTestSuit->pManager;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event, *pEvent;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;

    UT_LOG("EventLoopThread enter ...\n");
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }

    if ( !pManager->data ) {
        UT_ERROR("check data error\n");
        return NULL;
    }

    id = *(AccountID *)pManager->data;
    UT_VAL( id );


    for (;;) {
        UT_LOG("call PollEvent\n");
        pthread_cleanup_push( cleanup, (void *)pTestSuit->tid );
        ret = PollEvent( id, &type, &pEvent, 0 );
        pthread_cleanup_pop( 0 );
        UT_VAL( type );
        if ( type == EVENT_CALL ) {
            UT_LINE();
            pCallEvent = &pEvent->body.callEvent;
            if ( pManager->NotifyAllEvent ) {
                UT_VAL( pCallEvent->status );
                pManager->NotifyAllEvent( pCallEvent->status );
            }
        }
    }

    return NULL;
}

int MakeCallTestSuitGetTestCase( TestSuit *this, TestCase **testCase )
{
    MakeCallTestCase *pTestCases = NULL;

    if ( !testCase || !this ) {
        return -1;
    }

    pTestCases = (MakeCallTestCase *)this->testCases;
    *testCase = (TestCase *)&pTestCases[this->index];

    return 0;
}

int MakeCallTestSuitCallback( TestSuit *this )
{
    MakeCallTestCase *pTestCases = NULL;
    CallData *pData = NULL;
    MakeCallTestCase *pTestCase = NULL;
    int i = 0;
    int ret = 0, callId = 0;
    static ErrorID sts = 0;
    EventManger *pEventManager = &this->pManager->eventManager;

    UT_LINE();
    if ( !this ) {
        return -1;
    }

    pTestCases = (MakeCallTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;

    UT_STR( pData->id );

    sts = Register( "1003", "1003", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        DBG_ERROR("sts = %d\n", sts );
        return TEST_FAIL;
    }
    UT_VAL( sts );
    this->pManager->AddPrivateData( &sts );
    if ( pTestCase->father.threadEntry )
        this->pManager->startThread( this, pTestCase->father.threadEntry );
    else
        this->pManager->startThread( this, this->threadEntry );

    if ( pEventManager->WaitForEvent ) {
        UT_VAL( pTestCase->father.expact );
        ret = pEventManager->WaitForEvent( CALL_STATUS_REGISTERED, 10 );
        if ( ret == ERROR_TIMEOUT ) {
            UT_ERROR("ERROR_TIMEOUT\n");
            return TEST_FAIL;
        } else if ( ret == ERROR_INVAL ) {
            UT_ERROR("ERROR_INVAL\n");
            return TEST_FAIL;
        } else {
            UT_VAL( ret );
            sts = MakeCall( sts,  pData->id, pData->host, &callId );
            UT_VAL( sts );
            if ( sts != RET_OK ) {
                UT_ERROR( "make call error, sts = %d\n", sts );
                return TEST_FAIL;
            }
            UT_VAL( callId );
            ret = pEventManager->WaitForEvent( pTestCase->father.expact, pData->timeOut );
            if ( ret != STS_OK ) {
                UT_ERROR(" MakeCall fail, ret = %d\n", ret );
                return TEST_FAIL;
            } else {
                return TEST_PASS;
            }
        }
    }

    return TEST_FAIL;
}


int MakeCallTestSuitInit( TestSuit *this, TestSuitManager *_pManager )
{
    pthread_t tid = 0;
    ErrorID sts = 0;
    Media media[2];

    this->total = ARRSZ(gMakeCallTestCases);
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

    pthread_create( &tid,  NULL, CalleeThread, (void *)this );

    return 0;
}

void *CalleeThread( void *arg )
{
    ErrorID sts = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    Media media;

    UT_LOG("CalleeThread() entry...\n");

    sts = Register( "1015", "1015", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        UT_ERROR("Register error, sts = %d\n", sts );
        return NULL;
    }

    for (;;) {
        sts = PollEvent( sts, &type, &pEvent, 5 );
        if ( sts >= RET_MEM_ERROR ) {
            UT_ERROR("PollEvent error, sts = %d\n", sts );
            return NULL;
        }
        UT_VAL( type );
        switch( type ) {
        case EVENT_CALL:
            UT_LOG("get event EVENT_CALL\n");
            pCallEvent = &pEvent->body.callEvent;
            char *callSts = DbgCallStatusGetStr( pCallEvent->status );
            UT_LOG("status : %s\n", callSts );
            UT_STR( pCallEvent->pFromAccount );
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

    return NULL;
}

