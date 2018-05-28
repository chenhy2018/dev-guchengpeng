#include "mosquitto.h"
#include <stdio.h>
#include <pthread.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>

#define STATUS_IDLE 0
#define STATUS_CONNECTING 1
#define STATUS_CONNACK_RECVD 2
#define STATUS_WAITING 3

#define MinQueueSize 50

typedef struct Node
{
        char topic[MAX_MOSQUITTO_TOPIC_SIZE];
        struct Node *pNext;
}Node;

bool insertNode(Node* pHead, char* val) {
        int i = 0;
        Node* p = pHead;
        while(p->pNext) {
                i++;
                p = p->pNext;
        }
        Node* pNew = malloc(sizeof(Node));
        strcpy(pNew->topic, val);
        p->pNext = pNew;
        pNew->pNext = NULL;
        return true;
}


bool deleteNode(Node* PHead, char * pval)
{
        int i = 0;
        Node* p = PHead;
        while (p->pNext != NULL) {
               if (strcmp(p->pNext->topic, pval) == 0) {
                       Node* temp = p->pNext;
                       p->pNext = temp->pNext;
                       free(temp);
                       return true;
               }
        }
        printf("Can't find node \n");
        return false;
}

struct MosquittoInstance
{
        struct mosquitto *mosq;
        struct MosquittoOptions options;
        int status;
        bool connected;
        bool isDestroying;
        Node pSubsribeList;
};

void onEventCallback(struct MosquittoInstance* _pInstance, int rc, const char* _pStr)
{
        if (_pInstance->options.callbacks.onEvent != NULL) {
                _pInstance->options.callbacks.onEvent(_pInstance, rc, _pStr);
        }
}

void onLogCallback(struct mosquitto* _pMosq, void* _pObj, int level, const char* _pStr)
{
        printf("%s\n", _pStr);
}

void onMessageCallback(struct mosquitto* _pMosq, void* _pObj, const struct mosquitto_message* _pMessage)
{
        int rc = MOSQ_ERR_SUCCESS;
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pObj);
        if (pInstance->options.callbacks.onMessage) {
                pInstance->options.callbacks.onMessage(_pObj, _pMessage->payload, _pMessage->payloadlen);
        }
}

void onConnectCallback(struct mosquitto* _pMosq, void* _pObj, int result)
{
        int rc = MOSQ_ERR_SUCCESS;
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pObj);
        pInstance->connected = true;
        fprintf(stderr, " on_connect_callback \n ");
        if (result) {
                fprintf(stderr, "%s\n", mosquitto_connack_string(result));
        }
        else {
                pInstance->status = STATUS_CONNACK_RECVD;
                Node* p = pInstance->pSubsribeList.pNext;
                while (p) {
                        mosquitto_subscribe(pInstance->mosq, NULL, p->topic, pInstance->options.nQos);
                        p = p->pNext;
                }
        }
        onEventCallback(pInstance, result, (result == 0) ? "on connect success" : mosquitto_connack_string(result));
}


void onDisconnectCallback(struct mosquitto* _pMosq, void* _pObj, int rc)
{

        fprintf(stderr, " on_disconnect_callback \n ");
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pObj);
        pInstance->connected = false;
        pInstance->status = STATUS_IDLE;
        onEventCallback(pInstance, rc, (rc == 0) ? "on disconnect success" : mosquitto_connack_string(rc));
}

void onSubscribeCallback(struct mosquitto* _pMosq, void* pObj, int mid, int qos_count, const int* pGranted_qos)
{       
        fprintf(stderr, "Subscribed (mid: %d): %d", mid, pGranted_qos[0]);
}

void onUnsubscribeCallback(struct mosquitto* _pMosq, void* _pObj, int mid)
{
        fprintf(stderr, "Unsubscribed (mid: %d)", mid);
}

void onPublishCallback(struct mosquitto* _pMosq, void* _pObj, int mid)
{
        fprintf(stderr, " my_publish_callback \n ");
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pObj);
        int last_mid_sent = mid;
}


