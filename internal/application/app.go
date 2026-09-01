package application

import (
	"github.com/derispewss/gonami-projects/internal/ai"
	"github.com/derispewss/gonami-projects/internal/config"
	"github.com/derispewss/gonami-projects/internal/media"
	"github.com/derispewss/gonami-projects/internal/parser"
	"github.com/derispewss/gonami-projects/internal/repository"
	"github.com/derispewss/gonami-projects/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Record    *RecordTransaction
	Confirm   *ConfirmDraft
	Report    *GetReport
	Balance   *GetBalance
	Manage    *ManageTransaction
	Media     *ProcessMedia
	Budget    *BudgetUC
	Insight   *InsightUC
	Export    *ExportUC
	Wallet    *WalletUC
	Category  *ManageCategory
	ResetData *ResetData
}

func NewApp(db *pgxpool.Pool, prs *parser.Parser, cfg *config.Config,
	aiClient ai.AIClient, store storage.Storage) *App {

	userRepo := repository.NewUserRepo(db)
	catRepo := repository.NewCategoryRepo(db)
	txRepo := repository.NewTransactionRepo(db)
	draftRepo := repository.NewDraftRepo(db)
	budgetRepo := repository.NewBudgetRepo(db)
	walletRepo := repository.NewWalletRepo(db)

	app := &App{
		Record:    NewRecordTransaction(userRepo, catRepo, txRepo, draftRepo, prs, cfg, aiClient),
		Confirm:   NewConfirmDraft(userRepo, txRepo, catRepo, draftRepo),
		Report:    NewGetReport(userRepo, txRepo),
		Balance:   NewGetBalance(userRepo, txRepo),
		Manage:    NewManageTransaction(userRepo, txRepo),
		Budget:    NewBudgetUC(userRepo, txRepo, budgetRepo),
		Insight:   NewInsightUC(userRepo, txRepo),
		Export:    NewExportUC(userRepo, txRepo),
		Wallet:    NewWalletUC(userRepo, txRepo, walletRepo),
		Category:  NewManageCategory(userRepo, catRepo),
		ResetData: NewResetData(userRepo, txRepo, budgetRepo, walletRepo, catRepo, draftRepo),
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
