package copy
import (
    "os/exec"
)

func CopyToClipboard(text string) error {
    cmd := exec.Command("xclip", "-selection", "clipboard")
    in, _ := cmd.StdinPipe()
    go func() {
        defer in.Close()
        in.Write([]byte(text))
    }()
    return cmd.Run()
}

func NotifyUser(msg string) {
    exec.Command("notify-send", "LatexClipboard", msg).Run()
}
