package phpmyadmin

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Chyroc/phpmyadmin-cli/internal/common"
	"github.com/Chyroc/phpmyadmin-cli/internal/show"
)

func TestStart2(t *testing.T) {
	startServer(t, true)
}

func TestLogin2(t *testing.T) {
	as := assert.New(t)
	p := initClient()
	p.SetURI("localhost:8000")

	as.Nil(p.Login("root", "pass", "1"))
	as.Nil(p.Login("root", "pass", "2"))

	err := p.Login("root", "error", "1")
	as.NotNil(err)
	as.Equal(common.ErrLoginFailed, err)
}

func TestExecSQL2(t *testing.T) {
	p := initClient()
	as := assert.New(t)
	var buf = new(bytes.Buffer)
	show.TestSetOut(buf)
	p.SetURI("localhost:8000")

	as.Nil(p.Login("root", "pass", "1"))
	as.Nil(p.Login("root", "pass", "2"))

	b, err := p.ExecSQL("2", "", "", "CREATE DATABASE IF NOT EXISTS `test_phpmyadmin_cli`;")
	as.Nil(err)
	as.NotNil(b)

	b, err = p.ExecSQL("2", "test_phpmyadmin_cli", "", create_table)
	as.Nil(err)
	as.NotNil(b)

	b, err = p.ExecSQL("2", "test_phpmyadmin_cli", "", "show create table name")
	as.Nil(err)
	as.NotNil(b)
	buf.Reset()
	show.ParseFromHTML(string(b))
	as.Equal(show_create_table_name, buf.String())

	b, err = p.ExecSQL("2", "test_phpmyadmin_cli", "", "truncate name")
	as.Nil(err)
	as.NotNil(b)

	for i := 0; i < 100; i++ {
		b, err = p.ExecSQL("2", "test_phpmyadmin_cli", "name", fmt.Sprintf("insert into name (id, name) values ('id_%d', 'name_%d');", i, i))
		as.Nil(err)
		as.NotNil(b)
	}

	b, err = p.ExecSQL("2", "test_phpmyadmin_cli", "name", "select id, name from name;")
	as.Nil(err)
	as.NotNil(b)
	buf.Reset()
	show.ParseFromHTML(string(b))
	as.Equal(id_name, buf.String())
}

func TestShowDatabases2(t *testing.T) {
	p := initClient()
	as := assert.New(t)
	p.SetURI("localhost:8000")

	as.Nil(p.Login("root", "pass", "1"))
	as.Nil(p.Login("root", "pass", "2"))

	var buf = new(bytes.Buffer)

	b, err := p.ExecSQL("2", "", "", "show databases;")
	as.Nil(err)
	as.NotNil(b)
	show.TestSetOut(buf)
	show.ParseFromHTML(string(b))
	as.Equal(`+---------------------+
|      Database       |
+---------------------+
| information_schema  |
| mysql               |
| performance_schema  |
| sys                 |
| test_phpmyadmin_cli |
+---------------------+
`, buf.String())

	b, err = p.ExecSQL("1", "", "", "show databases;")
	as.Nil(err)
	as.NotNil(b)
	buf.Reset()
	show.TestSetOut(buf)
	show.ParseFromHTML(string(b))
	as.Equal(`+--------------------+
|      Database      |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
`, buf.String())
}

func TestShowTables2(t *testing.T) {
	p := initClient()
	as := assert.New(t)
	p.SetURI("localhost:8000")

	as.Nil(p.Login("root", "pass", "1"))

	b, err := p.ExecSQL("root", "information_schema", "", "show tables;")
	as.Nil(err)
	as.NotNil(b)

	var buf = new(bytes.Buffer)
	show.TestSetOut(buf)
	show.ParseFromHTML(string(b))
	as.Equal(information_schema_tables, buf.String())
}
