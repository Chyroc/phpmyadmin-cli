package common

import "fmt"

var IsDebug1 = false

var ErrNeedLogin = fmt.Errorf("need login")
var ErrLoginFailed = fmt.Errorf("login failed")
