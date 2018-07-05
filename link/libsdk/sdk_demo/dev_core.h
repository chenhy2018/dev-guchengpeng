// Last Update:2018-07-05 11:35:54
/**
 * @file dev_core.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-07-05
 */

#ifndef DEV_CORE_H
#define DEV_CORE_H

typedef int (*VideoFrameCb) ( char *_pFrame, 
                   int _nLen, int _nIskey, double _dTimeStamp, 
                   unsigned long _nFrameIndex, unsigned long _nKeyFrameIndex, 
                   void *_pContext );
typedef int ( *AudioFrameCb)( char *_pFrame, int _nLen, double _dTimeStamp,
                     unsigned long _nFrameIndex, void *_pContext );

typedef struct {
    int (*init)( VideoFrameCb videoCb, AudioFrameCb audioCb );
    int (*deInit)();
} CaptureDevice;

typedef struct {
    CaptureDevice *pCaptureDevice;
    int (*init)();
    int (*deInit)();
} CoreDevice;

CoreDevice * NewCoreDevice();

#endif  /*DEV_CORE_H*/
