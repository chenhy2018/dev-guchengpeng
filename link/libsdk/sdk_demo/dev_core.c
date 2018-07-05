// Last Update:2018-07-05 11:41:14
/**
 * @file dev_core.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-07-05
 */

#include "dev_core.h"
#include "stream.h"

static CoreDevice gCoreDevice, *pCoreDevice = &gCoreDevice;

// struct variable declear
#undef DEV_CORE_CAPTURE_DEVICE_ENTRY
#define DEV_CORE_CAPTURE_DEVICE_ENTRY( dev )  extern CaptureDevice dev; 
#include "dev_config.h"

// struct variable use
#undef DEV_CORE_CAPTURE_DEVICE_ENTRY
#define DEV_CORE_CAPTURE_DEVICE_ENTRY( dev )   pCaptureDevice = &dev;  

CaptureDevice *GetCaptureDevice()
{
    CaptureDevice *pCaptureDevice;

    #include "dev_config.h"

    return pCaptureDevice;
}

int CoreDeviceInit()
{
    if ( pCoreDevice->pCaptureDevice )
        pCoreDevice->pCaptureDevice->init( VideoGetFrameCb, AudioGetFrameCb );

    return 0;
}

int CoreDeviceDeInit()
{
    if ( pCoreDevice->pCaptureDevice )
        pCoreDevice->pCaptureDevice->deInit();

    return 0;
}

CoreDevice * NewCoreDevice()
{
    pCoreDevice->pCaptureDevice = GetCaptureDevice();
    pCoreDevice->init = CoreDeviceInit;
    pCoreDevice->deInit = CoreDeviceDeInit;

    return pCoreDevice;
}