static void MosquittoInstanceInit(struct MosquittoInstance* _pInstance, const struct MosquittoOptions* _pOption)
{
        /* copy options */
        memcpy(&_pInstance->options, _pOption, sizeof(struct MosquittoOptions));
        _pInstance->options.id[MAX_MOSQUITTO_ID_SIZE - 1] = 0;
        _pInstance->options.primaryUserInfo.username[MAX_MOSQUITTO_USR_SIZE - 1] = 0;
        _pInstance->options.primaryUserInfo.password[MAX_MOSQUITTO_PWD_SIZE - 1] = 0;
        _pInstance->options.primaryUserInfo.hostname[MAX_MOSQUITTO_HOST_SIZE - 1] = 0;
        _pInstance->options.secondaryUserInfo.username[MAX_MOSQUITTO_USR_SIZE - 1] = 0;
        _pInstance->options.secondaryUserInfo.password[MAX_MOSQUITTO_PWD_SIZE - 1] = 0;
        _pInstance->options.secondaryUserInfo.hostname[MAX_MOSQUITTO_HOST_SIZE - 1] = 0;

        _pInstance->mosq = NULL;
        _pInstance->connected = false;
        _pInstance->status = STATUS_IDLE;
        _pInstance->isDestroying = false;
        _pInstance->pSubsribeList.pNext = NULL;
}

bool ClientOptSet(struct MosquittoInstance* _pInstance, struct mosquitto* _pMosq, struct MosquittoUserInfo info)
{
        int rc = 0;
        if (info.nAuthenicatinMode & MOSQUITTO_AUTHENTICATION_USER) {
                printf("mosquitto_username_pw_set \n");
                rc = mosquitto_username_pw_set(_pMosq, info.username, info.password);
                if (rc)
                        return rc;
        }
        if (info.nAuthenicatinMode & MOSQUITTO_AUTHENTICATION_ONEWAY_SSL) {
                printf("mosquitto_tls_set \n");
                rc = mosquitto_tls_set(_pMosq, info.cafile, NULL, NULL, NULL, NULL);
        }
        else if (info.nAuthenicatinMode & MOSQUITTO_AUTHENTICATION_TWOWAY_SSL) {
                printf("mosquitto_tls_set 111 \n");
                rc = mosquitto_tls_set(_pMosq, info.cafile, NULL, info.certfile, info.keyfile, NULL);
        }
        return rc;
}

void * Mosquittothread(void* _pData)
{
        int rc;
        
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pData);

        pInstance->mosq = mosquitto_new(pInstance->options.id, true, pInstance);
        if (!pInstance->mosq) {
                switch(errno) {
                        case ENOMEM:
                                fprintf(stderr, "Error: Out of memory.\n");
                                break;
                        case EINVAL:
                                fprintf(stderr, "Error: Invalid id.\n");
                                break;
                }
                mosquitto_lib_cleanup();
                return NULL;
        }
        mosquitto_threaded_set(pInstance->mosq, true);
        //mosquitto_log_callback_set(pInstance->mosq, onLogCallback);
        mosquitto_connect_callback_set(pInstance->mosq, onConnectCallback);
        mosquitto_disconnect_callback_set(pInstance->mosq, onDisconnectCallback);
        mosquitto_publish_callback_set(pInstance->mosq, onPublishCallback);
        mosquitto_message_callback_set(pInstance->mosq, onMessageCallback);
        mosquitto_subscribe_callback_set(pInstance->mosq, onSubscribeCallback);
        mosquitto_unsubscribe_callback_set(pInstance->mosq, onUnsubscribeCallback);
        
        do {
                 if (!pInstance->connected && pInstance->status == STATUS_IDLE) {
                         fprintf(stderr, "connecting \n");
                         pInstance->status = STATUS_CONNECTING;
                         rc = ClientOptSet(pInstance, pInstance->mosq, pInstance->options.primaryUserInfo);
                         if (rc == 0) {
                                 rc = mosquitto_connect(pInstance->mosq, pInstance->options.primaryUserInfo.hostname, pInstance->options.primaryUserInfo.nPort, pInstance->options.nKeepalive);
                         }
                         if (rc) {
                                 onEventCallback(pInstance, rc, mosquitto_strerror(rc));
                                 fprintf(stderr, "Unable to connect (%s). try to reconnect to secondary server. \n", mosquitto_strerror(rc));
                                 rc = ClientOptSet(pInstance, pInstance->mosq, pInstance->options.secondaryUserInfo);
                                 if (rc == 0) {
                                         rc = mosquitto_connect(pInstance->mosq, pInstance->options.secondaryUserInfo.hostname, pInstance->options.secondaryUserInfo.nPort, pInstance->options.nKeepalive);
                                 }
                                 if (rc) {
                                         fprintf(stderr, "Unable to connect Secondary server  %s \n", mosquitto_strerror(rc) );
                                         pInstance->status = STATUS_IDLE;
                                         // TODO add error callback.
                                         onEventCallback(pInstance, rc, mosquitto_strerror(rc));
                                         sleep(30);
                                 }
                                 else {
                                         pInstance->status = STATUS_CONNECTING;
                                 }
                         }
                         sleep(1);
                 }
                 rc = mosquitto_loop(pInstance->mosq, -1, 1);
        } while (!pInstance->isDestroying);
        printf("quite !!! \n");
        if (pInstance->connected) {
                mosquitto_disconnect(pInstance->mosq);
        }
        mosquitto_destroy(pInstance->mosq);
        free(pInstance);
        if (rc) {
                fprintf(stderr, "Error: %s\n", mosquitto_strerror(rc));
        }
        return NULL;
}

