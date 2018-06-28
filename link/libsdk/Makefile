ARCH:=x86
HOST:=
ifeq ($(ARCH), x86)
	CC=gcc
else ifeq ($(ARCH), mstar)
	CC=arm-linux-gnueabihf-gcc
	HOST=arm-linux-gnueabihf
else ifeq ($(ARCH), A12)
	CC=arm-none-linux-gnueabi-gcc
	HOST=arm-none-linux-gnueabi
else
	CC=gcc
endif

CURRENT_DIR:=$(shell pwd)
PROJECT_PATH=../../third_party
PJSIP_PATH=$(PROJECT_PATH)/pjproject-2.7.2
LDFLAGS=
CFLAGS  = -g

INSTALL_DIR=$(CURRENT_DIR)/output/pjsip
PJ_CONFIG=--disable-sdl --disable-ffmpeg --disable-v4l2 \
		  --disable-openh264 --disable-libwebrtc --disable-sound \
		  --disable-libyuv CFLAGS='-g' \
		  --prefix=$(INSTALL_DIR)
INCLUDE_PATH=-I$(PJSIP_PATH)/pjlib/include \
			 -I$(PJSIP_PATH)/pjlib-util/include \
			 -I$(PJSIP_PATH)/pjsip/include \
			 -I$(PJSIP_PATH)/pjmedia/include \
			 -I$(PJSIP_PATH)/pjnath/include \
			 -I../libsip \
			 -I ../libmqtt \
			 -I ../librtp \
			 -I $(PROJECT_PATH)/mosquitto-1.4.8/lib/ 

SIP_PATH=../libsip
RTP_PATH=../librtp
MQTT_PATH=../libmqtt
OUTPUT=output
SIP_OUT_PATH=$(OUTPUT)/libsip
RTP_OUT_PATH=$(OUTPUT)/librtp
MQTT_OUT_PATH=$(OUTPUT)/libmqtt
SDK_OBJS_PATH=$(OUTPUT)/libsdk
LIBSDK=queue.o sdk_interface.o \
       test.o dbg.o uaMgr.o callMgr.o
LIBSIP=sip.o sip_internal.o
LIBRTP=JitterBuffer.o MediaStream.o PeerConnection.o heap.o
TARGET_LIBRARY=$(OUTPUT)/lib/$(ARCH)/libua.a
LIBPJ=$(OUTPUT)/ori/$(ARCH)
LIBMQTT=$(MQTT_OUT_PATH)/mqtt.o
LIBSIP_OBJS_PATH=$(OUTPUT)/libsip
LIBRTP_OBJS_PATH=$(OUTPUT)/librtp
LIBSIP_SOURCE_PATH=../libsip
LIBRTP_SOURCE_PATH=../librtp

all : $(TARGET_LIBRARY)

depclean:
	cd $(PJSIP_PATH); \
	find . -name "*.depend" -print -delete  

pjclean :
	rm -rvf $(OUTPUT)/pjsip
	rm -rvf $(LIBPJ)

pjsip $(LIBPJ) : #pjclean depclean
	@echo "compile pjsip"
	mkdir -p $(INSTALL_DIR)
	cd $(PJSIP_PATH); \
	make distclean && \
	make clean && \
	make realclean && \
	cd $(PJSIP_PATH); \
	if [ "$(HOST)" = "" ]; then \
		./configure $(PJ_CONFIG); \
	else \
		./configure --host=$(HOST) $(PJ_CONFIG); \
	fi \
	&& make dep \
	&& make clean \
	&& make \
	&& make install


vpath %.o $(OUTPUT)
vpath %.c $(LIBSIP_SOURCE_PATH)
$(LIBSIP) : %.o : %.c
	mkdir -p $(LIBSIP_OBJS_PATH) 
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(LIBSIP_OBJS_PATH)/$@

vpath %.c $(LIBRTP_SOURCE_PATH)
$(LIBRTP) : %.o : %.c
	mkdir -p $(LIBRTP_OBJS_PATH) 
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(LIBRTP_OBJS_PATH)/$@

$(LIBMQTT):$(MQTT_PATH)/mqtt.c
	mkdir -p $(MQTT_OUT_PATH)
	$(CC) $(INCLUDE_PATH) -c $< -o $@


# set .o file search path
#vpath %.o $(SDK_OBJS_PATH)
sdk $(TARGET_LIBRARY) : $(LIBPJ) $(LIBSIP) $(LIBSDK) $(LIBMQTT) $(LIBRTP)
	./merge-lib.sh $(ARCH)
	make -C test clean
	make -C test ARCH=$(ARCH)
	make -C test/callee clean
	make -C test/callee ARCH=$(ARCH)

$(LIBSDK) : %.o : %.c
	mkdir -p $(SDK_OBJS_PATH) 
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(SDK_OBJS_PATH)/$@
libsdk : $(LIBSDK)

clean:
	make -C test clean
	rm -rvf output/libmqtt
	rm -rvf output/libsdk
	rm -rvf output/libsip
	rm -rvf output/librtp


