// Last Update:2018-06-05 17:14:38
/**
 * @file unit_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-05
 */
#include <string.h>
#include <stdio.h>
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

    for ( i=0; i<TEST_SUIT_MAX; i++ ) {
        pTestSuitManager->testSuits[pTestSuitManager->num++] = *pTestSuit;
    }

    return 0;
}

int RunAllTestSuits()
{
    int i = 0;
    TestSuit *pTestSuit = NULL;
    int res = 0;
    TestCase *pTestCase = NULL;

    for ( i=0; i<pTestSuitManager->num; i++ ) {
        pTestSuit = &pTestSuitManager->testSuits[i];
        pTestSuit->index = i;
        LOG("run the test suit : %s\n", pTestSuit->suitName );
        if ( pTestSuit->OnInit ) {
            pTestSuit->OnInit( pTestSuit );
        }
        if ( pTestSuit->TestCaseCb ) {
            res = pTestSuit->TestCaseCb( pTestSuit );
            if ( pTestSuit->GetTestCase ) {
                pTestSuit->GetTestCase( pTestSuit, &pTestCase );
                LOG("----- test case [%s] result (%s) \n", pTestCase->caseName, 
                    res == TEST_PASS ? "pass" : "fail" );
            }
        }
    }

    return 0;
}


