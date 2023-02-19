package job

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
	"x-ui/logger"
	"x-ui/util/common"
	"x-ui/web/service"
)

var SSHLoginUser int

type LoginStatus byte

const (
	LoginSuccess LoginStatus = 1
	LoginFail    LoginStatus = 0
)

type StatsNotifyJob struct {
	enable          bool
	telegramService service.TelegramService
	xrayService     service.XrayService
	inboundService  service.InboundService
	settingService  service.SettingService
}

func NewStatsNotifyJob() *StatsNotifyJob {
	return new(StatsNotifyJob)
}

//Here run is a interface method of Job interface
func (j *StatsNotifyJob) Run() {
	if !j.xrayService.IsXrayRunning() {
		return
	}
	var info string
	info = j.telegramService.GetsystemStatus()
	j.telegramService.SendMsgToTgbot(info)
}

func (j *StatsNotifyJob) UserLoginNotify(username string, ip string, time string, status LoginStatus) {
	if username == "" || ip == "" || time == "" {
		logger.Warning("UserLoginNotify failed, invalid info")
		return
	}
	var msg string
	//get hostname
	name, err := os.Hostname()
	if err != nil {
		fmt.Println("get hostname error: ", err)
		return
	}
	if status == LoginSuccess {
		msg = fmt.Sprintf("Panel login successful reminder\r\nHost name: %s\r\n", name)
	} else if status == LoginFail {
		msg = fmt.Sprintf("Panel login failure reminder reminder\r\nHost name: %s\r\n", name)
	}
	msg += fmt.Sprintf("Time: %s\r\n", time)
	msg += fmt.Sprintf("User: %s\r\n", username)
	msg += fmt.Sprintf("IP: %s\r\n", ip)
	j.telegramService.SendMsgToTgbot(msg)
}

func (j *StatsNotifyJob) SSHStatusLoginNotify(xuiStartTime string) {
	getSSHUserNumber, error := exec.Command("bash", "-c", "who | awk  '{print $1}' | wc -l").Output()
	if error != nil {
		fmt.Println("getSSHUserNumber error: ", error)
		return
	}
	var numberInt int
	numberInt, error = strconv.Atoi(common.ByteToString(getSSHUserNumber))
	if error != nil {
		return
	}
	if numberInt > SSHLoginUser {
		var SSHLoginInfo string
		SSHLoginUser = numberInt
		//hostname
		name, err := os.Hostname()
		if err != nil {
			fmt.Println("get hostname error: ", err)
			return
		}
		//Time compare,need if x-ui got restart while ssh already exist users
		SSHLoginTime, error := exec.Command("bash", "-c", "who | awk  '{print $3,$4}' | tail -n 1 ").Output()
		if error != nil {
			fmt.Println("getLoginTime error: ", error.Error())
			return
		}
		var SSHLoginTimeStr string
		SSHLoginTimeStr = common.ByteToString(SSHLoginTime)
		t1, err := time.Parse("2006-01-02 15:04:05", SSHLoginTimeStr)
		t2, err := time.Parse("2006-01-02 15:04:05", xuiStartTime)
		if t1.Before(t2) || err != nil {
			fmt.Printf("SSHLogin[%s] early than XUI start[%s]\r\n", SSHLoginTimeStr, xuiStartTime)
		}

		SSHLoginUserName, error := exec.Command("bash", "-c", "who | awk  '{print $1}'| tail -n 1").Output()
		if error != nil {
			fmt.Println("getSSHLoginUserName error: ", error.Error())
			return
		}

		SSHLoginIpAddr, error := exec.Command("bash", "-c", "who | tail -n 1 | awk -F [\\(\\)] 'NR==1 {print $2}'").Output()
		if error != nil {
			fmt.Println("getSSHLoginIpAddr error: ", error)
			return
		}

		SSHLoginInfo = fmt.Sprintf("⚠ SSH user login reminder:\r\n")
		SSHLoginInfo += fmt.Sprintf("Host name: %s\r\n", name)
		SSHLoginInfo += fmt.Sprintf("Login user: %s", SSHLoginUserName)
		SSHLoginInfo += fmt.Sprintf("Log in time: %s", SSHLoginTime)
		SSHLoginInfo += fmt.Sprintf("Login IP: %s", SSHLoginIpAddr)
		SSHLoginInfo += fmt.Sprintf("Currently logging in the number of users: %s", getSSHUserNumber)
		j.telegramService.SendMsgToTgbot(SSHLoginInfo)
	} else {
		SSHLoginUser = numberInt
	}
}
