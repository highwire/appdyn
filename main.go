package main

import (
	"flag"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/vaughan0/go-ini"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
)

type ADConfig struct {
	PHPEnabled bool
	PHPAppName string
}

func NewADConfigFromEnvironment() *ADConfig {
	phpConfig, err := ini.LoadFile(phpConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	conf := new(ADConfig)
	_, conf.PHPEnabled = phpConfig.Get("", "extension")
	conf.PHPAppName, _ = phpConfig.Get("", "agent.applicationName")

	return conf
}

func (conf *ADConfig) String() string {
	var phpStatus string
	if conf.PHPEnabled {
		phpStatus = "enabled"
	} else {
		phpStatus = "disabled"
	}

	str := "php\t\t" + phpStatus + "\n"
	str += "php-name\t" + conf.PHPAppName + "\n"

	return str
}

func (conf *ADConfig) Write() {
	var phpConfig string
	if conf.PHPEnabled {
		phpConfig += "extension = appdynamics_agent.so\n"
	}
	if conf.PHPAppName != "" {
		phpConfig += "agent.applicationName = " + conf.PHPAppName + "\n"
	}

	fileInfo, err := os.Stat(phpConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(phpConfigPath, []byte(phpConfig), fileInfo.Mode())
	if err != nil {
		log.Fatal(err)
	}
}

func checkSudo() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	if usr.Uid != "0" {
		log.Fatal("You must be root or use sudo to change appDyanmic settings")
	}
}

func restartApache() {
	cmd := exec.Command("/etc/init.d/httpd", "restart")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Could not restart apache\n" + string(out))
	}
}

func commandHelp() {
	help := "appdyn is a command-line utility to enable and disable app-dynamic agents\n"
	help += "Example commands:\n"
	help += "  appdyn help\n"
	help += "  appdyn status\n"
	help += "  appdyn en-php\n"
	help += "  appdyn en-php <app-name>\n"
	help += "  appdyn dis-php\n"
	help += "  appdyn php-name <app-name>\n"

	fmt.Print(help)
}

func commandStatus() {
	conf := NewADConfigFromEnvironment()
	fmt.Print(conf)
}

func commandEnPHP() {
	checkSudo()

	conf := NewADConfigFromEnvironment()
	conf.PHPEnabled = true

	if flag.Arg(1) != "" {
		conf.PHPAppName = flag.Arg(1)
	}

	if conf.PHPAppName == "" {
		log.Fatal("PHP AppDynamics Machine Agent does not have an AppName. Please specify one.")
	}

	conf.Write()
	restartApache()
}

func commandDisPHP() {
	checkSudo()

	conf := NewADConfigFromEnvironment()
	conf.PHPEnabled = false

	conf.Write()
	restartApache()
}

func commandPHPName() {
	checkSudo()

	if flag.Arg(1) == "" {
		log.Fatal("You must specify a name")
	}

	conf := NewADConfigFromEnvironment()
	conf.PHPAppName = flag.Arg(1)

	conf.Write()
	if conf.PHPEnabled {
		restartApache()
	}
}

var (
	phpConfigPath = "/etc/php.d/appdynamics_agent.ini"
)

func main() {
	flag.Parse()
	command := flag.Arg(0)
	switch command {
	case "", "help":
		commandHelp()
	case "status":
		commandStatus()
	case "en-php":
		commandEnPHP()
	case "dis-php":
		commandDisPHP()
	case "php-name":
		commandPHPName()
	default:
		commandHelp()
	}
}
