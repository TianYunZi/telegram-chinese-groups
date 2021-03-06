package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"gopkg.in/redis.v3"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/kylelemons/go-gypsy/yaml"
)

func main() {
	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	groups, err := yaml.ReadFile("botconf.yaml")
	if err != nil {
		log.Panic(err)
	}

	botapi, err := groups.Get("botapi")
	if err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(botapi)
	if err != nil {
		log.Panic(err)
	}

	botname := bot.Self.UserName

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.UpdatesChan(u)

	for update := range updates {

		log.Printf("[%d](%s) -- [%s] -- %s",
			update.Message.Chat.ID, update.Message.Chat.Title,
			update.Message.From.UserName, update.Message.Text,
		)

		u := Updater{
			redis:            rc,
			bot:              bot,
			update:           update,
			enableGroupLimit: true,
			limitInterval:    time.Minute * 15,
			limitTimes:       2,
		}

		switch update.Message.Text {

		case "/help", "/start", "/help@" + botname, "/start@" + botname:
			go u.SendMessage(YamlList2String(groups, "help"))

		case "/rules", "/rules@" + botname:
			go u.SendMessage(YamlList2String(groups, "rules"))

		case "/about", "/about@" + botname:
			go u.SendMessage(YamlList2String(groups, "about"))

		case "/linux", "/linux@" + botname:
			go u.SendMessage(YamlList2String(groups, "Linux"))

		case "/programming", "/programming@" + botname:
			go u.SendMessage(YamlList2String(groups, "Programming"))

		case "/software", "/software@" + botname:
			go u.SendMessage(YamlList2String(groups, "Software"))

		case "/videos", "/videos@" + botname:
			go u.SendMessage(YamlList2String(groups, "影音"))

		case "/sci_fi", "/sci_fi@" + botname:
			go u.SendMessage(YamlList2String(groups, "科幻"))

		case "/acg", "/acg@" + botname:
			go u.SendMessage(YamlList2String(groups, "ACG"))

		case "/it", "/it@" + botname:
			go u.SendMessage(YamlList2String(groups, "IT"))

		case "/free_chat", "/free_chat@" + botname:
			go u.SendMessage(YamlList2String(groups, "闲聊"))

		case "/resources", "/resources@" + botname:
			go u.SendMessage(YamlList2String(groups, "资源"))

		case "/same_city", "/same_city@" + botname:
			go u.SendMessage(YamlList2String(groups, "同城"))

		case "/others", "/others@" + botname:
			go u.SendMessage(YamlList2String(groups, "Others"))

		case "/other_resources", "/other_resources@" + botname:
			go u.SendMessage(YamlList2String(groups, "其他资源"))

		default:

		}
	}
}

type Updater struct {
	redis            *redis.Client
	bot              *tgbotapi.BotAPI
	update           tgbotapi.Update
	enableGroupLimit bool
	limitTimes       int64
	limitInterval    time.Duration
}

func (u *Updater) SendMessage(msgText string) {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)

	if u.enableGroupLimit && u.update.Message.Chat.ID < 0 {
		if u.redis.Exists(chatIDStr).Val() {
			u.redis.Incr(chatIDStr)
			counter, _ := u.redis.Get(chatIDStr).Int64()
			if counter >= u.limitTimes {
				log.Println("--- " + u.update.Message.Chat.Title + " --- " + "防刷屏 ---")
				msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
					"刷屏是坏孩纸~！\n聪明宝宝是会跟奴家私聊的哟😊\n@"+u.bot.Self.UserName)
				msg.ReplyToMessageID = u.update.Message.MessageID
				u.bot.SendMessage(msg)
				return
			}
		} else {
			u.redis.Set(chatIDStr, "0", u.limitInterval)
		}
	}

	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID, msgText)
	u.bot.SendMessage(msg)
	return
}

func YamlList2String(config *yaml.File, text string) string {
	count, err := config.Count(text)
	if err != nil {
		log.Println(err)
		return ""
	}

	var resultGroup []string
	for i := 0; i < count; i++ {
		v, err := config.Get(text + "[" + strconv.Itoa(i) + "]")
		if err != nil {
			log.Println(err)
			return ""
		}
		resultGroup = append(resultGroup, v)
	}

	result := strings.Join(resultGroup, "\n")
	result = strings.Replace(result, "\\n", "", -1)

	return result
}
