// Last Update:2018-05-29 09:48:17
/**
 * @file mqtt.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-05-29
 */

#ifndef MQTT_H
#define MQTT_H

static const int MOSQUITTO_AUTHENTICATION_NULL = 0x0;
static const int MOSQUITTO_AUTHENTICATION_USER = 0x1;
static const int MOSQUITTO_AUTHENTICATION_ONEWAY_SSL = 0x2;
static const int MOSQUITTO_AUTHENTICATION_TWOWAY_SSL = 0x4;

struct MosquittoUserInfo
{
    int nAuthenicatinMode;
    char username[MAX_MOSQUITTO_USR_SIZE];
    char password[MAX_MOSQUITTO_PWD_SIZE];
    char hostname[MAX_MOSQUITTO_HOST_SIZE];
    int nPort;
    char cafile[MAX_MOSQUITTO_FILE_SIZE];
    char certfile[MAX_MOSQUITTO_FILE_SIZE];
    char keyfile[MAX_MOSQUITTO_FILE_SIZE];
    //char bindaddress[MAX_MOSQUITTO_HOST_SIZE]; //not used in current time.
};

struct MosquittoCallback
{
    void (*onMessage)(IN const void* instance, IN const char* message, IN size_t length);
    void (*onEvent)(IN const void* instance, IN int code, const char* reason);
};

struct MosquittoOptions
{
    char id[MAX_MOSQUITTO_ID_SIZE];
    bool bCleanSession;
    struct MosquittoUserInfo primaryUserInfo;
    struct MosquittoUserInfo secondaryUserInfo;
    int nKeepalive;
    struct MosquittoCallback callbacks; // A user pointer that will be passed as an argument to any callbacks that are specified.
    int nQos;
    bool bRetain;
};

/* step 1 : Init mosquitto lib */
extern int MosquittoLibInit();
extern int MosquittoLibCleanup();
/* step 2 : create mosquitto instance */
extern void* MosquittoCreateInstance(IN const struct MosquittoOptions* pOption);
extern void MosquittoDestroy(IN const void* pInstance);
/* step 3 : mosquitto pub/sub */
extern int MosquittoPublish(IN const void* _pInstance, OUT int* _pMid, IN char* _pTopic, IN int _nPayloadlen, IN const void* _pPayload);
extern int MosquittoSubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* _pTopic);
extern int MosquittoUnsubscribe(IN const void* _pInstance, OUT Int* _pMid, IN char* pSub);

#endif  /*MQTT_H*/
