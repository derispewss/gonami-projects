package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/derispewss/finwa-projects/internal/ai"
	"github.com/derispewss/finwa-projects/internal/application"
	"github.com/derispewss/finwa-projects/internal/config"
	"github.com/derispewss/finwa-projects/internal/database"
	"github.com/derispewss/finwa-projects/internal/parser"
	"github.com/derispewss/finwa-projects/internal/repository"
	"github.com/derispewss/finwa-projects/internal/storage"
	"github.com/derispewss/finwa-projects/internal/whatsapp"

	_ "github.com/jackc/pgx/v5/stdlib"

	_ "modernc.org/sqlite"
)

func main() {

	if err := godotenv.Load(); err != nil {

		slog.Debug("no .env file found, using environment variables")
	}

	setupLogger()

	slog.Info("starting finwa bot")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("running database migrations")
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations completed")

	waClient, err := whatsapp.NewClient(ctx, cfg.WhatsAppDBPath)
	if err != nil {
		slog.Error("failed to create whatsapp client", "error", err)
		os.Exit(1)
	}

	prs := parser.New()

	store := storage.New(cfg)

	var aiClient ai.AIClient
	if cfg.GeminiAPIKey == "" {
		slog.Info("GEMINI_API_KEY tidak diset — fitur voice note & foto struk dinonaktifkan")
	} else {
		budget := ai.NewTokenSaver(cfg.LLMDailyBudget)
		gemini, err := ai.NewGemini(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiModelTx,
			budget, int32(cfg.LLMMaxOutputTokens))
		if err != nil {
			slog.Warn("gagal inisialisasi Gemini — fitur media dinonaktifkan", "error", err)
		} else {
			aiClient = gemini
			if cfg.GeminiModelTx != "" {
				slog.Info("Gemini aktif", "model", cfg.GeminiModel,
					"model_text", cfg.GeminiModelTx, "budget_harian", cfg.LLMDailyBudget)
			} else {
				slog.Info("Gemini aktif", "model", cfg.GeminiModel,
					"budget_harian", cfg.LLMDailyBudget)
			}
		}
	}

	app := application.NewApp(db, prs, cfg, aiClient, store)

	sender := whatsapp.NewSender(waClient.WA)
	router := whatsapp.NewRouter(waClient, sender, app)
	handler := whatsapp.NewHandler(router, sender)
	waClient.SetHandler(handler)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		draftRepo := repository.NewDraftRepo(db)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count, _ := draftRepo.ExpireOldDrafts(ctx)
				if count > 0 {
					slog.Info("expired old drafts", "count", count)
				}
			}
		}
	}()

	if err := waClient.Connect(ctx); err != nil {
		slog.Error("failed to connect to whatsapp", "error", err)
		os.Exit(1)
	}

	slog.Info("finwa bot is running. Press Ctrl+C to stop.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received, stopping gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	cancel()

	waClient.Disconnect()

	db.Close()

	<-shutdownCtx.Done()
	slog.Info("finwa bot stopped")
}

func setupLogger() {
	levelStr := os.Getenv("LOG_LEVEL")
	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}

func runMigrations(databaseURL string) error {

	db, err := goose.OpenDBWithDriver("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "./migrations")
}
