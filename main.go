package main

import (
	"MaxKimServerBot/clcolor"
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/MscBaiMeow/go-mc/bot"
	"github.com/MscBaiMeow/go-mc/yggdrasil"
	"github.com/Tnze/go-mc/chat"
	"os"
	"time"
)

var c *bot.Client
var timer int64
var locker bool

type config struct {
	User     string
	Password string

	Server string
	Port   int

	InitMsg   []string
	TimingMsg []string
	Interval  int

	DelayAfterDisconnect int
}

var cfg = new(config)

func tLog(str string) {
	fmt.Println(clcolor.Magenta("["+time.Now().Format("2006-01-02 15:04:05")+"] ") + str)
}

func main() {
	fmt.Println(clcolor.Cyan("欢迎使用CmdBlock_zqg的1.15.2挂机钓鱼机器人"))

	fmt.Print(clcolor.Cyan("读取配置文件……"))
	if _, err := toml.DecodeFile("./config.toml", &cfg); err != nil {
		fmt.Println()
		panic("配置文件读取失败！请检查配置文件格式是否正确")
	}
	fmt.Println(clcolor.Green("成功"))

	fmt.Print(clcolor.Cyan("登录账号……"))
	loginResp, err := yggdrasil.Authenticate(cfg.User, cfg.Password)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		panic("账号登录失败！")
	}
	fmt.Println(clcolor.Green("成功"))

	c = bot.NewClient()
	c.Auth.UUID, c.Name = loginResp.SelectedProfile()
	c.AsTk = loginResp.AccessToken()

	fmt.Print(clcolor.Cyan("加入服务器……"))
	err = c.JoinServer(cfg.Server, cfg.Port)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		panic("加入服务器失败！")
	}
	fmt.Println(clcolor.Green("成功"))

	go sendAutoCmd()

	go watcher()

	go commander()

	c.Events.GameStart = onGameStart
	c.Events.SoundPlay = onSoundPlay

	c.Events.EntityRelativeMove = onEntityRelativeMove
	c.Events.SpawnObj = onSpawnObj
	c.Events.Disconnect = onDisconnect
	c.Events.ChatMsg = onChatMsg

	err = c.HandleGame()
	if err != nil {
		panic(err)
	}
}

func sendAutoCmd() {
	time.Sleep(time.Second * 5)
	tLog(clcolor.Cyan("执行启动命令："))
	for _, cmd := range cfg.InitMsg {
		time.Sleep(time.Second)
		tLog(cmd)
		_ = c.Chat(cmd)
	}
	for {
		tLog(clcolor.Cyan("执行定时命令："))
		for _, cmd := range cfg.TimingMsg {
			time.Sleep(time.Second)
			tLog(cmd)
			_ = c.Chat(cmd)
		}
		time.Sleep(time.Second * time.Duration(cfg.Interval))
	}
}

func onGameStart() error {
	time.Sleep(time.Second)
	tLog(clcolor.Yellow("抛竿！"))
	timer = time.Now().Unix()
	locker = false
	return c.UseItem(0)
}

func onSoundPlay(name string, category int, x, y, z float64, volume, pitch float32) error {
	if name != "block.bubble_column.whirlpool_inside" {
		return nil
	}
	locker = true
	if err := c.UseItem(0); err != nil {
		return err
	}
	tLog(clcolor.Yellow("收竿！"))
	time.Sleep(time.Millisecond * time.Duration(1500))
	if err := c.UseItem(0); err != nil {
		return err
	}
	tLog(clcolor.Yellow("抛竿！"))
	timer = time.Now().Unix()
	locker = false
	return nil
}

func onEntityRelativeMove(EID, DeltaX, DeltaY, DeltaZ int) error {
	return nil
}

func onSpawnObj(EID int, UUID [16]byte, Type int, x, y, z float64, Pitch, Yaw float32, Data int, VelocityX, VelocitY, VelocitZ int16) error {
	return nil
}

func onDisconnect(c chat.Message) error {
	fmt.Println(c)
	tLog("与服务器断开连接")
	time.Sleep(time.Millisecond * time.Duration(cfg.DelayAfterDisconnect))
	panic("与服务器断开连接")
}

func onChatMsg(c chat.Message, pos byte) error {
	tLog(clcolor.Cyan("聊天内容：") + c.String())
	return nil
}

func watcher() {
	time.Sleep(time.Second * 20)
	for {
		time.Sleep(time.Second * 5)
		if locker {
			continue
		}
		nowTime := time.Now().Unix()
		if nowTime >= timer+60 {
			locker = true
			_ = c.SelectItem(1)
			time.Sleep(time.Millisecond * time.Duration(1500))
			_ = c.SelectItem(0)
			tLog(clcolor.Yellow("等不及了！收竿！"))
			time.Sleep(time.Millisecond * time.Duration(1500))
			_ = c.UseItem(0)
			tLog(clcolor.Yellow("等不及了！抛竿！"))
			timer = time.Now().Unix()
			locker = false
		}
	}
}

func commander() {
	var cmd string
	for {
		Reader := bufio.NewReader(os.Stdin)
		cmd, _ = Reader.ReadString('\n')
		_ = c.Chat(cmd[:len(cmd)-1])
	}
}
