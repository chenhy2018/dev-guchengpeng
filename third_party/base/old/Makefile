TARGETS={ patch/errors patch/multipart patch/os patch/io 3rd com service db shell devtools/qcheckout devtools/qbuild }

all: devtool
	@bmake ${TARGETS}

install: devtool
	@bmake ${TARGETS} install

test: devtool
	@bmake ${TARGETS} test

clean: devtool
	@bmake ${TARGETS} clean
	cd devtools/bmake; make clean

devtool:
	cd devtools/bmake; make install
	cd devtools/qdelete; make install
	cd devtools/qinstall; make install
