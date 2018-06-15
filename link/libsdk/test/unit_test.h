// Last Update:2018-06-15 15:59:30
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

#define NONE                 "\e[0m"
#define BLACK                "\e[0;30m"
#define L_BLACK              "\e[1;30m"
#define RED                  "\e[0;31m"
#define L_RED                "\e[1;31m"
#define GREEN                "\e[0;32m"
#define L_GREEN              "\e[1;32m"
#define BROWN                "\e[0;33m"
#define YELLOW               "\e[1;33m"
#define BLUE                 "\e[0;34m"
#define L_BLUE               "\e[1;34m"
#define PURPLE               "\e[0;35m"
#define L_PURPLE             "\e[1;35m"
#define CYAN                 "\e[0;36m"
#define L_CYAN               "\e[1;36m"
#define GRAY                 "\e[0;37m"
#define WHITE                "\e[1;37m"

#define BOLD                 "\e[1m"
#define UNDERLINE            "\e[4m"
#define BLINK                "\e[5m"
#define REVERSE              "\e[7m"
#define HIDE                 "\e[8m"
#define CLEAR                "\e[2J"
#define CLRLINE              "\r\e[K" //or "\e[1K\r"

#define TEST_SUIT_MAX 256
#define LOG(args...) printf(args)
#define COND_MAX (256)
#define EVENT_WAIT_MAX (256)
#define THREAD_MAX (16)
#define ERROR_TIMEOUT (-2)
#define ERROR_INVAL (-3)
#define STS_OK (0)
#undef DBG_LOG
#undef DBG_BASIC
#undef DBG_ERROR
#define DBG_BASIC() printf("[ %s %s +%d ] ", __FILE__, __FUNCTION__, __LINE__)
#define DBG_LOG(args...) DBG_BASIC();printf(" ");printf(args)
#define DBG_ERROR DBG_LOG
#define UT_LOG(args...) printf("[ UNIT TEST ] ");DBG_LOG(args)
#define UT_ERROR(args...) printf("[ UNIT TEST] ");DBG_ERROR(args)
#define UT_VAL(v) printf("[ UNIT TEST ] ");DBG_LOG(#v" = %d\n", v)
#define UT_STR(s) printf("[ UNIT TEST ]");DBG_LOG(#s" = %s\n", s)
#define UT_LINE() printf("[ UNIT TEST ]");DBG_LOG("======================\n")
#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))
#define TEST_CASE_RESULT_MAX 512
#define UT_NOTICE(args...) printf( RED"[ %s %s +%d ] "NONE, __FILE__, __FUNCTION__, __LINE__);printf(args)

typedef struct _TestSuit TestSuit;
typedef void *(*ThreadFn)(void *);
typedef struct {
    char *caseName;
    int expact;
    ThreadFn threadEntry;
    int (*TestCaseCb)( TestSuit *this );
    pthread_t tid;
    void *data;
    unsigned char running;
} TestCase;

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
