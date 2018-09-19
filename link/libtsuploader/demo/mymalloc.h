// Last Update:2018-09-18 20:01:36
/**
 * @file mymalloc.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-09-18
 */

#ifndef MYMALLOC_H
#define MYMALLOC_H

#define malloc( size ) mymalloc( size, __FUNCTION__, __LINE__ )
#define free( ptr ) myfree( ptr, __FUNCTION__, __LINE__ )

void *mymalloc( size_t size, char *function, int line );
void myfree( void *ptr, char *function, int ine );
void MyMallocInit();

#endif  /*MYMALLOC_H*/
