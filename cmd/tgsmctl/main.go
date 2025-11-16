package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/covrom/smart-control/internal/api"
	"github.com/covrom/smart-control/internal/cron"
	"github.com/covrom/smart-control/internal/llmdesc"
	"github.com/covrom/smart-control/internal/smartdata"
	tele "gopkg.in/telebot.v3"
)

func main() {
	modes := strings.Split(os.Getenv("MODE"), ",") // agent,server
	openaiBaseUrl := os.Getenv("OPENAI_BASE_URL")
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	openaiModel := os.Getenv("OPENAI_MODEL")
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")
	token := os.Getenv("HTTP_AUTH_TOKEN")                      // agent
	hostname := os.Getenv("SMART_HOSTNAME")                    // agent
	apiUrl := os.Getenv("COLLECTOR_URL")                       // agent
	cronSched := strings.Trim(os.Getenv("CRON_SCHEDULE"), `"`) // agent
	if cronSched == "" {
		cronSched = "55 23 * * *"
	}

	var isAgent, isServer bool
	for _, mode := range modes {
		isAgent = isAgent || strings.EqualFold(strings.TrimSpace(mode), "agent")
		isServer = isServer || strings.EqualFold(strings.TrimSpace(mode), "server")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	wg := &sync.WaitGroup{}

	l := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelInfo},
	))
	slog.SetDefault(l)

	l.Info("start", "isAgent", isAgent, "isServer", isServer)
	defer l.Info("stop")

	// агент
	if isAgent {
		sched, err := cron.NewCronScheduleFromString(cronSched)
		if err != nil {
			log.Fatal(err)
			return
		}

		slog.Info("cron sheduling", "cronSched", cronSched)

		wg.Add(1)
		go workerSendReports(ctx, wg, apiUrl, token, hostname, sched)
	}

	// сервер
	if isServer {
		chReps := make(chan smartdata.CommonSMARTReport, 100)
		chTgMsg := make(chan string, 100)

		pref := tele.Settings{
			Token:  telegramToken,
			Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		}

		b, err := tele.NewBot(pref)
		if err != nil {
			log.Fatal(err)
			return
		}

		srv := api.NewHttpServer(token, chReps)

		wg.Add(1)
		go api.TgSendWorker(ctx, b, telegramChatID, chTgMsg, wg)

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			srv.Shutdown(context.Background())
		}()

		llmDescriber := llmdesc.NewLLMDescriber(openaiBaseUrl, openaiApiKey, openaiModel)

		wg.Add(1)
		go workerRecvReports(ctx, wg, hostname, chTgMsg, chReps, llmDescriber)
	}

	<-ctx.Done()

	wg.Wait()
}
