// Last Update:2018-09-19 15:20:54
/**
 * @file mymalloc.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-09-18
 */

#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>

static int up = 0, down = 0;
static pthread_mutex_t gMutex, gFreeMutex;

void MyMallocInit()
{
//    pthread_mutex_init( &gMutex, NULL );
//    pthread_mutex_init( &gFreeMutex, NULL );
}

#if 0
void *mymalloc( size_t size, char *function, int line )
{
    pthread_mutex_lock( &gMutex );
//    printf("+++ malloc, size = %d, %s() ---> %d, up = %d\n", size,  function, line, up++ );
    pthread_mutex_unlock( &gMutex );

    return malloc( size );
}

void myfree( void *ptr, char *function, int line )
{
    pthread_mutex_lock( &gFreeMutex  );
//    printf("+++ free, %s() ---> %d, down = %d\n", function, line, down++  );
    pthread_mutex_unlock( &gFreeMutex  );

    free( ptr );
}
#endif

void *mymalloc( size_t size, char *function, int line )
{
    fprintf(stderr, "+++ malloc, %s() ---> %d, up = %d\n", function, line,__sync_fetch_and_add(&up,1));

    return malloc( size );
}

void myfree( void *ptr, char *function, int line )
{
    fprintf(stderr, "+++ free, %s() ---> %d, down = %d\n", function, line, __sync_fetch_and_sub(&down,1));

    free( ptr );
}

