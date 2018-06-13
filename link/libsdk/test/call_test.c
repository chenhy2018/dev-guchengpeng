// Last Update:2018-06-11 18:53:12
/**
 * @file call_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-11
 */

#include "unit_test.h"
#include "dbg.h"
#include "sdk_interface.h"

typedef struct {
    char *id;
    char *host;
    unsigned char init;
    int timeOut;
} CallData;

typedef struct {
    TestCase father;
    CallData data;
} MakeCallTestCase;


void *MakeCallEventLoopThread( void *arg );
int MakeCallTestSuitCallback( TestSuit *this );
int MakeCallTestSuitGetTestCase( TestSuit *this, TestCase **testCase );
int MakeCallTestSuitInit( TestSuit *this, TestSuitManager *_pManager );

MakeCallTestCase gMakeCallTestCases[] =
{
    {
        { "valid_account1", CALL_STATUS_ESTABLISHED },
        { "1015", "123.59.204.198", 0, 10 }
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
    Media media;
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

    if ( pData->init ) {
        sts = InitSDK( &media, 1 );
        if ( RET_OK != sts ) {
            UT_ERROR("sdk init error\n");
            return TEST_FAIL;
        }
    }

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
    this->total = ARRSZ(gMakeCallTestCases);
    this->index = 0;
    this->pManager = _pManager;

    return 0;
}


