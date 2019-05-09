package phpmyadmin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Chyroc/phpmyadmin-cli/internal/requests"
)

var information_schema_tables = `+---------------------------------------+
|     Tables_in_information_schema      |
+---------------------------------------+
| CHARACTER_SETS                        |
| COLLATIONS                            |
| COLLATION_CHARACTER_SET_APPLICABILITY |
| COLUMNS                               |
| COLUMN_PRIVILEGES                     |
| ENGINES                               |
| EVENTS                                |
| FILES                                 |
| GLOBAL_STATUS                         |
| GLOBAL_VARIABLES                      |
| KEY_COLUMN_USAGE                      |
| OPTIMIZER_TRACE                       |
| PARAMETERS                            |
| PARTITIONS                            |
| PLUGINS                               |
| PROCESSLIST                           |
| PROFILING                             |
| REFERENTIAL_CONSTRAINTS               |
| ROUTINES                              |
| SCHEMATA                              |
| SCHEMA_PRIVILEGES                     |
| SESSION_STATUS                        |
| SESSION_VARIABLES                     |
| STATISTICS                            |
| TABLES                                |
| TABLESPACES                           |
| TABLE_CONSTRAINTS                     |
| TABLE_PRIVILEGES                      |
| TRIGGERS                              |
| USER_PRIVILEGES                       |
| VIEWS                                 |
| INNODB_LOCKS                          |
| INNODB_TRX                            |
| INNODB_SYS_DATAFILES                  |
| INNODB_FT_CONFIG                      |
| INNODB_SYS_VIRTUAL                    |
| INNODB_CMP                            |
| INNODB_FT_BEING_DELETED               |
| INNODB_CMP_RESET                      |
| INNODB_CMP_PER_INDEX                  |
| INNODB_CMPMEM_RESET                   |
| INNODB_FT_DELETED                     |
| INNODB_BUFFER_PAGE_LRU                |
| INNODB_LOCK_WAITS                     |
| INNODB_TEMP_TABLE_INFO                |
| INNODB_SYS_INDEXES                    |
| INNODB_SYS_TABLES                     |
| INNODB_SYS_FIELDS                     |
| INNODB_CMP_PER_INDEX_RESET            |
| INNODB_BUFFER_PAGE                    |
| INNODB_FT_DEFAULT_STOPWORD            |
| INNODB_FT_INDEX_TABLE                 |
| INNODB_FT_INDEX_CACHE                 |
| INNODB_SYS_TABLESPACES                |
| INNODB_METRICS                        |
| INNODB_SYS_FOREIGN_COLS               |
| INNODB_CMPMEM                         |
| INNODB_BUFFER_POOL_STATS              |
| INNODB_SYS_COLUMNS                    |
| INNODB_SYS_FOREIGN                    |
| INNODB_SYS_TABLESTATS                 |
+---------------------------------------+
`
var id_name = `+-------+---------+
|  id   |  name   |
+-------+---------+
| id_0  | name_0  |
| id_1  | name_1  |
| id_2  | name_2  |
| id_3  | name_3  |
| id_4  | name_4  |
| id_5  | name_5  |
| id_6  | name_6  |
| id_7  | name_7  |
| id_8  | name_8  |
| id_9  | name_9  |
| id_10 | name_10 |
| id_11 | name_11 |
| id_12 | name_12 |
| id_13 | name_13 |
| id_14 | name_14 |
| id_15 | name_15 |
| id_16 | name_16 |
| id_17 | name_17 |
| id_18 | name_18 |
| id_19 | name_19 |
| id_20 | name_20 |
| id_21 | name_21 |
| id_22 | name_22 |
| id_23 | name_23 |
| id_24 | name_24 |
| id_25 | name_25 |
| id_26 | name_26 |
| id_27 | name_27 |
| id_28 | name_28 |
| id_29 | name_29 |
| id_30 | name_30 |
| id_31 | name_31 |
| id_32 | name_32 |
| id_33 | name_33 |
| id_34 | name_34 |
| id_35 | name_35 |
| id_36 | name_36 |
| id_37 | name_37 |
| id_38 | name_38 |
| id_39 | name_39 |
| id_40 | name_40 |
| id_41 | name_41 |
| id_42 | name_42 |
| id_43 | name_43 |
| id_44 | name_44 |
| id_45 | name_45 |
| id_46 | name_46 |
| id_47 | name_47 |
| id_48 | name_48 |
| id_49 | name_49 |
| id_50 | name_50 |
| id_51 | name_51 |
| id_52 | name_52 |
| id_53 | name_53 |
| id_54 | name_54 |
| id_55 | name_55 |
| id_56 | name_56 |
| id_57 | name_57 |
| id_58 | name_58 |
| id_59 | name_59 |
| id_60 | name_60 |
| id_61 | name_61 |
| id_62 | name_62 |
| id_63 | name_63 |
| id_64 | name_64 |
| id_65 | name_65 |
| id_66 | name_66 |
| id_67 | name_67 |
| id_68 | name_68 |
| id_69 | name_69 |
| id_70 | name_70 |
| id_71 | name_71 |
| id_72 | name_72 |
| id_73 | name_73 |
| id_74 | name_74 |
| id_75 | name_75 |
| id_76 | name_76 |
| id_77 | name_77 |
| id_78 | name_78 |
| id_79 | name_79 |
| id_80 | name_80 |
| id_81 | name_81 |
| id_82 | name_82 |
| id_83 | name_83 |
| id_84 | name_84 |
| id_85 | name_85 |
| id_86 | name_86 |
| id_87 | name_87 |
| id_88 | name_88 |
| id_89 | name_89 |
| id_90 | name_90 |
| id_91 | name_91 |
| id_92 | name_92 |
| id_93 | name_93 |
| id_94 | name_94 |
| id_95 | name_95 |
| id_96 | name_96 |
| id_97 | name_97 |
| id_98 | name_98 |
| id_99 | name_99 |
+-------+---------+
`
var create_table = `
CREATE TABLE IF NOT EXISTS ` + "`" + `name` + "`" + ` (
  ` + "`" + `id` + "`" + ` varchar(32) NOT NULL COMMENT "id",
  ` + "`" + `name` + "`" + ` varchar(255) NOT NULL COMMENT "名称",
  ` + "`" + `created_at` + "`" + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT "创建时间",
  ` + "`" + `updated_at` + "`" + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT "更新时间"
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;`

