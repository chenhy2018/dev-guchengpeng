#ifndef __MOS_MQTT__
#define __MOS_MQTT__

#include <stdbool.h>
#include <stddef.h>
#include <mosquitto.h>
#include "../mqtt.h"
#include "../mqtt_internal.h"

typedef struct MQTTCtx mosquitto;

int ClientOptSet(struct MqttInstance* _pInstance, struct MqttUserInfo info);

#endif
