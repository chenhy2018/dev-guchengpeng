all:
	gcc qdelete.c -o qdelete

clean:
	rm -f qdelete

install: all
	@mkdir -p ~/bin
	cp qdelete ~/bin
	cp qdelete ${GOROOT}/bin

