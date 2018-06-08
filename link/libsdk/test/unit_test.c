// Last Update:2018-06-07 19:28:50
/**
 * @file unit_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-05
 */
#include <string.h>
#include <stdio.h>
#include <pthread.h>
#include <sys/time.h>
#include <errno.h>
#include "unit_test.h"
#include "dbg.h"

TestSuitManager gTestSuitManager, *pTestSuitManager = &gTestSuitManager;

int AddTestSuit( TestSuit *pTestSuit )
{
    int i = 0;

    if ( !pTestSuit ) {
        return -1;
    }

    if ( pTestSuitManager->num >= TEST_SUIT_MAX ) {
        return -1;
    }

    pTestSuitManager->testSuits[pTestSuitManager->num++] = *pTestSuit;

    return 0;
}

int RunAllTestSuits()
{
    int i = 0, j=0;
    TestSuit *pTestSuit = NULL;
    int res = 0;
    TestCase *pTestCase = NULL;
    ThreadManager *pThreadManager = &pTestSuitManager->threadManager;

    DBG_VAL( pThreadManager->num );
    for ( i=0; i<pThreadManager->num; i++ ) {
        int ret = pthread_create( &pThreadManager->threadList[i].threadId, NULL,
                        pThreadManager->threadList[i].threadFn, (void *)pTestSuitManager );
        if ( 0 != ret ) {
            DBG_ERROR("create thread error, ret = %d\n", ret );
            return -1;
        }
        DBG_VAL( ret );
    }

    DBG_VAL( pTestSuitManager->num );
    for ( i=0; i<pTestSuitManager->num; i++ ) {
        pTestSuit = &pTestSuitManager->testSuits[i];
        LOG("run the test suit : %s\n", pTestSuit->suitName );
        if ( pTestSuit->OnInit ) {
            pTestSuit->OnInit( pTestSuit, pTestSuitManager );
        }
        if ( pTestSuit->TestCaseCb ) {
            for ( j=0; j<pTestSuit->total; j++ ) {
                pTestSuit->index = j;
                res = pTestSuit->TestCaseCb( pTestSuit );
                if ( pTestSuit->GetTestCase ) {
                    pTestSuit->GetTestCase( pTestSuit, &pTestCase );
                    LOG("----- test case [ %s ] result ( %s ) \n", pTestCase->caseName, 
                        res == TEST_PASS ? "pass" : "fail" );
                }
            }
        }
    }


    return 0;
}

int TestSuitManagerInit()
{
    int i = 0;

    memset( pTestSuitManager, 0, sizeof(*pTestSuitManager) );
    pTestSuitManager->eventManager.WaitForEvent = WaitForEvent;
    pTestSuitManager->NotifyAllEvent = NotifyAllEvent;
    pTestSuitManager->AddPrivateData = AddPrivateData;
	pthread_mutex_init( &pTestSuitManager->eventManager.mutex, NULL );


    return 0;
}


/*
 *
 *  structure :
 *      eventID0 -
 *               |
 *               \----- condition 0
 *               |
 *               \------ condition 1
 *               |
 *               \------ condition 2
 *               ...
 *      eventId1 -
 *              |
 *              \------ condition 8
 *              |
 *              \------ condition 9
 *              |
 *              \------ condition 10
 *              ...
 *               
 *
 * */
int NotifyAllEvent( int _nEventId )
{
    int i = 0;
    int j = 0;
    EventManger *pEventManager = &pTestSuitManager->eventManager;

    DBG_VAL( pEventManager->eventNum );
    for ( i=0; i<pEventManager->eventNum; i++ ) {
        EventWait *pEventWait = &pEventManager->eventWait[i];
        if ( _nEventId == pEventWait->eventId ) {
            DBG_VAL( pEventWait->condNum );
            for ( j=0; j<pEventWait->condNum; j++ ) {
                pthread_cond_signal( &pEventWait->condList[j] );
            }
        }
    }

    return 0;
}

int AddEventWait( EventManger *pEventManager, EventWait *pEventWait, int nTimeOut )
{
    struct timeval now;
    struct timespec after;

    pthread_cond_init( &pEventWait->condList[pEventWait->condNum], NULL );
    pthread_mutex_lock( &pEventManager->mutex );
    gettimeofday(&now, NULL);
    after.tv_sec = now.tv_sec + nTimeOut;
    after.tv_nsec = now.tv_usec * 1000 + 10 * 1000 * 1000;
    int nReason = pthread_cond_timedwait( &pEventWait->condList[pEventWait->condNum++],
                                          &pEventManager->mutex, &after );
    if (nReason == ETIMEDOUT) {
        return ERROR_TIMEOUT;
    }
    if ( nReason == EINVAL ) {
        return ERROR_INVAL;
    }
    pthread_mutex_unlock( &pEventManager->mutex );

    return STS_OK;
}

int WaitForEvent( int _nEventId, int nTimeOut )
{
    int i = 0;
    int j = 0;
    int found = 0;
    int ret = 0;
    EventManger *pEventManager = &pTestSuitManager->eventManager;
    EventWait *pEventWait = NULL;

    DBG_VAL( pEventManager->eventNum );
    for ( i=0; i<pEventManager->eventNum; i++ ) {
        pEventWait= &pEventManager->eventWait[i];
        if ( _nEventId == pEventWait->eventId ) {
            ret = AddEventWait( pEventManager, pEventWait, nTimeOut);
            found = 1;
        }
    }

    if ( !found ) {
        DBG_LINE();
        pEventWait = &pEventManager->eventWait[pEventManager->eventNum];
        pEventWait->eventId = _nEventId;
        pEventManager->eventNum++;
        ret = AddEventWait( pEventManager, pEventWait, nTimeOut);
    }

    return ret;
} 

int ThreadRegister( ThreadFn threadFn )
{
    ThreadManager *pThreadManager = &pTestSuitManager->threadManager;

    if ( !threadFn ) {
        return -1;
    }

    pThreadManager->threadList[pThreadManager->num++].threadFn = threadFn;

    return 0;
}

int AddPrivateData( void *data )
{
    pTestSuitManager->data = data;
}

