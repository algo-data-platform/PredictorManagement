package util

import (
	"os/exec"
	"server/libs/logger"
	"strings"
)

// rsync同步文件到指定目录
func RsyncFile(src string, dest string) error {
	cmd := exec.Command("/bin/rsync", "-r", src, dest)
	logger.Infof("cmd:%v", cmd)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(stdout), "No such file or directory") {
			logger.Debugf("rsync file not exists, file_name=%s", src)
		}
		logger.Errorf("rsync cmd failed: err=%v, output=%s", err, stdout)
	}
	return err
}
