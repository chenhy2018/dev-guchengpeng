# Makefile for fastfwd

PROTOC = protoc
GRPC_CPP_PLUGIN = grpc_cpp_plugin
GRPC_CPP_PLUGIN_PATH ?= `which $(GRPC_CPP_PLUGIN)`
PROTOS_PATH = ./protos
vpath %.proto $(PROTOS_PATH)

RELEASEP=../bin

INCP=../../../release/fastfwd/include
LIBP=../../../release/fastfwd/lib/x86_64
LIBP_FFMPEG=${LIBP}/ffmpeg
LIBP_GRPC=${LIBP}/grpc
LIBP_GLIBC=${LIBP}/glibc
CXX = g++
CXXFLAGS += -I$(INCP) -std=c++14
LDFLAGS += -L$(LIBP_FFMPEG) -lavfilter -lavformat -lavcodec -lavutil -lswscale -lswresample \
    -L$(LIBP_GRPC) -lprotobuf -lgrpc++ -lgrpc -lgrpc++_reflection \
    -L$(LIBP_GLIBC) $(LIBP_GLIBC)/libm-2.27.a -Wl,--whole-archive $(LIBP_GLIBC)/libpthread.a -Wl,--no-whole-archive $(LIBP_GLIBC)/libc.a $(LIBP_GLIBC)/libdl.a \
    -static -static-libstdc++

TARGETS = linking_fastfwd
build: $(TARGETS)

linking_fastfwd: fast_forward.pb.o fast_forward.grpc.pb.o linking_fastfwd.o fastfwd.o
	$(CXX) $^ $(LDFLAGS) -o $@

clean:
	@rm -f *.o $(TARGETS)

release: build
	@echo "release binary"
	@strip ${TARGETS}
	@mkdir -p ${RELEASEP}
	@mv ${TARGETS} ${RELEASEP}

proto: fast_forward.proto
	@$(PROTOC) -I $(PROTOS_PATH) --grpc_out=. --plugin=protoc-gen-grpc=$(GRPC_CPP_PLUGIN_PATH) $<
	@$(PROTOC) -I $(PROTOS_PATH) --cpp_out=. $<

%.o : %.cpp
	@echo "CXX" $<
	@$(CXX) $(CXXFLAGS) -c $<

%.o : %.cc
	@echo "CXX" $<
	@$(CXX) $(CXXFLAGS) -c $<

