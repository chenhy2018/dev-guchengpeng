
THIRD_PARTY_PATH=../../../third_party

ifeq ($(WITH_MOSQUITTO), ON)
SSL_PATH=$(THIRD_PARTY_PATH)/openssl-1.1.0h
SSL_LIB=${OPENSSL_PATH}/prefix/$(ARCH)/lib
LDFLAGS=-L../../../output/lib/$(ARCH) -L$(SSL_LIB) -lua -lpthread -lm -lcrypto -lssl 
else
SSL_PATH=$(THIRD_PARTY_PATH)/wolfssl
MQTT_PATH=$(THIRD_PARTY_PATH)/wolfMQTT
SSL_LIB=${SSL_PATH}/prefix/$(ARCH)/lib
MQTT_LIB=${MQTT_PATH}/prefix/$(ARCH)/lib
LDFLAGS=-L../../../output/lib/$(ARCH) -L$(SSL_LIB) -L$(MQTT_LIB) -lua -lpthread -lm -lwolfssl -lwolfmqtt
endif

CFLAGS=-g -std=gnu99

ifeq ($(WITH_P2P), ON)
CFLAGS=-g -std=gnu99 -D WITH_P2P
TARGET_OBJS=test.o unit_test.o call_test.o register_test.o send_pkt_test.o hangup_call_test.o answercall_test.o \
			rejectcall_test.o
else
CFLAGS=-g -std=gnu99
TARGET_OBJS=test.o unit_test.o call_test.o register_test.o  hangup_call_test.o answercall_test.o \
                        rejectcall_test.o
endif
TESTCASE_OBJS_DIR=./objs
INCLUDE_PATH=-I../ -I../../libsip/
TARGET=test

ifeq ($(ARCH), x86)
	CC=gcc
else ifeq ($(ARCH), mstar)
	CC=arm-linux-gnueabihf-gcc
endif


all:$(TARGET)


$(TARGET_OBJS) : %.o : %.c
	mkdir -p $(TESTCASE_OBJS_DIR) 
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(TESTCASE_OBJS_DIR)/$@

OBJ_FILES = $(addprefix objs/,$(notdir $(TARGET_OBJS)))
vpath %.o $(TESTCASE_OBJS_DIR)
$(TARGET):$(TARGET_OBJS) $(LIBUA)
	$(CC) -o $@ $(OBJ_FILES) $(LDFLAGS)

clean:
	rm -rvf $(TESTCASE_OBJS_DIR)/*.o test

