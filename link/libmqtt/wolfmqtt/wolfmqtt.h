#ifndef __WOLF_MQTT__
#define __WOLF_MQTT__

#include <stdbool.h>
#include <stddef.h>
#include "mqtt_client.h"
#include "mqttnet.h"
#include "../mqtt.h"
#include "../mqtt_internal.h"

typedef struct MQTTCtx MqttCtx;

#define SOCK_ADDR_IN    struct sockaddr_in

typedef enum {
        SOCK_BEGIN = 0,
        SOCK_CONN,
} NB_Stat;

typedef struct _SocketContext {
        SOCKET_T fd;
        NB_Stat stat;
        SOCK_ADDR_IN addr;
}SocketContext;

typedef struct MQTTCtx {
        /* client and net containers */
        MqttClient client;
        MqttNet net;
	void* pInstance;
        /* temp mqtt containers */
        MqttConnect connect;
        MqttMessage lwt_msg;
        MqttSubscribe subscribe;
        MqttUnsubscribe unsubscribe;
        MqttTopic topics[10];
        MqttPublish publish;
        MqttDisconnect disconnect;

        byte *tx_buf, *rx_buf;
        word32 cmd_timeout_ms;
	int timeoutCount;
} MQTTCtx;

int ClientOptSet(struct MqttInstance* _pInstance, struct MqttUserInfo info);

#endif
