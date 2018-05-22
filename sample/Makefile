
CC=arm-linux-gnueabihf-gcc
PROJECT_PATH=../third_party
PJSIP_PATH=$(PROJECT_PATH)/pjproject-2.7.2
LDFLAGS=-L $(PJSIP_PATH)/pjmedia/lib -L $(PJSIP_PATH)/pjsip/lib -L $(PJSIP_PATH)/pjnath/lib -L $(PJSIP_PATH)/pjlib-util/lib \
		-L $(PJSIP_PATH)/pjlib/lib \
		-lpjmedia-arm-unknown-linux-gnueabihf -lpjmedia-codec-arm-unknown-linux-gnueabihf \
		-lpjmedia-audiodev-arm-unknown-linux-gnueabihf -lpjsip-arm-unknown-linux-gnueabihf \
		-lpjsdp-arm-unknown-linux-gnueabihf -lpj-arm-unknown-linux-gnueabihf \
		-lpjlib-util-arm-unknown-linux-gnueabihf -lpjsua-arm-unknown-linux-gnueabihf \
		-lpjsip-ua-arm-unknown-linux-gnueabihf -lpthread -lm
INCLUDE_PATH=-I../third_party/pjproject-2.7.2/pjlib/include -I../third_party/pjproject-2.7.2/pjlib-util/include -I../third_party/pjproject-2.7.2/pjsip/include -I../third_party/pjproject-2.7.2/pjmedia/include -I../third_party/pjproject-2.7.2/pjnath/include

sip-demo:sip.o main.o
	$(CC) main.o sip.o -o sip-demo $(LDFLAGS)

sip.o:sip.c
	$(CC) $(INCLUDE_PATH) -c sip.c -o sip.o

main.o:main.c
	$(CC) $(INCLUDE_PAT) -c main.c -o main.o

clean:
	rm *.o sip-demo