void* MosquittoCreateInstance(IN const struct MosquittoOptions* pOption)
{
        /* allocate one mosquitto instance struct */
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)malloc(sizeof(struct MosquittoInstance));
        if (pInstance == NULL) {
                return NULL;
        }
        
        MosquittoInstanceInit(pInstance, pOption);
        pthread_t t;
        pthread_create(&t, NULL, Mosquittothread, pInstance);
        return pInstance;
}

void MosquittoDestroy(IN const void* _pInstance)
{	
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pInstance);
        pInstance->isDestroying = true;;
        pInstance->options.callbacks.onMessage = NULL;
        pInstance->options.callbacks.onEvent = NULL;
}

static int MosquittoErrorStatusChange(int nStatus)
{
        switch (nStatus) {
                case MOSQ_ERR_CONN_PENDING:
                        return MOSQUITTO_ERR_CONN_PENDING;
                case MOSQ_ERR_NOMEM:
                        return MOSQUITTO_ERR_NOMEM;
                case MOSQ_ERR_INVAL:
                        return MOSQUITTO_ERR_INVAL;
                case MOSQ_ERR_NO_CONN:
                        return MOSQUITTO_ERR_NO_CONN;
                case MOSQ_ERR_CONN_REFUSED:
                        return MOSQUITTO_ERR_CONN_REFUSED;
                case MOSQ_ERR_NOT_FOUND:
                        return MOSQUITTO_ERR_NOT_FOUND;
                case MOSQ_ERR_CONN_LOST:
                        return MOSQUITTO_ERR_CONN_LOST;
                case MOSQ_ERR_TLS:
                        return MOSQUITTO_ERR_TLS;
                case MOSQ_ERR_PAYLOAD_SIZE:
                        return MOSQUITTO_ERR_PAYLOAD_SIZE;
                case MOSQ_ERR_NOT_SUPPORTED:
                        return MOSQUITTO_ERR_NOT_SUPPORTED;
                case MOSQ_ERR_AUTH:
                        return MOSQUITTO_ERR_AUTH;
                case MOSQ_ERR_ACL_DENIED:
                        return MOSQUITTO_ERR_ACL_DENIED;
                case MOSQ_ERR_UNKNOWN:
                        return MOSQUITTO_ERR_UNKNOWN;
                case MOSQ_ERR_ERRNO:
                        return MOSQUITTO_ERR_ERRNO;
                case MOSQ_ERR_EAI:
                        return MOSQUITTO_ERR_EAI;
                case MOSQ_ERR_PROXY:
                        return MOSQUITTO_ERR_PROXY;
                case MOSQ_ERR_PROTOCOL:
                        return MOSQUITTO_ERR_PROTOCOL;
                case MOSQ_ERR_SUCCESS:
                        return MOSQUITTO_ERR_SUCCESS;
                default:
                        return MOSQUITTO_ERR_OTHERS;
        }
        return MOSQUITTO_ERR_OTHERS;
}

