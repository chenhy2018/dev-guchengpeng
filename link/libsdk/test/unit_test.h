// Last Update:2018-06-12 12:22:07
/**
 * @file unit_test.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-05
 */

#ifndef UNIT_TEST_H
#define UNIT_TEST_H

#include <pthread.h>
#include <stdio.h>

#define TEST_FAIL 1
#define TEST_PASS 0
#define TEST_EQUAL( actual, expect ) ( actual == expect ? TEST_PASS : TEST_FAIL )
#define TEST_GT( actual, expect ) ( actual >= expect ? TEST_PASS : TEST_FAIL )

#define TEST_SUIT_MAX 256
#define LOG(args...) printf(args)
#define COND_MAX (256)
#define EVENT_WAIT_MAX (256)
#define THREAD_MAX (16)
#define ERROR_TIMEOUT (-2)
#define ERROR_INVAL (-3)
#define STS_OK (0)
#define UT_LOG(args...) printf("[ UNIT TEST ] ");DBG_LOG(args)
#define UT_ERROR(args...) printf("[ UNIT TEST] ");DBG_ERROR(args)
#define UT_VAL(v) printf("[ UNIT TEST ] ");DBG_LOG(#v" = %d\n", v)
#define UT_STR(s) printf("[ UNIT TEST ]");DBG_LOG(#s" = %s\n", s)
#define UT_LINE() printf("[ UNIT TEST ]");DBG_LOG("======================\n")
#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))
#define TEST_CASE_RESULT_MAX 512

typedef void *(*ThreadFn)(void *);
typedef struct {
    char *caseName;
    int expact;
    ThreadFn threadEntry;
    pthread_t tid;
} TestCase;

typedef struct _TestSuit TestSuit;
typedef struct _TestSuitManager TestSuitManager; 
struct _TestSuit {
    char *suitName;
    int (*TestCaseCb)( TestSuit *this );
    int (*OnInit)( TestSuit *this, TestSuitManager *pManager );
    int (*GetTestCase) ( TestSuit *this, TestCase **testCase );
    void *testCases;
    unsigned char enable;
    ThreadFn threadEntry;
    int total;
    int failNum;
    int index;
    TestSuitManager *pManager;
    pthread_t tid;
};

typedef struct {
    int eventId;
    int condNum;
    pthread_cond_t condList[COND_MAX];
} EventWait;

typedef struct {
    int eventNum;
	pthread_mutex_t    mutex;
    EventWait eventWait[EVENT_WAIT_MAX];
    int (*WaitForEvent)( int event, int timeout );
} EventManger;

typedef struct {
    char *pTestCaseName;
    int res;
} TestCaseResult;

typedef struct {
    TestCaseResult results[TEST_CASE_RESULT_MAX];
    char *pTestSuitName;
    int num;
}  TestSuitResult;

struct _TestSuitManager{
    void *data;
    int num;
    TestSuit testSuits[TEST_SUIT_MAX];
    TestSuitResult testSuitResults[TEST_SUIT_MAX];
    EventManger eventManager;
    int (* NotifyAllEvent)( int _nEventId );
    int (*AddPrivateData)( void *data );
    int (*startThread)( TestSuit *_pTestSuit,  ThreadFn threadFn );
    int (*CancelThread)( TestSuit *pTestSuit );
    int (*Report)();
};

extern int AddTestSuit( TestSuit *pTestSuit );
extern int RunAllTestSuits();
extern int TestSuitManagerInit();
extern int NotifyAllEvent( int event );
extern int WaitForEvent( int _nEventId, int nTimeOut );
extern int AddPrivateData( void *data );
extern int startThread( TestSuit *_pTestSuit, ThreadFn threadFn );
extern int CancelThread( TestSuit *_pTestSuit );
extern int ResultReport();

#endif  /*UNIT_TEST_H*/
