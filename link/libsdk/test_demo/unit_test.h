// Last Update:2018-06-05 22:43:40
/**
 * @file unit_test.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-05
 */

#ifndef UNIT_TEST_H
#define UNIT_TEST_H

#define TEST_FAIL 1
#define TEST_PASS 0
#define TEST_EQUAL( actual, expect ) ( actual == expect ? TEST_PASS : TEST_FAIL )
#define TEST_GT( actual, expect ) ( actual >= expect ? TEST_PASS : TEST_FAIL )

#define TEST_SUIT_MAX 256
#define LOG(args...) printf(args)

typedef struct {
    char *caseName;
    int expact;
} TestCase;


typedef struct _TestSuit TestSuit;
struct _TestSuit {
    char *suitName;
    int (*TestCaseCb)( TestSuit *this );
    int (*OnInit)( TestSuit *this );
    int (*GetTestCase) ( TestSuit *this, TestCase **testCase );
    void *testCases;
    int total;
    int failNum;
    int index;
};

typedef struct {
    TestSuit testSuits[TEST_SUIT_MAX];
    int num;
} TestSuitManager;

extern int AddTestSuit( TestSuit *pTestSuit );
extern int RunAllTestSuits();
extern int TestSuitManagerInit();

#endif  /*UNIT_TEST_H*/
