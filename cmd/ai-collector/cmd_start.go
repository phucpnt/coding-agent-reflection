package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"

	cfg "github.com/phuc/coding-agent-reflection/internal/config"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the collector in the background",
	RunE:  runStart,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background collector",
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	pidFile := cfg.PidPath()
	if pid := readPid(pidFile); pid > 0 && processRunning(pid) {
		fmt.Printf("Collector already running (PID %d)\n", pid)
		return nil
	}

	os.MkdirAll(cfg.DataDir(), 0o755)

	logFile, err := os.OpenFile(cfg.LogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		exe = "ai-collector"
	}

	proc := exec.Command(exe, "serve")
	proc.Stdout = logFile
	proc.Stderr = logFile
	proc.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := proc.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("start collector: %w", err)
	}
	logFile.Close()

	pid := proc.Process.Pid
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}

	fmt.Printf("Collector started (PID %d)\n", pid)
	fmt.Printf("Logs: %s\n", cfg.LogPath())
	return nil
}

func runStop(cmd *cobra.Command, args []string) error {
	pidFile := cfg.PidPath()
	pid := readPid(pidFile)
	if pid <= 0 {
		fmt.Println("No collector running (no PID file)")
		return nil
	}

	if !processRunning(pid) {
		os.Remove(pidFile)
		fmt.Println("No collector running (stale PID file removed)")
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM: %w", err)
	}

	os.Remove(pidFile)
	fmt.Printf("Collector stopped (PID %d)\n", pid)
	return nil
}

func readPid(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0
	}
	return pid
}

func processRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
