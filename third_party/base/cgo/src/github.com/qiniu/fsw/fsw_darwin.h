#ifndef QINIU_FSW_DARWIN_H
#define QINIU_FSW_DARWIN_H

#include <CoreServices/CoreServices.h>

// -------------------------------------------------------------------------

typedef struct tagFSEvent {
	int IdxEvent; // numEvents .. 1
	int EventFlags;
	FSEventStreamEventId EventId;
	char CPath[1024];
} FSEvent;

typedef struct tagFSWatcher {
	CFRunLoopRef runLoop;
	CFStringRef dirNames[1];
	CFArrayRef pathsToWatch;
	FSEventStreamRef stream;
} FSWatcher;

int FSWatcher_Close(FSWatcher* w);
int FSWatcher_Open(FSWatcher* w, char* dir, void* param);
int FSWatcher_Run(FSWatcher* w);

// -------------------------------------------------------------------------

#endif