var show_create_table_name = `+-------+------------------------------------------------------------------------------------------------------------+
| Table |                                                Create Table                                                |
+-------+------------------------------------------------------------------------------------------------------------+
| name  | CREATE TABLE ` + "`" + `name` + "`" + ` (                                                                                      |
|       |   ` + "`" + `id` + "`" + ` varchar(32) NOT NULL COMMENT 'id',                                                                  |
|       |   ` + "`" + `name` + "`" + ` varchar(255) NOT NULL COMMENT '名称',                                                             |
|       |   ` + "`" + `created_at` + "`" + ` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',                            |
|       |   ` + "`" + `updated_at` + "`" + ` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间' |
|       | ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4                                                                    |
+-------+------------------------------------------------------------------------------------------------------------+
`

func initClient() *phpMyAdmin {
	s, err := requests.NewSession()
	if err != nil {
		panic(err)
	}
	return &phpMyAdmin{
		session: s,
	}
}

func startServer(t *testing.T, multi bool) {
	as := assert.New(t)
	stdout := new(bytes.Buffer)
	var bin string
	var containers []string

	{
		c := exec.Command("which", "docker")
		c.Stdout = stdout
		as.Nil(c.Run())
		bin = strings.Replace(stdout.String(), "\n", "", -1)
	}

	{
		stdout.Reset()
		c := exec.Command(bin, "ps", "-aq")
		c.Stdout = stdout
		as.Nil(c.Run())
		containers = strings.Split(stdout.String(), "\n")
	}

	{
		for _, v := range containers {
			if v != "" {
				stdout.Reset()
				c := exec.Command(bin, "rm", "-f", v)
				c.Stdout = stdout
				as.Nil(c.Run())
			}
		}
	}

	pwd, err := os.Getwd()
	as.Nil(err)
	if multi {
		as.Nil(exec.Command(fmt.Sprintf("%s/../../testdata/start_server_%d.sh", pwd, 2)).Run())
	} else {
		as.Nil(exec.Command(fmt.Sprintf("%s/../../testdata/start_server_%d.sh", pwd, 1)).Run())
	}
}

// test case
//
// no need login
// success: phpmyadmin-cli -url 127.0.0.1:8000
// success: phpmyadmin-cli -url 127.0.0.1:8000 -username root
// success: phpmyadmin-cli -url 127.0.0.1:8000 -username root -password pass

// need login
// need login: phpmyadmin-cli -url 127.0.0.1:8000
// login failed: phpmyadmin-cli -url 127.0.0.1:8000 -username root
// success: phpmyadmin-cli -url 127.0.0.1:8000 -username root -password pass
