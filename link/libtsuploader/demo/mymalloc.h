// Last Update:2018-09-21 12:17:48
/**
 * @file mymalloc.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-09-18
 */

#ifndef MYMALLOC_H
#define MYMALLOC_H

#ifdef USE_OWN_MALLOC
#define malloc( size ) mymalloc( size, __FUNCTION__, __LINE__ )
#define free( ptr ) myfree( ptr, __FUNCTION__, __LINE__ )
#endif

void *mymalloc( size_t size, char *function, int line );
void myfree( void *ptr, char *function, int ine );

#endif  /*MYMALLOC_H*/
