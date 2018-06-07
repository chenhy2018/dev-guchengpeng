#include "mqtt.h"
#include <unistd.h>
#include <string.h>
#include <stdio.h>

struct connect_status {
        int status;
        void* pInstance;
} connect_status;

struct connect_status Status[10];

void OnMessage(IN const void* _pInstance, IN const char* _pTopic, IN const char* _pMessage, IN size_t nLength)
{
        fprintf(stderr, "%p topic %s message %s \n", _pInstance, _pTopic, _pMessage);
}

void OnEvent(IN const void* _pInstance, IN int _nId,  IN const char* _pReason)
{
        fprintf(stderr, "%p id %d, reason  %s \n",_pInstance, _nId, _pReason);
        struct connect_status* pStatus;
        for (int i = 0; i < 10; i++) {
                if (Status[i].pInstance == _pInstance) {
                        pStatus = &Status[i];
                }
        }
        pStatus->status = _nId;
}

int main()
{
        struct MqttOptions options;
        MqttLibInit();
        options.pId = "test";
        options.bCleanSession = false;
        options.primaryUserInfo.nAuthenicatinMode = MQTT_AUTHENTICATION_USER;
        options.primaryUserInfo.pHostname = "123.59.204.198";
        //strcpy(options.bindaddress, "172.17.0.2");
        options.secondaryUserInfo.pHostname = "172.17.0.4";
        //strcpy(options.secondBindaddress, "172.17.0.2`");
        options.primaryUserInfo.pUsername = "test_sub";
        options.primaryUserInfo.pPassword = "testsub1";
        options.secondaryUserInfo.pUsername = "test";
        options.secondaryUserInfo.pPassword = "111";
        options.secondaryUserInfo.nPort = 1883;
        options.primaryUserInfo.nPort = 1883;
        options.primaryUserInfo.pCafile = NULL;
        options.primaryUserInfo.pCertfile = NULL;
        options.primaryUserInfo.pKeyfile = NULL;
        options.secondaryUserInfo.pCafile = NULL;
        options.secondaryUserInfo.pCertfile = NULL;
        options.secondaryUserInfo.pKeyfile = NULL;
        options.nKeepalive = 10;
        options.nQos = 0;
        options.bRetain = false;
        options.callbacks.OnMessage = &OnMessage;
        options.callbacks.OnEvent = &OnEvent;
        void* instance = NULL;
        printf("try first sub \n");
        instance = MqttCreateInstance(&options);
        Status[0].pInstance = instance;
        while (!(Status[0].status & 3000)) {
                sleep(1);
        }
        MqttSubscribe(instance, "test/#");
        printf("try pub %p \n", instance);
        options.pId = "pubtest";
        options.primaryUserInfo.pUsername = "test_pub";
        options.primaryUserInfo.pPassword = "testpub";
        void* pubInstance = MqttCreateInstance(&options);
        Status[1].pInstance = pubInstance;
        while (!(Status[1].status & 3000)) {
                sleep(1);
        }
        for (int i = 0 ; i < 10; ++i) {
            MqttPublish(pubInstance, "test/pub", 10, "test_pub");
            MqttPublish(pubInstance, "test/pub3", 10, "test_pub3");
            
        }
        sleep(10);
        Status[1].pInstance = NULL;
        Status[0].pInstance = NULL;
        Status[1].status = 0;
        Status[0].status = 0;
        MqttDestroy(instance);
        MqttDestroy(pubInstance);
        printf("try second \n");
        options.primaryUserInfo.nAuthenicatinMode = MQTT_AUTHENTICATION_USER;
        options.primaryUserInfo.pHostname = "123.59.204.198";
        options.primaryUserInfo.pUsername = "root";
        options.primaryUserInfo.pPassword = "root";
        instance = MqttCreateInstance(&options);
        Status[1].pInstance = instance;
        while (!(Status[1].status & 3000)) {
                sleep(1);
        }
        MqttSubscribe(instance, "sensor/room1/#");
        //usleep(100000);
        for (int i = 0 ; i < 100; ++i) {
              MqttPublish(instance, "sensor/room1/temperature", 10, "test1234456");
        }
        sleep(5);
        Status[1].pInstance = NULL;
        MqttDestroy(instance);
        MqttLibCleanup();
        printf("try third \n");
        //memset(options.primaryUserInfo.hostname, 0, MAX_MQTT_USR_SIZE);
        //strcpy(options.primaryUserInfo.hostname, "172.17.0.8");
        options.primaryUserInfo.pCafile = "./test/ca.crt";
        options.primaryUserInfo.nPort = 8883;
        options.primaryUserInfo.nAuthenicatinMode = MQTT_AUTHENTICATION_USER | MQTT_AUTHENTICATION_ONEWAY_SSL;
        MqttLibInit();
        while(1) {
                //MqttLibInit();
                instance = MqttCreateInstance(&options);
                Status[0].pInstance = instance;
                while (!(Status[0].status & 3000)) {
                        sleep(1);
                }
                MqttSubscribe(instance, "sensor/room1/#");
                for (int i = 0; i < 1000000; ++ i) {
                        MqttPublish(instance, "sensor/room1/temperature", 10, "test1234456");
                        usleep(100000);
                }
                sleep(3);
                Status[0].pInstance = NULL;
                MqttDestroy(instance);
                //MqttLibCleanup();
        }
        MqttLibCleanup();
        return 1;
}
