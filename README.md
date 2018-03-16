# phpmyadmin-cli
access phpmyadmin from cli / 通过shell操作phpmyadmin

## features
* access phpmyadmin from cli
* grammar tip

## install
```
go get github.com/Chyroc/phpmyadmin-cli
```

## use

```
➜  ~ phpmyadmin-cli -h
NAME:
   phpmyadmin-cli - access phpmyadmin from shell cli

USAGE:
   phpmyadmin-cli [global options] [arguments...]

GLOBAL OPTIONS:
   --url value      phpmyadmin url
   --prune          清理命令记录
   --history value  command history file (default: "~/.phpmyadmin_cli_history")
   --help, -h       show help
```

```
➜  ~ phpmyadmin-cli -url ip:port
>>>
```
