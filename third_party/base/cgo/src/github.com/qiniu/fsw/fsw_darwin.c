#include "fsw_darwin.h"
#include "_cgo_export.h"

static void my_eventStreamCallback(
	ConstFSEventStreamRef streamRef, void* clientCallBackInfo, size_t numEvents, void* eventPaths,
	const FSEventStreamEventFlags eventFlags[],
	const FSEventStreamEventId eventIds[]) {

	size_t i;
	FSEvent ev;
	char** paths = (char**)eventPaths;

	for (i = 0; i < numEvents; i++) {
		ev.IdxEvent = numEvents - i;
		strncpy(ev.CPath, paths[i], sizeof(ev.CPath));
		ev.EventFlags = eventFlags[i];
		ev.EventId = eventIds[i];
		goEventCallback(clientCallBackInfo, &ev);
	}
}

static int FSWatcher_open_(FSWatcher* w, char* dir, void* clientCallBackInfo) {

	FSEventStreamContext context = { 0, clientCallBackInfo, NULL, NULL, NULL };

	w->runLoop = NULL;
	w->dirNames[0] = CFStringCreateWithCString( NULL, dir, kCFStringEncodingUTF8 );
	w->pathsToWatch = CFArrayCreate( NULL, (const void**)&w->dirNames, 1, NULL );
	w->stream = FSEventStreamCreate( NULL,
		my_eventStreamCallback, &context, w->pathsToWatch,
		kFSEventStreamEventIdSinceNow, (CFAbsoluteTime)1,
		kFSEventStreamCreateFlagNone | kFSEventStreamCreateFlagWatchRoot );
	if (w->stream == NULL) {
		return -1;
	}
	return 0;
}

static int FSWatcher_close_(FSWatcher* w) {

	if (w->stream != NULL) {
		FSEventStreamRelease( w->stream );
		w->stream = NULL;
	}
	if (w->pathsToWatch != NULL) {
		CFRelease(w->pathsToWatch);
		w->pathsToWatch = NULL;
	}
	if (w->dirNames[0] != NULL) {
		CFRelease(w->dirNames[0]);
		w->dirNames[0] = NULL;
	}
	return 0;
}

int FSWatcher_Close(FSWatcher* w) {

	if (w->runLoop != NULL) {
		CFRunLoopStop(w->runLoop);
	}
	return 0;
}

int FSWatcher_Open(FSWatcher* w, char* dir, void* clientCallBackInfo) {

	int err = FSWatcher_open_(w, dir, clientCallBackInfo);
	if (err != 0) {
		FSWatcher_close_(w);
	}
	return err;
}

int FSWatcher_Run(FSWatcher* w) {

	w->runLoop = CFRunLoopGetCurrent();
	FSEventStreamScheduleWithRunLoop( w->stream, w->runLoop, kCFRunLoopDefaultMode );
	FSEventStreamStart( w->stream );
	CFRunLoopRun();
	FSEventStreamStop( w->stream );
	FSEventStreamUnscheduleFromRunLoop( w->stream, w->runLoop, kCFRunLoopDefaultMode );
	FSEventStreamInvalidate( w->stream );
	FSWatcher_close_(w);
	w->runLoop = NULL;
	return 0;
}