int MosquittoPublish(IN const void* _pInstance, IN int* _pMid, IN char* _pTopic, IN int _nPayloadlen, IN const void* _pPayload)
{
       struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pInstance);
       int rc = mosquitto_publish(pInstance->mosq, NULL, _pTopic, _nPayloadlen, _pPayload, pInstance->options.nQos, pInstance->options.bRetain);
       if (rc) {
               switch (rc) {
                       case MOSQ_ERR_INVAL:
                               fprintf(stderr, "Error: Invalid input. Does your topic contain '+' or '#'?\n");
                               break;
                       case MOSQ_ERR_NOMEM:
                               fprintf(stderr, "Error: Out of memory when trying to publish message.\n");
                               break;
                       case MOSQ_ERR_NO_CONN:
                               fprintf(stderr, "Error: Client not connected when trying to publish.\n");
                               break;
                       case MOSQ_ERR_PROTOCOL:
                               fprintf(stderr, "Error: Protocol error when communicating with broker.\n");
                               break;
                       case MOSQ_ERR_PAYLOAD_SIZE:
                               fprintf(stderr, "Error: Message payload is too large.\n");
                               break;
               }
       }
       return rc;
}

int MosquittoSubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* _pTopic)
{
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pInstance);
        int rc = mosquitto_subscribe(pInstance->mosq, _pMid, _pTopic, pInstance->options.nQos);
        fprintf(stderr, "mos sub %d", rc);
        if (!rc) {
                insertNode(&pInstance->pSubsribeList, _pTopic);
        }
        else {
               switch (rc) {
                       case MOSQ_ERR_INVAL:
                               fprintf(stderr, "Error: Invalid input.\n");
                               break;
                       case MOSQ_ERR_NOMEM:
                               fprintf(stderr, "Error: Out of memory when trying to subscribe message.\n");
                               break;
                       case MOSQ_ERR_NO_CONN:
                               fprintf(stderr, "Error: Client not connected when trying to subscribe.\n");
                               break;
                       case MOSQ_ERR_PROTOCOL:
                               fprintf(stderr, "Error: Protocol error when communicating with broker.\n");
                               break;
                       case MOSQ_ERR_PAYLOAD_SIZE:
                               fprintf(stderr, "Error: Message payload is too large.\n");
                               break;
               }
        }
        return rc;
}

int MosquittoUnsubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* _pSub)
{
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pInstance);
        int rc = mosquitto_unsubscribe(pInstance->mosq, _pMid, _pSub);
        fprintf(stderr, "mos sub %d", rc);
        if (!rc) {
               deleteNode(&pInstance->pSubsribeList, _pSub);
        }
        else {
               switch (rc) {
                       case MOSQ_ERR_INVAL:
                               fprintf(stderr, "Error: Invalid input.\n");
                               break;
                       case MOSQ_ERR_NOMEM:
                               fprintf(stderr, "Error: Out of memory when trying to unsubscribe message.\n");
                               break;
                       case MOSQ_ERR_NO_CONN: 
                               fprintf(stderr, "Error: Client not connected when trying to unsubscribe.\n");
                               break; 
                       case MOSQ_ERR_PROTOCOL:
                               fprintf(stderr, "Error: Protocol error when communicating with broker.\n");
                               break;
               }
        }
        return rc;
}

int MosquittoLibInit()
{
        int rc = mosquitto_lib_init();
        return MosquittoErrorStatusChange(rc);
}

int MosquittoLibCleanup()
{
        int rc = mosquitto_lib_cleanup();
        return MosquittoErrorStatusChange(rc);
}
