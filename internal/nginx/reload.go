package nginx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/pkg/utils/file"
	"github.com/mitchellh/go-ps"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

const (
	pid        = "/var/run/nginx.pid"
	backupPath = "/etc/nginx/nginx.conf.bak"
	bin        = "/usr/sbin/nginx"
	conf       = "/etc/nginx/nginx.conf"
)

// IsRunning returns true if a process with the name 'nginx' is found
func isRunning() bool {
	processes, err := ps.Processes()
	if err != nil {
		klog.ErrorS(err, "unexpected error obtaining process list")
	}
	for _, p := range processes {
		if p.Executable() == "nginx" {
			return true
		}
	}

	return false
}

func backupNginxConfig(nginxConf string) error {
	if err := overwrite(nginxConf, backupPath); err != nil {
		klog.ErrorS(err, "fail to backup nginx.conf")
		return err
	}
	return nil
}

func checkAndUpdateNginxConfig(nginxConf, testConf string) bool {
	if file.SHA1(nginxConf) == file.SHA1(testConf) {
		klog.Info("the nginx.conf has not been changed, there is no need to reload")
		return false
	}

	if err := backupNginxConfig(nginxConf); err != nil {
		return false
	}

	cmd := exec.Command("nginx", "-c", "-t", testConf)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			klog.ErrorS(err, "the detection of the nginx.conf file did not pass, and it will be rolled back")
			if err := rolloutNginxConfig(backupPath, nginxConf); err != nil {
				return false
			}
			return false
		}
	}

	klog.Info(fmt.Sprintf("%s file test is successful", nginxConf))

	if err := overwrite(testConf, conf); err != nil {
		klog.ErrorS(err, "fail to update nginx.conf")
		return false
	}

	return true
}

func overwrite(src, dst string) error {
	readFile, err := os.ReadFile(src)
	if err == nil {
		if err := os.WriteFile(dst, readFile, 0644); err != nil {
			return err
		}
	}

	return nil
}

func rolloutNginxConfig(src, dst string) error {
	if err := overwrite(src, dst); err != nil {
		klog.ErrorS(err, "fail to rollout nginx.conf")
	}

	return nil
}

func Reload(nginxConf, testConf string) error {
	if !checkAndUpdateNginxConfig(nginxConf, testConf) {
		return nil
	}

	return reload(nginxConf, testConf)
}

func reload(nginxConf, testConf string) error {
	output, err := exec.Command("cat", pid).Output()
	if err != nil {
		return err
	}

	ngxPid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return err
	}

	if err = syscall.Kill(ngxPid, syscall.SIGHUP); err != nil {
		klog.Warningln(err, "failed to reload nginx, rollback in progress")
		if err := rolloutNginxConfig(backupPath, nginxConf); err != nil {
			return err
		}
	}

	if !isRunning() {
		klog.Warningln(err, "nginx process has exited")
		return errors.New("nginx process has exited")
	}

	klog.Info("reload nginx successfully")

	return nil
}

func Start() {
	cmd := exec.Command(bin, "-c", conf)
	var out strings.Builder
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		klog.ErrorS(err, "nginx process has exited")
	}

	if isRunning() {
		klog.Info("start nginx successfully")
	} else {
		klog.Warningln("nginx process has exited")
	}

}
