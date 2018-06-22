#include "mqtt.h"
#include "mqtt_internal.h"

static int MqttErrorStatusChange(int nStatus)
{
        switch (nStatus) {
                case MOSQ_ERR_CONN_PENDING:
                        return MQTT_ERR_CONN_PENDING;
                case MOSQ_ERR_NOMEM:
                        return MQTT_ERR_NOMEM;
                case MOSQ_ERR_INVAL:
                        return MQTT_ERR_INVAL;
                case MOSQ_ERR_NO_CONN:
                        return MQTT_ERR_NO_CONN;
                case MOSQ_ERR_CONN_REFUSED:
                        return MQTT_ERR_CONN_REFUSED;
                case MOSQ_ERR_NOT_FOUND:
                        return MQTT_ERR_NOT_FOUND;
                case MOSQ_ERR_CONN_LOST:
                        return MQTT_ERR_CONN_LOST;
                case MOSQ_ERR_TLS:
                        return MQTT_ERR_TLS;
                case MOSQ_ERR_PAYLOAD_SIZE:
                        return MQTT_ERR_PAYLOAD_SIZE;
                case MOSQ_ERR_NOT_SUPPORTED:
                        return MQTT_ERR_NOT_SUPPORTED;
                case MOSQ_ERR_AUTH:
                        return MQTT_ERR_AUTH;
                case MOSQ_ERR_ACL_DENIED:
                        return MQTT_ERR_ACL_DENIED;
                case MOSQ_ERR_UNKNOWN:
                        return MQTT_ERR_UNKNOWN;
                case MOSQ_ERR_ERRNO:
                        return MQTT_ERR_ERRNO;
                case MOSQ_ERR_EAI:
                        return MQTT_ERR_EAI;
                case MOSQ_ERR_PROXY:
                        return MQTT_ERR_PROXY;
                case MOSQ_ERR_PROTOCOL:
                        return MQTT_ERR_PROTOCOL;
                case MOSQ_ERR_SUCCESS:
                        return MQTT_SUCCESS;
                default:
                        return MQTT_ERR_OTHERS;
        }
        return MQTT_ERR_OTHERS;
}

static void MallocAndStrcpy(char** des, const char* src)
{
      if (src) {
              *des = malloc(strlen(src) + 1);
              if (*des) {
                      strncpy(*des, src, strlen(src));
                      des[strlen(src)] = '\0';
              }
      }
      else {
              *des = NULL;
      }
}

static void SafeFree(char* des)
{
      if (des) free(des);
}

static bool InsertNode(Node* pHead, char* val) {
        Node* p = pHead;
        while(p->pNext) {
                p = p->pNext;
        }
        Node* pNew = malloc(sizeof(Node));
        strcpy(pNew->topic, val);
        p->pNext = pNew;
        pNew->pNext = NULL;
        return true;
}

static bool DeleteNode(Node* PHead, char * pval)
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

static void ClearNode(Node* PHead)
{
        int i = 0;
        Node* p = PHead;
        while (p->pNext != NULL) {
                Node* temp = p->pNext;
                p->pNext = temp->pNext;
                free(temp);
        }
}

void OnEventCallback(struct MqttInstance* _pInstance, int rc, const char* _pStr)
{
        if (_pInstance->options.callbacks.OnEvent != NULL) {
                _pInstance->options.callbacks.OnEvent(_pInstance, _pInstance->options.nAccountId, rc, _pStr);
        }
}

void OnLogCallback(struct mosquitto* _pMosq, void* _pObj, int level, const char* _pStr)
{
        printf("%s\n", _pStr);
}

void OnMessageCallback(struct mosquitto* _pMosq, void* _pObj, const struct mosquitto_message* _pMessage)
{
        int rc = MOSQ_ERR_SUCCESS;
        struct MqttInstance* pInstance = (struct MqttInstance*)(_pObj);
        if (pInstance->options.callbacks.OnMessage) {
                pInstance->options.callbacks.OnMessage(_pObj, pInstance->options.nAccountId,  _pMessage->topic, _pMessage->payload, _pMessage->payloadlen);
        }
}

