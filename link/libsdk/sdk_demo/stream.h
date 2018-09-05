// Last Update:2018-07-05 11:42:28
/**
 * @file stream.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-29
 */

#ifndef STREAM_H
#define STREAM_H

extern int VideoGetFrameCb( char *_pFrame, 
                   int _nLen, int _nIskey, double _dTimeStamp, 
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex, 
                   void *_pContext);

extern int AudioGetFrameCb( char *_pFrame, int _nLen, double _dTimeStamp,
                     unsigned long _nFrameIndex, void *_pContext );

#endif  /*STREAM_H*/
