package nginx

import (
	"errors"
	"fmt"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/config"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/pkg/utils/file"
	"github.com/mitchellh/go-ps"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func backupConf(src, dstTest, dstBak string) error {
	defer CleanConf(dstTest)

	// backup
	if err := generateConf(src, dstBak); err != nil {
		return err
	}

	// overwrite produce file
	if err := generateConf(dstTest, src); err != nil {
		return err
	}

	return nil
}

func rolloutConf(src, dst string) error {
	if err := generateConf(src, dst); err != nil {
		return err
	}
	return nil
}

func generateConf(src, dst string) error {
	readFile, err := os.ReadFile(src)
	if err == nil {
		if err := os.WriteFile(dst, readFile, 0644); err != nil {
			return err
		}

		stat, err := os.Stat(dst)
		if err != nil || stat.Size() == 0 || file.SHA1(src) != file.SHA1(dst) {
			return fmt.Errorf("failed to generate nginx configuration file and cannot proceed to the next step, file: %s", dst)
		}
	}

	return nil
}

func CleanConf(files ...string) {
	for _, v := range files {
		if _, err := os.Stat(v); err == nil {
			if err := os.Remove(v); err != nil {
				klog.ErrorS(err, fmt.Sprintf("fail to clear %s", v))
			}
		}
	}

}

func Reload(name string) error {
	return reload(name)
}

func reload(name string) error {
	var isFirstReload bool
	var exitError *exec.ExitError

	productConf := name + ".conf"
	testConf := name + "-test.conf"
	backupFile := name + ".tmp"

	defer CleanConf(backupFile, testConf)

	if _, err := os.Stat(productConf); err == nil {
		if file.SHA1(productConf) == file.SHA1(testConf) {
			klog.Info(fmt.Sprintf("%s has not changed, no need to reload nginx", productConf))
			return err
		}

		if err := backupConf(productConf, testConf, backupFile); err != nil {
			return err
		}
	} else {
		isFirstReload = true
	}

	cmd := exec.Command(config.Bin, "-t")
	if err := cmd.Run(); err != nil {
		if errors.As(err, &exitError) {
			klog.ErrorS(err, fmt.Sprintf("nginx configuration: %s file verification fails, pls check", productConf))
			if !isFirstReload {
				if err := rolloutConf(backupFile, productConf); err != nil {
					return err
				}
			}
			return err
		}
	}

	if err := generateConf(testConf, productConf); err != nil {
		return err
	}

	if err := gracefulRestart(); err != nil {
		return err
	}

	return nil
}

func reloadIfWatchFileCurd() {
	var exitError *exec.ExitError
	cmd := exec.Command(config.Bin, "-t")
	if err := cmd.Run(); err != nil {
		if errors.As(err, &exitError) {
			klog.ErrorS(err, "failed to successfully reload nginx upon detecting file changes")
			return
		}
	}

	if err := gracefulRestart(); err != nil {
		return
	}
}

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

func gracefulRestart() error {
	output, err := exec.Command("cat", config.Pid).Output()
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("the pid of nginx cannot be found"))
		return err
	}

	ngxPid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return err
	}

	if err = syscall.Kill(ngxPid, syscall.SIGHUP); err != nil {
		klog.ErrorS(err, "failed to reload nginx")
		return err
	}

	if !isRunning() {
		klog.Fatalln("Fatal error nginx process does not exist")
	}

	return nil
}

func Start() {
	klog.Info("start nginx")
	var done = make(chan struct{})
	go func() {
		stopSingle := time.NewTimer(time.Duration(10) * time.Second)
		defer stopSingle.Stop()

		for {
			select {
			case <-stopSingle.C:
				klog.Error("timeout waiting for nginx to start")
				return
			case <-done:
				dirWatcher()
				return
			default:
				if isRunning() {
					close(done)
				}
			}
		}
	}()

	cmd := exec.Command(config.Bin, "-c", config.MainConf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		klog.Fatalln(err, "Failed to start nginx")
	}
}

func dirWatcher() {
	klog.Info("start watcher")
	if _, err := file.NewFileWatcher(config.SslPath, reloadIfWatchFileCurd); err != nil {
		klog.Fatal(fmt.Sprintf("fail to watch %s, error %v", config.SslPath, err))
	}

	if _, err := file.NewFileWatcher(config.ConfDir, reloadIfWatchFileCurd); err != nil {
		klog.Fatal(fmt.Sprintf("fail to watch %s, error %v", config.ConfDir, err))
	}
}
