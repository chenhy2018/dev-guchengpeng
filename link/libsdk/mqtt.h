#ifndef __MQTT__
#define __MQTT__

#include <stdbool.h>
#include <stddef.h>
#include "sdk_interface.h"

enum MQTT_ERR_STATUS {
        MQTT_SUCCESS = 3000,
        MQTT_CONNECT_SUCCESS = 3001,
        MQTT_DISCONNECT_SUCCESS = 3002,
        MQTT_ERR_NOMEM,
        MQTT_ERR_PROTOCOL,
        MQTT_ERR_INVAL,
        MQTT_ERR_NO_CONN,
        MQTT_ERR_CONN_REFUSED,
        MQTT_ERR_NOT_FOUND,
        MQTT_ERR_CONN_LOST,
        MQTT_ERR_TLS,
        MQTT_ERR_PAYLOAD_SIZE,
        MQTT_ERR_NOT_SUPPORTED,
        MQTT_ERR_AUTH,
        MQTT_ERR_ACL_DENIED,
        MQTT_ERR_UNKNOWN,
        MQTT_ERR_ERRNO,
        MQTT_ERR_EAI,
        MQTT_ERR_PROXY,
        MQTT_ERR_CONN_PENDING,
        MQTT_ERR_OTHERS
};

static const int MQTT_AUTHENTICATION_NULL = 0x0;
static const int MQTT_AUTHENTICATION_USER = 0x1;
static const int MQTT_AUTHENTICATION_ONEWAY_SSL = 0x2;
static const int MQTT_AUTHENTICATION_TWOWAY_SSL = 0x4;

typedef struct MqttOptions MqttOptions;

struct MqttCallback
{
        void (*OnMessage)(IN const void* _pInstance, IN const char* _pTopic, IN const char* _pMessage, IN size_t nLength);
        void (*OnEvent)(IN const void* _pInstance, IN int nCode, const char* _pReason);
};

struct MqttUserInfo
{
        int nAuthenicatinMode;
        const char* pUsername;
        const char* pPassword;
        const char* pHostname;
        int nPort;
        const char* pCafile;
        const char* pCertfile;
        const char* pKeyfile;
        //char* pBindaddress; //not used in current time.
};

struct MqttOptions
{
        const char* pId;
        bool bCleanSession;
        struct MqttUserInfo primaryUserInfo;
        struct MqttUserInfo secondaryUserInfo;
        int nKeepalive;
        struct MqttCallback callbacks; // A user pointer that will be passed as an argument to any callbacks that are specified.
        int nQos;
        bool bRetain;
};

/* step 1 : Init mosquitto lib */
extern int MqttLibInit();

extern int MqttLibCleanup();

/* step 2 : create mosquitto instance */
extern void* MqttCreateInstance(IN const struct MqttOptions* _pOption);

extern void MqttDestroy(IN const void* _pInstance);

/* step 3 : mosquitto pub/sub */

extern int MqttPublish(IN const void* _pInstance, IN char* _pTopic, IN int _nPayloadlen, IN const void* _pPayload);

extern int MqttSubscribe(IN const void* _pInstance, IN char* _pTopic);

extern int MqttUnsubscribe(IN const void* _pInstance, IN char* _pTopic);

#endif

