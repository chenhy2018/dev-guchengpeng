# Makefile for fastfwd test

SRC_PATH=../src
vpath %.o $(SRC_PATH)
INCP=../../../release/fastfwd/include
LIBP=../../../release/fastfwd/lib/x86_64
LIBP_FFMPEG=${LIBP}/ffmpeg
LIBP_GRPC=${LIBP}/grpc
LIBP_GLIBC=${LIBP}/glibc
CXX = g++
CXXFLAGS += -I$(SRC_PATH) -I$(INCP) -std=c++14
LDFLAGS1 = -L$(LIBP_FFMPEG) -lavfilter -lavformat -lavcodec -lavutil -lswscale -lswresample
LDFLAGS2 = -L$(LIBP_GRPC) -lprotobuf -lgrpc++ -lgrpc -lgrpc++_reflection
LDFLAGS3 = -L$(LIBP_GLIBC) $(LIBP_GLIBC)/libm-2.27.a -Wl,--whole-archive $(LIBP_GLIBC)/libpthread.a -Wl,--no-whole-archive $(LIBP_GLIBC)/libc.a $(LIBP_GLIBC)/libdl.a
LDFLAGS4 = -static -static-libstdc++

TARGETS = fastfwd_file_test fastfwd_grpc_client fastfwd_load_test
test: $(TARGETS)

fastfwd_file_test: fastfwd_file_test.o
	make -C ${SRC_PATH} build
	$(CXX) ${SRC_PATH}/fastfwd.o $^ $(LDFLAGS1) $(LDFLAGS3) $(LDFLAGS4) -o $@

fastfwd_grpc_client: fastfwd_grpc_client.o
	make -C ${SRC_PATH} build
	$(CXX) ${SRC_PATH}/fast_forward.pb.o  ${SRC_PATH}/fast_forward.grpc.pb.o $^ $(LDFLAGS2) $(LDFLAGS3) $(LDFLAGS4) -o $@

fastfwd_load_test: fastfwd_load_test.o
	make -C ${SRC_PATH} build
	$(CXX) ${SRC_PATH}/fast_forward.pb.o  ${SRC_PATH}/fast_forward.grpc.pb.o $^ $(LDFLAGS2) $(LDFLAGS3) $(LDFLAGS4) -o $@

clean:
	make -C ${SRC_PATH} clean
	@rm -f *.o $(TARGETS)

%.o : %.cpp
	@echo "CXX" $<
	@$(CXX) $(CXXFLAGS) -c $<

%.o : %.cc
	@echo "CXX" $<
	@$(CXX) $(CXXFLAGS) -c $<
