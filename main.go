package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/xwb1989/sqlparser"

	"github.com/Chyroc/phpmyadmin-cli/internal/common"
	"github.com/Chyroc/phpmyadmin-cli/internal/phpmyadmin"
	"github.com/Chyroc/phpmyadmin-cli/internal/show"
)

var currentDB string
var port int
var host string
var historyPath string
var logPath string
var history []string
var help bool
var prune bool
var list bool
var server string
var username string
var password string
var ErrNotSetServer = errors.New("not set server")

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

func printHelp() {
	fmt.Printf(`NAME:
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
   -h              show help`)
}

func addHistory(word string) {
	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		common.Error(err)
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		common.Error(err)
		return
	}

	bs := strings.Split(string(b), "\n")
	if bs[len(bs)-1] != word {
		_, err := f.WriteString("\n" + word)
		if err != nil {
			common.Error(err)
			return
		}
	}
}

func setServer(sql string) (err error) {
	defer func() {
		if err != nil && err != ErrNotSetServer {
			common.Error(err)
		}
	}()

	sqls := strings.Split(strings.TrimSpace(strings.ToLower(sql)), " ")
	if len(sqls) == 2 && sqls[0] == "set" {
		if sqls[1] == "" {
			return fmt.Errorf("请选择一个server; 输入`show servers`获取所有server; 输入`set <id>`设置server")
		}

		s, err := phpmyadmin.DefaultPHPMyAdmin.GetServerList()
		if err != nil {
			return err
		}

		if s.S == nil {
			return fmt.Errorf("no server found")
		}

		for _, v := range s.S {
			if v.ID == sqls[1] {
				server = v.ID
				setPrefix()
				common.Info("server[%s] seted", v.ID)
				return nil
			}
		}
		return fmt.Errorf("server seted[%s] is invalid", sqls[1])
	}

	if server == "" {
		return fmt.Errorf("请选择一个server; 输入`show servers`获取所有server; 输入`set <id>`设置server")
	}

	return ErrNotSetServer
}

func setDatabase(sql string) (name string, err error) {
	defer func() {
		if err != nil {
			common.Error(err)
		}
	}()

	sqls := strings.Split(strings.TrimSpace(sql), " ")
	if len(sqls) == 2 && strings.ToLower(sqls[0]) == "use" && strings.HasPrefix(sqls[1], "`") && strings.HasSuffix(sqls[1], "`") {
		if len(sqls[1]) <= 2 {
			return "", fmt.Errorf("invalid database name(``)")
		}
		return sqls[1][1 : len(sqls[1])-1], nil
	}

	return sqls[1], nil
}

func setPrefix() {
	promptPrefix := ""
	if server != "" && currentDB != "" {
		promptPrefix = "[" + server + "] " + currentDB + " >>> "
	} else if currentDB == "" {
		promptPrefix = "[" + server + "]" + " >>> "
	}

	LivePrefixState.LivePrefix = promptPrefix
	LivePrefixState.IsEnable = true
}

func execSQL(sql string) {
	sql = strings.TrimSpace(sql)

	if strings.ToLower(sql) == "show servers" {
		s, err := phpmyadmin.DefaultPHPMyAdmin.GetServerList()
		if err != nil {
			common.Error(err)
		}
		s.Print()
		return
	}

	if err := setServer(sql); err != ErrNotSetServer {
		return
	}

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		common.Warn("syntax error: %s %s\n", sql, err)
		return
	}

	result, err := phpmyadmin.DefaultPHPMyAdmin.ExecSQL(server, currentDB, "", sql)
	if err != nil {
		common.Error(err)
		return
	}

	switch stmt.(type) {
	case *sqlparser.Use:
		db, err := setDatabase(sql)
		if err != nil {
			return
		}
		if result == nil {
			common.Warn(`(1049, u"Unknown database '%s'")`+"\n", db)
			return
		}

		currentDB = db
		setPrefix()

		common.Info("Database changed: %s.\n", currentDB)
		return
	default:
		if strings.ToUpper(sql) != "SHOW DATABASES" && currentDB == "" {
			common.Warn("(1046, u'No database selected')\n")
			return
		}

		html := string(result)
		show.ParseFromHTML(fmt.Sprintf("<table>%s</table>", html))
	}
}

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

func executor(in string) {
	common.Logf(in)
	addHistory(in)
	sqls := strings.Split(in, ";")
	for _, s := range sqls {
		if s != "" {
			execSQL(s)
		}
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	var suggest []prompt.Suggest
	for _, v := range common.MySQLKeywords {
		suggest = append(suggest, prompt.Suggest{Text: v})
	}

	suggest = append(suggest, prompt.Suggest{Text: "SERVERS"})

	return prompt.FilterHasPrefix(suggest, in.GetWordBeforeCursor(), true)
}

func initConfig() {
	flag.StringVar(&host, "host", "127.0.0.1", "phpMyAdmin host")
	flag.IntVar(&port, "port", 0, "phpMyAdmin port")
	flag.StringVar(&historyPath, "history", getHomeDir()+"/.phpmyadmin_cli_history", "phpmyadmin history path")
	flag.StringVar(&logPath, "log", getHomeDir()+"/.phpmyadmin_cli.log", "phpmyadmin log path")
	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&prune, "prune", false, "清理命令记录")
	flag.BoolVar(&list, "list", false, "获取server列表")
	flag.StringVar(&server, "server", "1", "选择server")
	flag.BoolVar(&common.IsDebug1, "v", false, "开启调试信息 v")
	flag.BoolVar(&common.IsDebug2, "vv", false, "开启调试信息 vv")
	flag.BoolVar(&common.IsDebug3, "vvv", false, "开启调试信息 vvv")
	flag.StringVar(&username, "username", "", "phpMyAdmin用户名")
	flag.StringVar(&password, "password", "", "phpMyAdmin密码")
	flag.Parse()

	common.InitLog(logPath)

	phpmyadmin.DefaultPHPMyAdmin.SetURI(fmt.Sprintf("%s:%d", host, port))

	if len(flag.Args()) > 0 {
		fmt.Printf("%#v", flag.Args())
		// execSQL("use " + flag.Args()[0])
	}

	body, err := ioutil.ReadFile(historyPath)
	if err != nil {
		panic(err)
	}
	history = strings.Split(string(body), "\n")
}

func reverseStrings(input []string) []string {
	if len(input) == 0 {
		return input
	}
	return append(reverseStrings(input[1:]), input[0])
}

func shortHistory() error {
	body, err := ioutil.ReadFile(historyPath)
	if err != nil {
		return err
	}

	lines := reverseStrings(strings.Split(string(body), "\n"))

	var newLines []string
	var lineSet = make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !lineSet[line] {
			newLines = append(newLines, line)
			lineSet[line] = true
		}
	}

	return ioutil.WriteFile(historyPath, []byte(strings.Join(reverseStrings(newLines), "\n")), 0644)
}

func main() {
	initConfig()

	if help || len(os.Args) == 1 {
		printHelp()
		return
	} else if prune {
		err := shortHistory()
		if err != nil {
			common.Error(err)
		}
		return
	}

	if err := phpmyadmin.DefaultPHPMyAdmin.Login(username, password, server); err != nil {
		common.Exit(err)
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("phpmyadmin cli"),
		prompt.OptionHistory(history),
		prompt.OptionLivePrefix(func() (prefix string, useLivePrefix bool) {
			return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
		}),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x62},
			Fn:        prompt.GoLeftWord,
		}),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x66},
			Fn:        prompt.GoRightWord,
		}),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x08},
			Fn:        prompt.DeleteWord,
		}),
	)
	p.Run()
}