void OnConnectCallback(struct mosquitto* _pMosq, void* _pObj, int result)
{
        int rc = MOSQ_ERR_SUCCESS;
        struct MqttInstance* pInstance = (struct MqttInstance*)(_pObj);
        pInstance->connected = true;
        if (result) {
                pInstance->status = STATUS_CONNECT_ERROR;
        }
        else {
                pInstance->status = STATUS_CONNACK_RECVD;
                pthread_mutex_lock(&pInstance->listMutex);
                Node* p = pInstance->pSubsribeList.pNext;
                while (p) {
                        mosquitto_subscribe(pInstance->mosq, NULL, p->topic, pInstance->options.nQos);
                        p = p->pNext;
                }
                pthread_mutex_unlock(&pInstance->listMutex);
        }
        OnEventCallback(pInstance,
                        (result == 0) ? MQTT_CONNECT_SUCCESS : MqttErrorStatusChange(result),
                        (result == 0) ? "on connect success" : mosquitto_connack_string(result));
}


void OnDisconnectCallback(struct mosquitto* _pMosq, void* _pObj, int rc)
{

        struct MqttInstance* pInstance = (struct MqttInstance*)(_pObj);
        pInstance->connected = false;
        if (!rc) {
                pInstance->status = STATUS_IDLE;
        }
        OnEventCallback(pInstance,
                       (rc == 0) ? MQTT_DISCONNECT_SUCCESS : MqttErrorStatusChange(rc),
                       (rc == 0) ? "on disconnect success" : mosquitto_connack_string(rc));
}

void OnSubscribeCallback(struct mosquitto* _pMosq, void* pObj, int mid, int qos_count, const int* pGranted_qos)
{       
        fprintf(stderr, "Subscribed (mid: %d): %d \n", mid, pGranted_qos[0]);
}

void OnUnsubscribeCallback(struct mosquitto* _pMosq, void* _pObj, int mid)
{
        fprintf(stderr, "Unsubscribed (mid: %d) \n", mid);
}

void OnPublishCallback(struct mosquitto* _pMosq, void* _pObj, int mid)
{
        fprintf(stderr, " my_publish_callback \n ");
        struct MqttInstance* pInstance = (struct MqttInstance*)(_pObj);
        int last_mid_sent = mid;
}


static void MqttInstanceInit(struct MqttInstance* _pInstance, const struct MqttOptions* _pOption)
{
        /* copy options */
        memcpy(&_pInstance->options, _pOption, sizeof(struct MqttOptions));
        MallocAndStrcpy(&_pInstance->options.pId, _pOption->pId);
        MallocAndStrcpy(&_pInstance->options.primaryUserInfo.pUsername, _pOption->primaryUserInfo.pUsername);
        MallocAndStrcpy(&_pInstance->options.primaryUserInfo.pPassword, _pOption->primaryUserInfo.pPassword);
        MallocAndStrcpy(&_pInstance->options.primaryUserInfo.pHostname, _pOption->primaryUserInfo.pHostname);
        MallocAndStrcpy(&_pInstance->options.primaryUserInfo.pCafile, _pOption->primaryUserInfo.pCafile);
        MallocAndStrcpy(&_pInstance->options.primaryUserInfo.pCertfile, _pOption->primaryUserInfo.pCertfile);
        MallocAndStrcpy(&_pInstance->options.primaryUserInfo.pKeyfile, _pOption->primaryUserInfo.pKeyfile);
        MallocAndStrcpy(&_pInstance->options.secondaryUserInfo.pUsername, _pOption->secondaryUserInfo.pUsername);
        MallocAndStrcpy(&_pInstance->options.secondaryUserInfo.pPassword, _pOption->secondaryUserInfo.pPassword);
        MallocAndStrcpy(&_pInstance->options.secondaryUserInfo.pHostname, _pOption->secondaryUserInfo.pHostname);
        MallocAndStrcpy(&_pInstance->options.secondaryUserInfo.pCafile, _pOption->secondaryUserInfo.pCafile);
        MallocAndStrcpy(&_pInstance->options.secondaryUserInfo.pCertfile, _pOption->secondaryUserInfo.pCertfile);
        MallocAndStrcpy(&_pInstance->options.secondaryUserInfo.pKeyfile, _pOption->secondaryUserInfo.pKeyfile);
        _pInstance->mosq = NULL;
        _pInstance->connected = false;
        _pInstance->status = STATUS_IDLE;
        _pInstance->isDestroying = false;
        _pInstance->pSubsribeList.pNext = NULL;
        pthread_mutex_init(&_pInstance->listMutex, NULL);
}

