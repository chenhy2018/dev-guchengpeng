LIBUA=../../../output/lib/$(ARCH)/libua.a
THIRD_PARTY_PATH=../../../third_party
OPENSSL_PATH=$(THIRD_PARTY_PATH)/openssl-1.1.0h
OPENSSL_LIB=${OPENSSL_PATH}/prefix/$(ARCH)/lib
LDFLAGS=-L../../../output/lib/$(ARCH) -L$(OPENSSL_LIB) -lua -lpthread -lm  -lssl -lcrypto
CFLAGS=-g -std=gnu99
TARGET_OBJS=test.o unit_test.o
TESTCASE_OBJS_DIR=./objs
INCLUDE_PATH=-I../
TARGET=test

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

