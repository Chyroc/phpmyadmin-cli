package phpmyadmin

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"encoding/json"

	"github.com/Chyroc/phpmyadmin-cli/internal/requests"
)

type phpmyadmin struct {
	*requests.Session
	Token string
	uri   string
}

type PhpmyadminResp struct {
	Message string
	Success bool
	Error   string
}

var DefaultPhpmyadmin *phpmyadmin

func init() {
	DefaultPhpmyadmin = &phpmyadmin{
		Session: requests.DefaultSession,
	}
}

var tokenRegexp = regexp.MustCompile("<input type=\"hidden\" name=\"token\" value=\"(.*?)\" >")

func (p *phpmyadmin) SetURI(uri string) {
	p.uri = uri
}

func (p *phpmyadmin) initCookie() error {
	resp, err := requests.DefaultSession.Get(p.uri+"/index.php", "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matchToken := tokenRegexp.FindStringSubmatch(string(b))
	if len(matchToken) != 2 {
		return fmt.Errorf("match token err: %s", strings.Join(matchToken, ";"))
	} else if matchToken[1] == "" {
		return fmt.Errorf("empty token")
	}

	p.Token = matchToken[1]

	return nil
}

func (p *phpmyadmin) GetDatabases(server string) error {
	if p.Token == "" {
		p.initCookie()
	}

	body := strings.NewReader(fmt.Sprintf(`token=%s&server=%s`, p.Token, server))
	header := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	resp, err := requests.DefaultSession.Post(p.uri+"/index.php", "", nil, header, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	databases, err := docDatabases(resp)
	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", databases)

	return nil
}

func (p *phpmyadmin) GetTables(server, database string) error {
	if p.Token == "" {
		p.initCookie()
	}

	resp, err := requests.DefaultSession.Get(fmt.Sprintf("%s/db_structure.php?server=%s&db=%s&ajax_request=true&ajax_page_request=true", p.uri, server, database), "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r PhpmyadminResp
	if err = json.Unmarshal(b, &r); err != nil {
		return err
	}

	tables, err := docTables(r.Message)
	if err != nil {
		return err
	}

	fmt.Printf("tables %#v\n", tables)

	return nil
}