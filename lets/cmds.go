package lets

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type execJob struct {
	dir  string
	bin  string
	args []string
	out  io.Writer
}

func cmdCopyEnv(cmd *exec.Cmd, k string) {
	v := os.Getenv(k)
	if v == "" {
		return
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
}

func (j *execJob) command() *exec.Cmd {
	cmd := exec.Command(j.bin, j.args...)
	cmd.Dir = j.dir
	if j.out == nil {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = j.out
	}
	cmd.Stderr = os.Stderr
	cmdCopyEnv(cmd, "HOME")
	cmdCopyEnv(cmd, "PATH")
	cmdCopyEnv(cmd, "SSH_AUTH_SOCK")
	return cmd
}

func runCmd(dir, bin string, args ...string) error {
	j := &execJob{
		dir:  dir,
		bin:  bin,
		args: args,
	}
	return j.command().Run()
}

func runCmdOutput(dir, bin string, args ...string) ([]byte, error) {
	j := &execJob{
		dir:  dir,
		bin:  bin,
		args: args,
	}
	cmd := j.command()
	cmd.Stdout = nil
	return cmd.Output()
}

func callCmd(dir, bin string, args ...string) (bool, error) {
	if err := runCmd(dir, bin, args...); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return err.Success(), nil
		}
		return false, err
	}
	return true, nil
}
