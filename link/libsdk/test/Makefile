
ARCH:=x86
THIRD_PARTY_PATH=../../../third_party
OPENSSL_PATH=$(THIRD_PARTY_PATH)/openssl-1.1.0h
OPENSSL_LIB=${OPENSSL_PATH}/output/lib
LIBUA=../output/lib/$(ARCH)/libua.a
LDFLAGS=-L../output/lib/$(ARCH) -L$(OPENSSL_LIB) -lua -lpthread -lm -lcrypto -lssl 
CFLAGS=-g
TARGET_OBJS=test.o unit_test.o call_test.o register_test.o send_pkt_test.o hangup_call_test.o answercall_test.o \
			rejectcall_test.o
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
#	mkdir -p $(TESTCASE_OBJS_DIR) 
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(TESTCASE_OBJS_DIR)/$@

OBJ_FILES = $(addprefix objs/,$(notdir $(TARGET_OBJS)))
vpath %.o $(TESTCASE_OBJS_DIR)
$(TARGET):$(TARGET_OBJS) $(LIBUA)
	$(CC) -o $@ $(OBJ_FILES) $(LDFLAGS)

clean:
	rm -rvf $(TESTCASE_OBJS_DIR)/*.o test

