#include <stdlib.h>
#include <dirent.h>
#include <limits.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <string>
#include <vector>
#include <sys/stat.h>

// bmake { TargetDirs } MakeTargets

bool fileExists(const char* file) {
	FILE* fp = fopen(file, "rb");
	if (fp == NULL)
		return false;
	fclose(fp);
	return true;
}

void make(const char* dir, const char* cmd) {
	char cwd[PATH_MAX];
	getcwd(cwd, PATH_MAX);
	if (chdir(dir) != 0) {
		printf("\n===> Make %s failed: not found\n", dir);
		return;
	}
	if (fileExists("Makefile")) {
		printf("\n===> Make %s ...\n", dir);
		system(cmd);
	}
	chdir(cwd);
}

void makeAllDirs(const char* cmd) {
	DIR* p = opendir(".");
	if (p == NULL)
		return;
	for (;;) {
		struct dirent* e = readdir(p);
		if (e == NULL)
			break;
#if defined(__MINGW32_VERSION)
		if (*e->d_name != '.') {
			struct stat fi;
			fi.st_mode = 0;
			stat(e->d_name, &fi);
			if (_S_ISDIR(fi.st_mode)) {
				make(e->d_name, cmd);
			}
		}
#else
		if ((e->d_type & DT_DIR) && *e->d_name != '.') {
			make(e->d_name, cmd);
		}
#endif
	}
	closedir(p);
}

typedef std::vector<const char*> TargetDirs;

void makeTargetDirs(const TargetDirs& dirs, const char* cmd) {
	for (size_t i = 0; i < dirs.size(); ++i) {
		make(dirs[i], cmd);
	}
}

int main(int argc, const char* argv[]) {
	if (argc < 2) {
		makeAllDirs("make");
		return 0;
	}
	
	int i = 1;
	TargetDirs dirs;
	if (strcmp(argv[1], "{") == 0) {
		for (i = 2; i < argc; ++i) {
			if (strcmp(argv[i], "}") == 0) {
				++i;
				break;
			}
			dirs.push_back(argv[i]);
		}
	}
	
	std::string cmd = "make";
	for (; i < argc; ++i) {
		cmd.append(1, ' ');
		cmd.append(argv[i]);
	}
	const char* cmd1 = cmd.c_str();

	if (dirs.empty()) {
		makeAllDirs(cmd1);
	} else {
		makeTargetDirs(dirs, cmd1);
	}
	return 0;
}

