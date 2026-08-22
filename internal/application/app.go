package application

import (
	"github.com/derispewss/finwa-projects/internal/ai"
	"github.com/derispewss/finwa-projects/internal/config"
	"github.com/derispewss/finwa-projects/internal/media"
	"github.com/derispewss/finwa-projects/internal/parser"
	"github.com/derispewss/finwa-projects/internal/repository"
	"github.com/derispewss/finwa-projects/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Record  *RecordTransaction
	Confirm *ConfirmDraft
	Report  *GetReport
	Balance *GetBalance
	Manage  *ManageTransaction
	Media   *ProcessMedia
}

func NewApp(db *pgxpool.Pool, prs *parser.Parser, cfg *config.Config,
	aiClient ai.AIClient, store storage.Storage) *App {

	userRepo := repository.NewUserRepo(db)
	catRepo := repository.NewCategoryRepo(db)
	txRepo := repository.NewTransactionRepo(db)
	draftRepo := repository.NewDraftRepo(db)

	app := &App{
		Record:  NewRecordTransaction(userRepo, catRepo, txRepo, draftRepo, prs, cfg, aiClient),
		Confirm: NewConfirmDraft(userRepo, txRepo, catRepo, draftRepo),
		Report:  NewGetReport(userRepo, txRepo),
		Balance: NewGetBalance(userRepo, txRepo),
		Manage:  NewManageTransaction(userRepo, txRepo),
	}

	if aiClient != nil {
		app.Media = NewProcessMedia(
			app.Record,
			store,
			media.NewAudioProcessor(aiClient),
			media.NewImageProcessor(aiClient),
			media.NewDocumentProcessor(aiClient),
			cfg.MaxAudioSizeBytes,
			cfg.MaxImageSizeBytes,
			cfg.MaxPDFSizeBytes,
		)
	}
	return app
}
