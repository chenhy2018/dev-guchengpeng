
#CC=arm-linux-gnueabihf-gcc
#ARCH:=x86
THIRD_PARTY_PATH=../../../third_party

ifeq ($(WITH_OPENSSL), ON)
OPENSSL_PATH=$(THIRD_PARTY_PATH)/openssl-1.1.0h
OPENSSL_LIB=${OPENSSL_PATH}/prefix/$(ARCH)/lib
LDFLAGS=-L../../../output/lib/$(ARCH) -L$(OPENSSL_LIB) -lua -lpthread -lm  -lssl -lcrypto -lcurl -lrtmp_sdk -lev -ldevsdk -ltools -lstdc++ -lfdk-aac
else
LDFLAGS=-L../../../output/lib/$(ARCH) -lua -lpthread -lm -lcurl -lrtmp_sdk -lev -ldevsdk -ltools -lstdc++ -lfdk-aac
endif

CFLAGS=-g -std=gnu99
TARGET_OBJS= main.o dbg.o stream.o h264_parse.o command.o dev_core.o aj_ipc.o
OBJS_DIR=./objs
INCLUDE_PATH=-I./inc
TARGET=sdk-demo

all:$(TARGET)

$(TARGET_OBJS) : %.o : %.c
	mkdir -p $(OBJS_DIR) 
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(OBJS_DIR)/$@

OBJ_FILES = $(addprefix objs/,$(notdir $(TARGET_OBJS)))
vpath %.o $(OBJS_DIR)
$(TARGET):$(TARGET_OBJS)
	$(CC) -o $@ $(OBJ_FILES) $(LDFLAGS)
	cp sdk-demo /home/share/

clean:
	rm -rvf $(OBJS_DIR)/*.o sdk-demo

