package cmd

import (
	b64 "encoding/base64"
	"fmt"

	"os"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"

)

func init() {
	log.SetHandler(cli.Default)
}

func NormalizePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(UserHomeDir(), path[2:])
	}
	return path
}
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}
	return os.Getenv("HOME")
}

func (e *RetriableError) Error() string {
	return fmt.Sprintf("%s (retry after %v)", e.Err.Error())
}

func decodeB64(text string) string {
	b, _ := b64.StdEncoding.DecodeString(text)
	return string(b)

}
func encodeB64(text string) string {
	b := b64.StdEncoding.EncodeToString([]byte(text))
	return string(b)
}
func removeFile(file string) {
	ctx := log.WithFields(log.Fields{
		"file": file,
	})
	if err := os.Remove(file); err != nil {

		ctx.Fatalf("%s", err)

	} else {
		ctx.Info("File removed")

	}

}
func Contains[T comparable](arr []T, x T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}