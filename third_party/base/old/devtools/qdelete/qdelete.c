#include <stdlib.h>
#include <dirent.h>
#include <string.h>

int main(int argc, const char* argv[]) {

	int i;
	char path[1024];
	for (i = 3; i < argc; i++) {
		strcpy(path, argv[1]);
		strcat(path, argv[i]);
		strcat(path, argv[2]);
		remove(path);
	}
	return 0;
}

