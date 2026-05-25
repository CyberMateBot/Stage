package internal

import (
	"context"

	"github.com/jackc/pgx/v5"
	rcModels "github.com/twelvepills-936/tgapp-/internal/rest/client/models"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
)

type Repository interface {
	DBBeginTransaction(ctx context.Context) (pgx.Tx, error)

	ReadUser(ctx context.Context, id int64, dbTx pgx.Tx) (user repoModels.User, err error)

	// CyberMate repositories
<<<<<<< HEAD
    CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error)
    GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error)
    UpdateProfileTheme(ctx context.Context, tx pgx.Tx, telegramID, theme string) error
    CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error)
    AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error
=======
	CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error)
	GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error)
	CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error)
	AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error
>>>>>>> 3489ac71c17ae6e070eec77e5b2b0b383107f257
}

type UseCase interface {
	GetUser(ctx context.Context, input ucModels.GetUserInput) (output ucModels.GetUserOutput, err error)

	// CyberMate usecases
<<<<<<< HEAD
    RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error)
    GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error)
    UpdateProfileTheme(ctx context.Context, input ucModels.UpdateProfileThemeInput) (ucModels.UpdateProfileThemeOutput, error)
=======
	RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error)
	GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error)
>>>>>>> 3489ac71c17ae6e070eec77e5b2b0b383107f257
}

type Client interface {
	PostingsToCancel(ctx context.Context, token string, req rcModels.PostingsToCancelReq) (rcModels.PostingsToCancelResp, error)
	PostingsCancelResponse(ctx context.Context, token string, req rcModels.PostingsCancelResponseReq) (rcModels.PostingsCancelResponseResp, error)
}
