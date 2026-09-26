#import <Cocoa/Cocoa.h>
#import <CoreVideo/CoreVideo.h>
#include <errno.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct FamFrameClock {
    CVDisplayLinkRef link;
    dispatch_semaphore_t tick;
    NSWindow *window;
    CGDirectDisplayID display;
} FamFrameClock;

static CGDirectDisplayID windowDisplay(NSWindow *window) {
    @autoreleasepool {
        NSScreen *screen = [window screen];
        return [[[screen deviceDescription] objectForKey:@"NSScreenNumber"] unsignedIntValue];
    }
}

static CVReturn displayTick(CVDisplayLinkRef link, const CVTimeStamp *now,
                           const CVTimeStamp *output, CVOptionFlags flags,
                           CVOptionFlags *outFlags, void *context) {
    FamFrameClock *clock = context;
    dispatch_semaphore_signal(clock->tick);
    return kCVReturnSuccess;
}

FamFrameClock *famFrameClockCreate(uintptr_t window, int *status) {
    FamFrameClock *clock = calloc(1, sizeof(*clock));
    if (!clock) {
        *status = ENOMEM;
        return NULL;
    }
    clock->window = (NSWindow *)window;
    clock->display = windowDisplay(clock->window);
    clock->tick = dispatch_semaphore_create(0);
    if (!clock->tick) {
        *status = ENOMEM;
        free(clock);
        return NULL;
    }
    *status = CVDisplayLinkCreateWithCGDisplay(clock->display, &clock->link);
    if (*status == kCVReturnSuccess) {
        *status = CVDisplayLinkSetOutputCallback(clock->link, displayTick, clock);
    }
    if (*status == kCVReturnSuccess) {
        *status = CVDisplayLinkStart(clock->link);
    }
    if (*status != kCVReturnSuccess) {
        if (clock->link) {
            CVDisplayLinkRelease(clock->link);
        }
        dispatch_release(clock->tick);
        free(clock);
        return NULL;
    }
    return clock;
}

int famFrameClockWait(FamFrameClock *clock) {
    CGDirectDisplayID display = windowDisplay(clock->window);
    if (display && display != clock->display) {
        CVReturn status = CVDisplayLinkSetCurrentCGDisplay(clock->link, display);
        if (status != kCVReturnSuccess) {
            return status;
        }
        clock->display = display;
    }
    // Discard old ticks rather than rendering catch-up frames after a stall.
    while (dispatch_semaphore_wait(clock->tick, DISPATCH_TIME_NOW) == 0) {}
    // A sleeping display may stop ticking. Let the main loop keep polling events.
    return dispatch_semaphore_wait(clock->tick,
                                  dispatch_time(DISPATCH_TIME_NOW, 100 * NSEC_PER_MSEC)) == 0;
}

int famFrameClockDestroy(FamFrameClock *clock) {
    if (CVDisplayLinkIsRunning(clock->link)) {
        CVReturn status = CVDisplayLinkStop(clock->link);
        if (status != kCVReturnSuccess) {
            return status;
        }
    }
    CVDisplayLinkRelease(clock->link);
    dispatch_release(clock->tick);
    free(clock);
    return kCVReturnSuccess;
}
