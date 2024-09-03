package fsstats

import (
    "os"
    "context"
    "time"
    "fmt"
    "github.com/moby/sys/mountinfo"
    "os/exec"
)

func Stat(filePath string) (os.FileInfo, error) {
    // Defaulting to 3 secs for now.
    // TODO : Get this value as input param to kubelet (FS_STAT_TIMEOUT).
    if !FileSystemHung(filePath, 3) {
        // Redundant stat call here else we get into problems of ipc and output parsing.
        return os.Stat(filePath)
    }
    return nil, fmt.Errorf("File System Hung.")
}

func FileExists(file string) bool {
    if _, err := Stat(file); err != nil {
        return false
    }
    return true
}

// This is needed since cAdvisor uses moby's package to check for mountedFast.
func MountedFast(path string) (mounted, sure bool, err error) {
    if !FileSystemHung(path, 5) {
        isMnt, sure, isMntErr := mountinfo.MountedFast(path)
        return isMnt, sure, isMntErr
    }
       return false, false, fmt.Errorf("Mount Path is hung ", path)
}

func FileSystemHung(filePath string, timeout int) bool {
    // Fork a new process and check if stat for filepath is hung or not.
    // kill the fork if timeout.
    ctx := context.Background()
    if timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
        defer cancel()
    }
    cmd := exec.CommandContext(ctx, "stat", filePath)
    err := cmd.Run()
    if err != nil && ctx.Err() == context.DeadlineExceeded {
        return true
    }
    return false
}
