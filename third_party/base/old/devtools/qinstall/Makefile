all:
	gcc qinstall.c -o qinstall

clean:
	rm -f qinstall

install: all
	@mkdir -p ~/bin
	cp qinstall ~/bin
	cp qinstall ${GOROOT}/bin

