bmake:
	g++ bmake.cpp -o bmake

clean:
	rm -f bmake

install: bmake
	@mkdir -p ~/bin
	cp bmake ~/bin
	cp bmake ${GOROOT}/bin

