package internal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

var helpInfo = `NAME:
   phpmyadmin-cli - access phpmyadmin from shell cli

USAGE:
   phpmyadmin-cli [global options] [arguments...]

GLOBAL OPTIONS:
   -host           phpMyAdmin host
   -port           phpMyAdmin port
   -server         选择server
   -username       phpMyAdmin用户名（为空则跳过验证）
   -password       phpMyAdmin密码
   -history        command history file (default: "~/.phpmyadmin_cli_history")
   -log            command log file (default: "~/.phpmyadmin_cli.log")
   -v              开启调试信息 v
   -vv             开启调试信息 vv
   -vvv            开启调试信息 vvv

   -list           获取server列表
   -prune          清理命令记录
   -h              show help`

type Cli struct {
	suite.Suite
	t *testing.T
	c *exec.Cmd

	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	command []string
	bin     string

	expectStdout interface{}
	expectStderr interface{}
	expectError  interface{}
}

func (t *Cli) addCommand(s ...string) {
	t.command = append(t.command, s...)
}

func (t *Cli) SetupTest() {
	t.stdout = new(bytes.Buffer)
	t.stderr = new(bytes.Buffer)
	t.command = nil
	t.expectStdout = nil
	t.expectStderr = nil
	t.expectError = nil
}

func (t *Cli) SetupSuite() {
	stdout := new(bytes.Buffer)
	var bin string
	var containers []string

	{
		c := exec.Command("which", "docker")
		c.Stdout = stdout
		t.Nil(c.Run())
		bin = strings.Replace(stdout.String(), "\n", "", -1)
	}

	{
		stdout.Reset()
		c := exec.Command(bin, "ps", "-aq")
		c.Stdout = stdout
		t.Nil(c.Run())
		containers = strings.Split(stdout.String(), "\n")
	}

	{
		for _, v := range containers {
			if v != "" {
				stdout.Reset()
				c := exec.Command(bin, "rm", "-f", v)
				c.Stdout = stdout
				t.Nil(c.Run())
			}
		}
	}

	pwd, err := os.Getwd()
	t.Nil(err)
	t.Nil(exec.Command(fmt.Sprintf("%s/../testdata/start_server_%d.sh", pwd, 1)).Run())
}

func (t *Cli) TestRunCommand() {
	if len(t.command) > 0 {
		var e = []interface{}{"run", strings.Join(t.command, " ")}
		t.c = exec.Command(t.command[0], t.command[1:]...)
		t.c.Stdout = t.stdout
		t.c.Stderr = t.stderr
		if t.expectError == nil {
			t.Nil(t.c.Run(), e...)
		} else {
			err := t.c.Run()
			t.NotNil(err, e...)
			t.Equal(t.expectError.(string), err.Error(), e...)
		}
		if t.expectStdout != nil {
			t.Equal(t.expectStdout.(string), t.stdout.String(), e...)
		}
		if t.expectStderr != nil {
			t.Equal(t.expectStderr.(string), t.stderr.String(), e...)
		}
	}

	t.SetupTest()
}

func (t *Cli) TearDownTest() {
	t.TestRunCommand()
}

func (t *Cli) TestFindBin() {
	pwd, err := os.Getwd()
	t.Nil(err)
	t.bin = fmt.Sprintf("%s/../bin/phpmyadmin-cli", pwd)
	fmt.Printf("bin %s", t.bin)
}

func (t *Cli) TestHelp() {
	{
		t.addCommand(t.bin)
		t.expectStdout = helpInfo
		t.TestRunCommand()
	}

	{
		t.addCommand(t.bin, "-h")
		t.expectStdout = helpInfo
		t.TestRunCommand()
	}
}

func (t *Cli) TestLogin() {
	{
		t.addCommand(t.bin, "-port", "8000")
		t.expectStdout = "need login\nneed login\n"
		t.expectError = "exit status 1"
		t.TestRunCommand()
	}

	{
		t.addCommand(t.bin, "-port", "8000", "-username", "root", "-password", "error")
		t.expectStdout = "login failed\nlogin failed\n"
		t.expectError = "exit status 1"
		t.TestRunCommand()
	}

	{
		t.addCommand(t.bin, "-port", "8000", "-username", "root", "-password", "pass")
		t.expectStdout = "login as [root] success\n\x1b]2;phpmyadmin cli\a"
		t.expectError = "exit status 2"
		t.TestRunCommand()
	}
}

func TestCli(t *testing.T) {
	suite.Run(t, &Cli{
		t: t,
	})
}
