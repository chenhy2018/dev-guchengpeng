#include <stdlib.h>
#include <stdio.h>
#include <dirent.h>
#include <string.h>

#ifndef PATH_MAX
#define PATH_MAX	1024
#endif

#if defined(_WIN32) || defined(_WIN64)

int main(int argc, const char* argv[]) {
	return -1;
}

#else

const char* getpkg(char* buf, const char* goroot) {

	strcpy(buf, goroot);
	strcat(buf, "/pkg");

	DIR* p = opendir(buf);
	if (p == NULL) {
		*buf = '\0';
		return NULL;
	}
	for (;;) {
		struct dirent* e = readdir(p);
		if (e == NULL) {
			*buf = '\0';
			buf = NULL;
			break;
		}
		if ((e->d_type & DT_DIR) && *e->d_name != '_' && strchr(e->d_name, '_') != NULL) {
			strcat(buf, "/");
			strcat(buf, e->d_name);
			break;
		}
	}
	closedir(p);
	return buf;
}

int dirExists(const char* dir) {
	DIR* fp = opendir(dir);
	if (fp == NULL)
		return 0;
	closedir(fp);
	return 1;
}

int main(int argc, const char* argv[]) {

	char* p;
	char* goroot = getenv("GOROOT");

	char cwd[PATH_MAX], pkg[PATH_MAX];
	getcwd(cwd, PATH_MAX);
	getpkg(pkg, goroot);
	setenv("GOPKG_OSPATH", pkg, 1);

	if (dirExists("src")) {
		setenv("GOPATH", cwd, 1);
		system("go list ./... | xargs qdelete $GOPKG_OSPATH/ .a");
		system("go install -v ./...");
		system("cp -f -r pkg/* $GOROOT/pkg/");
	} else {
		p = strstr(cwd, "/src/");
		if (p != NULL) {
			*p = '\0';
			setenv("GOPATH", cwd, 1);
			system("go list | xargs qdelete $GOPKG_OSPATH/ .a");
			system("go install -v");
			system("cp -f -r $GOPATH/pkg/* $GOROOT/pkg/");
		}
	}
	return 0;
}

#endif
