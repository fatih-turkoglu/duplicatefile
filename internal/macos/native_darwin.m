#import <Foundation/Foundation.h>
#include <stdlib.h>
#include <string.h>

char *df_languages(void) {
    @autoreleasepool {
        NSData *data = [NSJSONSerialization dataWithJSONObject:[NSLocale preferredLanguages]
                                                     options:0 error:nil];
        return data ? strndup(data.bytes, data.length) : NULL;
    }
}

int df_dark_mode(void) {
    @autoreleasepool {
        NSString *style = [[NSUserDefaults standardUserDefaults] stringForKey:@"AppleInterfaceStyle"];
        return style && [style caseInsensitiveCompare:@"Dark"] == NSOrderedSame;
    }
}

int df_trash(const char *path, char **message) {
    @autoreleasepool {
        NSString *name = [[NSFileManager defaultManager] stringWithFileSystemRepresentation:path
                                                                                  length:strlen(path)];
        NSError *error = nil;
        if (name && [[NSFileManager defaultManager] trashItemAtURL:[NSURL fileURLWithPath:name]
                                                resultingItemURL:nil error:&error]) {
            return 0;
        }
        *message = strdup(error.localizedDescription.UTF8String ?: "Invalid file path");
        if ([error.domain isEqualToString:NSCocoaErrorDomain]) {
            switch (error.code) {
                case NSFeatureUnsupportedError: return 1;
                case NSFileReadNoPermissionError:
                case NSFileWriteNoPermissionError: return 2;
                case NSFileNoSuchFileError:
                case NSFileReadNoSuchFileError: return 3;
            }
        }
        return 4;
    }
}