bool ClientOptSet(struct MqttInstance* _pInstance, struct mosquitto* _pMosq, struct MqttUserInfo info)
{
        int rc = 0;
        if (info.nAuthenicatinMode & MQTT_AUTHENTICATION_USER) {
                rc = mosquitto_username_pw_set(_pMosq, info.pUsername, info.pPassword);
                if (rc)
                        return rc;
        }
        if (info.nAuthenicatinMode & MQTT_AUTHENTICATION_ONEWAY_SSL) {
                rc = mosquitto_tls_set(_pMosq, info.pCafile, NULL, NULL, NULL, NULL);
        }
        else if (info.nAuthenicatinMode & MQTT_AUTHENTICATION_TWOWAY_SSL) {
                rc = mosquitto_tls_set(_pMosq, info.pCafile, NULL, info.pCertfile, info.pKeyfile, NULL);
        }
        if (rc) {
                printf("ClientOptSet error %d\n", rc);
        }
        return rc;
}

void * Mqttthread(void* _pData)
{
        int rc;
        
        struct MqttInstance* pInstance = (struct MqttInstance*)(_pData);

        pInstance->mosq = mosquitto_new(pInstance->options.pId, true, pInstance);
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
        //mosquitto_log_callback_set(pInstance->mosq, OnLogCallback);
        mosquitto_connect_callback_set(pInstance->mosq, OnConnectCallback);
        mosquitto_disconnect_callback_set(pInstance->mosq, OnDisconnectCallback);
        mosquitto_publish_callback_set(pInstance->mosq, OnPublishCallback);
        mosquitto_message_callback_set(pInstance->mosq, OnMessageCallback);
        mosquitto_subscribe_callback_set(pInstance->mosq, OnSubscribeCallback);
        mosquitto_unsubscribe_callback_set(pInstance->mosq, OnUnsubscribeCallback);
        
        do {
                 if (!pInstance->connected && pInstance->status == STATUS_IDLE) {
                         pInstance->status = STATUS_CONNECTING;
                         rc = ClientOptSet(pInstance, pInstance->mosq, pInstance->options.primaryUserInfo);
                         if (rc == 0) {
                                 rc = mosquitto_connect(pInstance->mosq, pInstance->options.primaryUserInfo.pHostname, pInstance->options.primaryUserInfo.nPort, pInstance->options.nKeepalive);
                         }
                         if (rc) {
                                 OnEventCallback(pInstance, rc, mosquitto_strerror(rc));
                                 rc = ClientOptSet(pInstance, pInstance->mosq, pInstance->options.secondaryUserInfo);
                                 if (rc == 0) {
                                         rc = mosquitto_connect(pInstance->mosq, pInstance->options.secondaryUserInfo.pHostname, pInstance->options.secondaryUserInfo.nPort, pInstance->options.nKeepalive);
                                 }
                                 if (rc) {
                                         pInstance->status = STATUS_IDLE;
                                         OnEventCallback(pInstance, rc, mosquitto_strerror(rc));
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
        ClearNode(&pInstance->pSubsribeList);
        SafeFree(pInstance->options.pId);
        SafeFree(pInstance->options.primaryUserInfo.pUsername);
        SafeFree(pInstance->options.primaryUserInfo.pPassword);
        SafeFree(pInstance->options.primaryUserInfo.pHostname);
        SafeFree(pInstance->options.primaryUserInfo.pCafile);
        SafeFree(pInstance->options.primaryUserInfo.pCertfile);
        SafeFree(pInstance->options.primaryUserInfo.pKeyfile);
        SafeFree(pInstance->options.secondaryUserInfo.pUsername);
        SafeFree(pInstance->options.secondaryUserInfo.pPassword);
        SafeFree(pInstance->options.secondaryUserInfo.pHostname);
        SafeFree(pInstance->options.secondaryUserInfo.pCafile);
        SafeFree(pInstance->options.secondaryUserInfo.pCertfile);
        SafeFree(pInstance->options.secondaryUserInfo.pKeyfile);
        pthread_mutex_destroy(&pInstance->listMutex);
        if (pInstance) free(pInstance);
        if (rc) {
                fprintf(stderr, "Error: %s\n", mosquitto_strerror(rc));
        }
        return NULL;
}

void* MqttCreateInstance(IN const struct MqttOptions* pOption)
{
        /* allocate one mosquitto instance struct */
        struct MqttInstance* pInstance = (struct MqttInstance*)malloc(sizeof(struct MqttInstance));
        if (pInstance == NULL) {
                return NULL;
        }
        
        MqttInstanceInit(pInstance, pOption);
        pthread_t t;
        pthread_attr_t attr;
        pthread_attr_init(&attr);
        pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_DETACHED);
        pthread_create(&t, &attr, Mqttthread, pInstance);
        return pInstance;
}

void MqttDestroy(IN const void* _pInstance)
{	
        struct MqttInstance* pInstance = (struct MqttInstance*)(_pInstance);
        pInstance->isDestroying = true;;
        pInstance->options.callbacks.OnMessage = NULL;
        pInstance->options.callbacks.OnEvent = NULL;
}

int MqttPublish(IN const void* _pInstance, IN char* _pTopic, IN int _nPayloadlen, IN const void* _pPayload)
{
       struct MqttInstance* pInstance = (struct MqttInstance*)(_pInstance);
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
       return MqttErrorStatusChange(rc);
}

int MqttSubscribe(IN const void* _pInstance, IN char* _pTopic)
{
        struct MqttInstance* pInstance = (struct MqttInstance*)(_pInstance);
        if (_pTopic == NULL) {
               return MQTT_ERR_INVAL;
        }
        int rc = mosquitto_subscribe(pInstance->mosq, NULL, _pTopic, pInstance->options.nQos);
        if (!rc) {
                pthread_mutex_lock(&pInstance->listMutex);
                InsertNode(&pInstance->pSubsribeList, _pTopic);
                pthread_mutex_unlock(&pInstance->listMutex);
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
        return MqttErrorStatusChange(rc);
}

int MqttUnsubscribe(IN const void* _pInstance, IN char* _pTopic)
{
        struct MqttInstance* pInstance = (struct MqttInstance*)(_pInstance);
        if (_pTopic == NULL) {
               fprintf(stderr, "Error: Invalid input.\n");
               return MQTT_ERR_INVAL;
        }
        int rc = mosquitto_unsubscribe(pInstance->mosq, NULL, _pTopic);
        fprintf(stderr, "mos sub %d", rc);
        if (!rc) {
               pthread_mutex_lock(&pInstance->listMutex);
               DeleteNode(&pInstance->pSubsribeList, _pTopic);
               pthread_mutex_unlock(&pInstance->listMutex);
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
        return MqttErrorStatusChange(rc);
}

int MqttLibInit()
{
        int rc = mosquitto_lib_init();
        return MqttErrorStatusChange(rc);
}

int MqttLibCleanup()
{
        int rc = mosquitto_lib_cleanup();
        return MqttErrorStatusChange(rc);
}
