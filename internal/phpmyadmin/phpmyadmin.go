package phpmyadmin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/Chyroc/phpmyadmin-cli/internal/common"
	"github.com/Chyroc/phpmyadmin-cli/internal/requests"
	"github.com/Chyroc/phpmyadmin-cli/internal/utils"
)

var DefaultPHPMyAdmin *phpMyAdmin
var tokenRegexp = regexp.MustCompile("<input type=\"hidden\" name=\"token\" value=\"(.*?)\" [/]>")

type phpMyAdmin struct {
	session *requests.Session
	Token   string
	uri     string
}

type phpMyAdminResp struct {
	Message string
	Success bool
	Error   string
}

type Server struct {
	ID   string
	Name string
}

type Servers struct {
	S []Server
}

func (s *Servers) Print() {
	if len(s.S) == 0 {
		common.Info("no servers found")
		return
	}

	for _, v := range s.S {
		common.Info(fmt.Sprintf("%s: %s\n", v.ID, v.Name))
	}
}
func init() {
	DefaultPHPMyAdmin = &phpMyAdmin{
		session: requests.DefaultSession,
	}
}

func (p *phpMyAdmin) SetURI(uri string) {
	p.uri = uri
}

func (p *phpMyAdmin) Login(username, password, server string) (err error) {
	defer func() {
		if err != nil {
			common.Error(err)
		}
	}()

	if _, err = p.Get(p.uri, "index.php", nil); err != nil {
		// username == "", but check login err
		// username != "" and common.ErrNeedLogin, try to login
		if username == "" || (username != "" && err != common.ErrNeedLogin) {
			return err
		}
	}
	if username == "" {
		return nil
	}

	body := fmt.Sprintf("pma_username=%s&pma_password=%s&token=%s&server=%s", username, password, utils.Escape(p.Token), server)
	header := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	if _, err := p.Post(p.uri, "index.php", nil, header, strings.NewReader(body)); err != nil {
		if err == common.ErrNeedLogin {
			return common.ErrLoginFailed
		}
		return err
	}

	result, err := p.Get(p.uri, "server_status_processes.php", nil)
	if err != nil {
		return err
	}

	if !strings.Contains(string(result), "SHOW PROCESSLIST") {
		return common.ErrLoginFailed
	}

	common.Info("login as [%s] success\n", username)

	return nil
}

func (p *phpMyAdmin) GetServerList() (*Servers, error) {
	b, err := p.Get(p.uri, "", nil)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	var s []Server
	doc.Find("#select_server > option").Each(func(_ int, selection *goquery.Selection) {
		id := strings.TrimSpace(selection.AttrOr("value", ""))
		name := strings.TrimSpace(selection.Text())

		if id != "" {
			s = append(s, Server{id, name})
		}
	})

	return &Servers{s}, nil
}

func (p *phpMyAdmin) ExecSQL(server, database, table, sql string) ([]byte, error) {
	data := map[string]string{
		// "table":             table,
		// "prev_sql_query":    "",
		"db":                   database,
		"server":               server,
		"token":                p.Token,
		"sql_query":            sql,
		"ajax_request":         "true",
		"ajax_page_request":    "true",
		"display_options_form": "1",
		"pftext":               "F",
		"hide_transformation":  "on",
		"session_max_rows":     "all",
	}
	var bs []string
	for k, v := range data {
		bs = append(bs, k+"="+utils.Escape(v))
	}
	body := strings.NewReader(strings.Join(bs, "&"))
	header := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	b, err := p.Post(p.uri+"/import.php", "", nil, header, body)
	if err != nil {
		return nil, err
	}

	var r phpMyAdminResp
	if err = json.Unmarshal(b, &r); err != nil {
		return nil, err
	}

	return handlerPhpmyadminResp(r)
}
