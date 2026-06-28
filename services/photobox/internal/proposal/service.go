package proposal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"csort.ru/coffeebot/internal/database"
	"csort.ru/coffeebot/internal/dto"
	"csort.ru/coffeebot/internal/minio"
	"csort.ru/coffeebot/internal/weed"
	"csort.ru/coffeebot/internal/weed/analysis"
	"csort.ru/coffeebot/internal/weed/image"
	"csort.ru/coffeebot/internal/weed/stats"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kxrxh/gopt"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	queries         database.Querier
	txRunner        database.TxRunner
	objectStore     minio.ObjectStorage
	weedService     *weed.Service
	weedStatsSvc    *stats.Service
	weedAnalysesSvc *analysis.Service
	weedImageSvc    *image.Service
}

var proposalListRowConverter rowConverter = &rowConverterImpl{}

func NewService(
	queries database.Querier,
	txRunner database.TxRunner,
	objectStore minio.ObjectStorage,
	weedService *weed.Service,
	weedStatsSvc *stats.Service,
	weedAnalysesSvc *analysis.Service,
	weedImageSvc *image.Service,
) *Service {
	return &Service{
		queries:         queries,
		txRunner:        txRunner,
		objectStore:     objectStore,
		weedService:     weedService,
		weedStatsSvc:    weedStatsSvc,
		weedAnalysesSvc: weedAnalysesSvc,
		weedImageSvc:    weedImageSvc,
	}
}

// CreateProposal creates a submitted proposal; with TargetWeedID the draft is prefilled from that weed.
// RequestBy stores the internal user_id (not a messenger id).
func (s *Service) CreateProposal(
	ctx context.Context,
	params CreateProposalParams,
	userID int32,
) (*Proposal, error) {
	var requestID int32
	err := s.txRunner.Run(ctx, func(qTx database.Querier) error {
		request, createErr := qTx.CreateCatalogProposal(ctx, database.CreateCatalogProposalParams{
			Status:       "submitted",
			RequestBy:    int64(userID),
			TargetWeedID: params.TargetWeedID,
		})
		if createErr != nil {
			return fmt.Errorf("failed to create proposal: %w", createErr)
		}
		requestID = request.ID

		var pendingWeed database.PendingWeed
		if targetID, ok := gopt.FromPtr(params.TargetWeedID).Get(); ok {
			weed, err := s.weedService.GetWeedByID(ctx, targetID)
			if err != nil {
				return fmt.Errorf("failed to load target weed: %w", err)
			}

			name := weed.Name
			if params.Name != "" {
				name = params.Name
			}
			description := weed.Description
			if params.Description != "" {
				d := params.Description
				description = &d
			}
			harmfulness := params.Harmfulness
			if harmfulness == nil {
				harmfulness = weed.Harmfulness
			}
			mainGroup := params.MainGroup
			if mainGroup == nil || *mainGroup == "" {
				mainGroup = weed.MainGroup
			}
			mainSubgroup := params.MainSubgroup
			if mainSubgroup == nil || *mainSubgroup == "" {
				mainSubgroup = weed.MainSubgroup
			}
			subgroup := params.Subgroup
			if subgroup == nil || *subgroup == "" {
				subgroup = weed.Subgroup
			}
			length, width := weed.Length, weed.Width
			if params.Statistics != nil {
				if params.Statistics.LMedian != 0 {
					l := params.Statistics.LMedian
					length = &l
				}
				if params.Statistics.WMedian != 0 {
					w := params.Statistics.WMedian
					width = &w
				}
			}

			var createErr error
			pendingWeed, createErr = qTx.CreatePendingWeed(ctx, database.CreatePendingWeedParams{
				Name:         name,
				LatinName:    weed.LatinName,
				Description:  description,
				Length:       length,
				Width:        width,
				MainGroup:    mainGroup,
				MainSubgroup: mainSubgroup,
				Subgroup:     subgroup,
				IsQuarantine: weed.IsQuarantine,
				Harmfulness:  harmfulness,
				ProposalID:   request.ID,
			})
			if createErr != nil {
				return fmt.Errorf("failed to create pending weed from existing: %w", createErr)
			}

			if err := s.copyWeedDataToProposal(ctx, qTx, targetID, pendingWeed.ID); err != nil {
				return fmt.Errorf("failed to copy weed data: %w", err)
			}

			if params.AnalysisIDs != nil && len(*params.AnalysisIDs) > 0 {
				if err := replacePendingWeedAnalyses(
					ctx,
					qTx,
					pendingWeed.ID,
					*params.AnalysisIDs,
				); err != nil {
					return fmt.Errorf("failed to replace analyses: %w", err)
				}
			}

			if params.Statistics != nil {
				excludedObjects := []int64{}
				if params.ExcludedObjects != nil {
					excludedObjects = *params.ExcludedObjects
				}
				if len(excludedObjects) == 0 && params.Statistics.ExcludedObjects != nil {
					excludedObjects = params.Statistics.ExcludedObjects
				}
				excludedBytes, err := sonic.Marshal(excludedObjects)
				if err != nil {
					return fmt.Errorf("failed to serialize excluded objects: %w", err)
				}
				_, getStatsErr := qTx.GetPendingWeedStatsByPendingWeedID(ctx, pendingWeed.ID)
				if getStatsErr == nil {
					_, err = qTx.UpdatePendingWeedStats(ctx, newUpdatePendingWeedStatsParams(
						pendingWeed.ID,
						params.Statistics,
						excludedBytes,
					))
					if err != nil {
						return fmt.Errorf("failed to update stats override: %w", err)
					}
				} else {
					_, err = qTx.CreatePendingWeedStats(ctx, newCreatePendingWeedStatsParams(
						pendingWeed.ID,
						params.Statistics,
						excludedBytes,
					))
					if err != nil {
						return fmt.Errorf("failed to create stats override: %w", err)
					}
				}
			}
		} else {
			var description *string
			if params.Description != "" {
				d := params.Description
				description = &d
			}
			var length, width *float32
			if params.Statistics != nil {
				if params.Statistics.LMedian != 0 {
					l := params.Statistics.LMedian
					length = &l
				}
				if params.Statistics.WMedian != 0 {
					w := params.Statistics.WMedian
					width = &w
				}
			}

			var createErr error
			pendingWeed, createErr = qTx.CreatePendingWeed(ctx, database.CreatePendingWeedParams{
				Name:         params.Name,
				LatinName:    nil,
				Description:  description,
				Length:       length,
				Width:        width,
				MainGroup:    params.MainGroup,
				MainSubgroup: params.MainSubgroup,
				Subgroup:     params.Subgroup,
				IsQuarantine: false,
				Harmfulness:  params.Harmfulness,
				ProposalID:   request.ID,
			})
			if createErr != nil {
				return fmt.Errorf("failed to create pending weed: %w", createErr)
			}
		}

		if params.AnalysisIDs != nil && len(*params.AnalysisIDs) > 0 {
			if err := qTx.BulkInsertPendingWeedAnalyses(
				ctx,
				database.BulkInsertPendingWeedAnalysesParams{
					PendingWeedID: pendingWeed.ID,
					AnalysisIds:   *params.AnalysisIDs,
				},
			); err != nil {
				return fmt.Errorf("failed to add analysis: %w", err)
			}
		}

		if params.Statistics != nil {
			if params.ExcludedObjects != nil && len(*params.ExcludedObjects) > 0 {
				params.Statistics.ExcludedObjects = *params.ExcludedObjects
			}
			excludedBytes, err := sonic.Marshal(params.Statistics.ExcludedObjects)
			if err != nil {
				return fmt.Errorf("failed to serialize excluded objects: %w", err)
			}

			if _, err := qTx.CreatePendingWeedStats(ctx, newCreatePendingWeedStatsParams(
				pendingWeed.ID,
				params.Statistics,
				excludedBytes,
			)); err != nil {
				return fmt.Errorf("failed to create stats: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetProposalByID(ctx, requestID)
}

func replacePendingWeedAnalyses(
	ctx context.Context,
	q database.Querier,
	pendingWeedID int32,
	analysisIDs []string,
) error {
	if err := q.DeletePendingWeedAnalysesByPendingWeedID(ctx, pendingWeedID); err != nil {
		return fmt.Errorf("failed to clear pending analyses: %w", err)
	}
	if len(analysisIDs) == 0 {
		return nil
	}
	if err := q.BulkInsertPendingWeedAnalyses(ctx, database.BulkInsertPendingWeedAnalysesParams{
		PendingWeedID: pendingWeedID,
		AnalysisIds:   analysisIDs,
	}); err != nil {
		return fmt.Errorf("failed to insert pending analyses: %w", err)
	}
	return nil
}

func (s *Service) copyWeedDataToProposal(
	ctx context.Context,
	qTx database.Querier,
	weedID int32,
	pendingWeedID int32,
) error {
	var (
		images      []database.WeedImage
		analyses    []database.WeedAnalysis
		stats       database.WeedStat
		statsLookup error
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		images, err = s.weedImageSvc.GetWeedImages(gctx, weedID)
		if err != nil {
			return fmt.Errorf("failed to get weed images: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		analyses, err = s.weedAnalysesSvc.GetWeedAnalyses(gctx, weedID)
		if err != nil {
			return fmt.Errorf("failed to get analyses: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		stats, err = s.queries.GetWeedStatsByWeedID(gctx, weedID)
		statsLookup = err
		return nil
	})
	if err := g.Wait(); err != nil {
		return err
	}

	if len(images) > 0 {
		imageKeys := make([]string, len(images))
		isPrimary := make([]bool, len(images))
		for i, img := range images {
			imageKeys[i] = img.ImageKey
			isPrimary[i] = img.IsPrimary
		}
		if err := qTx.BulkInsertPendingWeedImages(ctx, database.BulkInsertPendingWeedImagesParams{
			PendingWeedID:  pendingWeedID,
			ImageKeys:      imageKeys,
			IsPrimaryFlags: isPrimary,
		}); err != nil {
			return fmt.Errorf("failed to copy images: %w", err)
		}
	}

	if len(analyses) > 0 {
		ids := make([]string, len(analyses))
		for i, a := range analyses {
			ids[i] = a.AnalysisID
		}
		if err := qTx.BulkInsertPendingWeedAnalyses(
			ctx,
			database.BulkInsertPendingWeedAnalysesParams{
				PendingWeedID: pendingWeedID,
				AnalysisIds:   ids,
			},
		); err != nil {
			return fmt.Errorf("failed to copy analyses: %w", err)
		}
	}

	if statsLookup == nil {
		if _, err := qTx.CreatePendingWeedStats(
			ctx,
			newCreatePendingWeedStatsFromWeedStat(pendingWeedID, stats),
		); err != nil {
			return fmt.Errorf("failed to copy stats: %w", err)
		}
	}

	return nil
}

// GetProposalByID loads the proposal with draft, images, analyses, and stats when a pending weed still exists.
func (s *Service) GetProposalByID(ctx context.Context, proposalID int32) (*Proposal, error) {
	request, err := s.queries.GetCatalogProposalByID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	pendingWeed, err := s.queries.GetPendingWeedByProposalID(ctx, proposalID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.proposalFromRequest(request)
		}
		return nil, fmt.Errorf("failed to get pending weed: %w", err)
	}

	var (
		images            []database.PendingWeedImage
		analyses          []database.PendingWeedAnalysis
		dbStats           database.PendingWeedStat
		statsQueryErr     error
		pendingWeedLoadID = pendingWeed.ID
	)
	{
		g, gctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			var err error
			images, err = s.queries.GetPendingWeedImages(gctx, pendingWeedLoadID)
			if err != nil {
				return fmt.Errorf("failed to get images: %w", err)
			}
			return nil
		})
		g.Go(func() error {
			var err error
			analyses, err = s.queries.GetPendingWeedAnalyses(gctx, pendingWeedLoadID)
			if err != nil {
				return fmt.Errorf("failed to get analyses: %w", err)
			}
			return nil
		})
		g.Go(func() error {
			var err error
			dbStats, err = s.queries.GetPendingWeedStatsByPendingWeedID(gctx, pendingWeedLoadID)
			statsQueryErr = err
			return nil
		})
		if err := g.Wait(); err != nil {
			return nil, err
		}
	}

	var imageURLs []PendingWeedImageURL
	if len(images) == 0 {
		imageURLs = []PendingWeedImageURL{}
	} else {
		keys := make([]string, len(images))
		for i, img := range images {
			keys[i] = img.ImageKey
		}
		var presigned []string
		if s.objectStore != nil {
			var presignErr error
			presigned, presignErr = s.objectStore.GeneratePresignedURLs(ctx, keys, time.Hour*24*7)
			if presignErr != nil {
				presigned = make([]string, len(keys))
				for i, k := range keys {
					if k != "" {
						u, _ := s.objectStore.GeneratePresignedURL(ctx, k, time.Hour*24*7)
						presigned[i] = u
					}
				}
			}
		} else {
			presigned = make([]string, len(keys))
		}
		imageURLs = make([]PendingWeedImageURL, len(images))
		for i, img := range images {
			u := ""
			if i < len(presigned) {
				u = presigned[i]
			}
			imageURLs[i] = PendingWeedImageURL{
				ID:            img.ID,
				PendingWeedID: img.PendingWeedID,
				URL:           u,
				IsPrimary:     img.IsPrimary,
				ImageKey:      img.ImageKey,
			}
		}
	}

	analysisIDs := make([]string, 0, len(analyses))
	for _, a := range analyses {
		if _, err := uuid.Parse(a.AnalysisID); err == nil {
			analysisIDs = append(analysisIDs, a.AnalysisID)
		}
	}

	var weedStats *stats.WeedStatistics
	if statsQueryErr == nil {
		excludedObjects, _ := s.weedStatsSvc.DeserializeExcludedObjects(dbStats.ExcludedObjects)
		weedStats = weedStatisticsFromPendingWeedStat(dbStats, excludedObjects)
	}

	return s.proposalFromRequestWithDraft(
		request,
		&pendingWeed,
		imageURLs,
		analysisIDs,
		weedStats,
	), nil
}

func (s *Service) proposalFromRequest(request database.CatalogProposal) (*Proposal, error) {
	return s.proposalFromRequestWithDraft(request, nil, nil, nil, nil), nil
}

func (s *Service) proposalFromRequestWithDraft(
	request database.CatalogProposal,
	pendingWeed *database.PendingWeed,
	imageURLs []PendingWeedImageURL,
	analysisIDs []string,
	weedStats *stats.WeedStatistics,
) *Proposal {
	targetWeedID := request.TargetWeedID
	reviewedBy := request.ReviewedBy
	appliedBy := request.AppliedBy
	appliedWeedID := request.AppliedWeedID
	reviewedAt := gopt.Cond(request.ReviewedAt.Valid, request.ReviewedAt.Time).ToPointer()
	submittedAt := gopt.Cond(request.SubmittedAt.Valid, request.SubmittedAt.Time).ToPointer()
	appliedAt := gopt.Cond(request.AppliedAt.Valid, request.AppliedAt.Time).ToPointer()
	reviewNotes := request.ReviewNotes
	draft := PendingWeedDraft{}
	if pendingWeed != nil {
		draft = PendingWeedDraft{
			ID:           pendingWeed.ID,
			Name:         pendingWeed.Name,
			LatinName:    gopt.FromPtr(pendingWeed.LatinName).UnwrapOr(""),
			Description:  gopt.FromPtr(pendingWeed.Description).UnwrapOr(""),
			Length:       gopt.FromPtr(pendingWeed.Length).UnwrapOr(float32(0)),
			Width:        gopt.FromPtr(pendingWeed.Width).UnwrapOr(float32(0)),
			MainGroup:    gopt.FromPtr(pendingWeed.MainGroup).UnwrapOr(""),
			MainSubgroup: gopt.FromPtr(pendingWeed.MainSubgroup).UnwrapOr(""),
			Subgroup:     gopt.FromPtr(pendingWeed.Subgroup).UnwrapOr(""),
			IsQuarantine: pendingWeed.IsQuarantine,
			Harmfulness:  gopt.FromPtr(pendingWeed.Harmfulness).UnwrapOr(""),
		}
	}

	if imageURLs == nil {
		imageURLs = []PendingWeedImageURL{}
	}
	if analysisIDs == nil {
		analysisIDs = []string{}
	}

	return &Proposal{
		ID:            request.ID,
		Status:        request.Status,
		RequestBy:     request.RequestBy,
		TargetWeedID:  targetWeedID,
		ReviewedBy:    reviewedBy,
		ReviewedAt:    reviewedAt,
		ReviewNotes:   reviewNotes,
		SubmittedAt:   submittedAt,
		AppliedBy:     appliedBy,
		AppliedAt:     appliedAt,
		AppliedWeedID: appliedWeedID,
		CreatedAt:     request.CreatedAt.Time,
		UpdatedAt:     request.UpdatedAt.Time,
		Draft:         draft,
		Images:        imageURLs,
		Analyses:      analysisIDs,
		Statistics:    weedStats,
	}
}

// ListProposals returns a filtered, paginated list of proposals.
func (s *Service) ListProposals(
	ctx context.Context,
	params ListProposalsParams,
) (*dto.PaginatedResponse[ProposalListItem], error) {
	listParams := proposalListRowConverter.ListProposalsParamsToListDB(params)

	rows, err := s.queries.ListCatalogProposalsWithPendingWeed(ctx, listParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list proposals: %w", err)
	}

	items := make([]ProposalListItem, len(rows))
	for i, row := range rows {
		items[i] = proposalListRowConverter.ListRowToItem(row)
	}

	countParams := proposalListRowConverter.ListProposalsParamsToCountDB(params)

	total, err := s.queries.CountCatalogProposals(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count proposals: %w", err)
	}

	return &dto.PaginatedResponse[ProposalListItem]{
		Data:   items,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// UpdateProposalDraft updates the pending weed; only allowed in changes_requested state.
func (s *Service) UpdateProposalDraft(
	ctx context.Context,
	proposalID int32,
	params UpdateProposalDraftParams,
) (*Proposal, error) {
	request, err := s.queries.GetCatalogProposalByIDForUpdate(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	if request.Status != "changes_requested" {
		return nil, fmt.Errorf(
			"proposal can only be edited in changes_requested state, current: %s",
			request.Status,
		)
	}

	pendingWeed, err := s.queries.GetPendingWeedByProposalID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending weed: %w", err)
	}

	updateParams := database.UpdatePendingWeedParams{
		ID: pendingWeed.ID,
	}

	updateParams.Name = gopt.FromPtr(params.Name).UnwrapOr(pendingWeed.Name)
	updateParams.Description = params.Description
	if updateParams.Description == nil {
		updateParams.Description = pendingWeed.Description
	}
	updateParams.Harmfulness = params.Harmfulness
	if updateParams.Harmfulness == nil {
		updateParams.Harmfulness = pendingWeed.Harmfulness
	}
	updateParams.MainGroup = params.MainGroup
	if updateParams.MainGroup == nil {
		updateParams.MainGroup = pendingWeed.MainGroup
	}
	updateParams.MainSubgroup = params.MainSubgroup
	if updateParams.MainSubgroup == nil {
		updateParams.MainSubgroup = pendingWeed.MainSubgroup
	}
	updateParams.Subgroup = params.Subgroup
	if updateParams.Subgroup == nil {
		updateParams.Subgroup = pendingWeed.Subgroup
	}

	updateParams.IsQuarantine = pendingWeed.IsQuarantine
	updateParams.Length = pendingWeed.Length
	updateParams.Width = pendingWeed.Width

	if params.Statistics != nil {
		if params.Statistics.LMedian != 0 {
			l := params.Statistics.LMedian
			updateParams.Length = &l
		}
		if params.Statistics.WMedian != 0 {
			w := params.Statistics.WMedian
			updateParams.Width = &w
		}
	}

	if _, err := s.queries.UpdatePendingWeed(ctx, updateParams); err != nil {
		return nil, fmt.Errorf("failed to update pending weed: %w", err)
	}

	if params.AnalysisIDs != nil {
		if err := replacePendingWeedAnalyses(
			ctx,
			s.queries,
			pendingWeed.ID,
			*params.AnalysisIDs,
		); err != nil {
			return nil, fmt.Errorf("failed to replace analyses: %w", err)
		}
	}

	if params.Statistics != nil {
		if len(*params.ExcludedObjects) > 0 {
			params.Statistics.ExcludedObjects = *params.ExcludedObjects
		}
		excludedBytes, err := sonic.Marshal(params.Statistics.ExcludedObjects)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize excluded objects: %w", err)
		}

		_, err = s.queries.GetPendingWeedStatsByPendingWeedID(ctx, pendingWeed.ID)
		if err != nil {
			if _, err := s.queries.CreatePendingWeedStats(
				ctx,
				newCreatePendingWeedStatsParams(pendingWeed.ID, params.Statistics, excludedBytes),
			); err != nil {
				return nil, fmt.Errorf("failed to create stats: %w", err)
			}
		} else {
			if _, err := s.queries.UpdatePendingWeedStats(
				ctx,
				newUpdatePendingWeedStatsParams(pendingWeed.ID, params.Statistics, excludedBytes),
			); err != nil {
				return nil, fmt.Errorf("failed to update stats: %w", err)
			}
		}
	}

	return s.GetProposalByID(ctx, proposalID)
}

// SubmitProposal moves a proposal from changes_requested to submitted.
func (s *Service) SubmitProposal(ctx context.Context, proposalID int32) (*Proposal, error) {
	request, err := s.queries.GetCatalogProposalByIDForUpdate(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	if request.Status == "submitted" {
		return s.GetProposalByID(ctx, proposalID)
	}

	if request.Status != "changes_requested" {
		return nil, fmt.Errorf(
			"proposal can only be resubmitted from changes_requested state, current: %s",
			request.Status,
		)
	}

	updated, err := s.queries.UpdateCatalogProposalStatus(
		ctx,
		database.UpdateCatalogProposalStatusParams{
			ID:     proposalID,
			Status: "submitted",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit proposal: %w", err)
	}

	_ = updated
	return s.GetProposalByID(ctx, proposalID)
}

// RequestChanges moves a submitted proposal to changes_requested.
func (s *Service) RequestChanges(
	ctx context.Context,
	proposalID int32,
	reviewedBy int32,
	message string,
) (*Proposal, error) {
	request, err := s.queries.GetCatalogProposalByIDForUpdate(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	if request.Status == "changes_requested" {
		return s.GetProposalByID(ctx, proposalID)
	}

	if request.Status != "submitted" {
		return nil, fmt.Errorf(
			"can only request changes on submitted proposals, current: %s",
			request.Status,
		)
	}

	updated, err := s.queries.UpdateCatalogProposal(ctx, database.UpdateCatalogProposalParams{
		ID:          proposalID,
		Status:      "changes_requested",
		ReviewedBy:  &reviewedBy,
		ReviewNotes: &message,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request changes: %w", err)
	}

	_ = updated
	return s.GetProposalByID(ctx, proposalID)
}

// RejectProposal marks the proposal rejected and drops the pending weed.
func (s *Service) RejectProposal(
	ctx context.Context,
	proposalID int32,
	reviewedBy int32,
	reason string,
) (*Proposal, error) {
	request, err := s.queries.GetCatalogProposalByIDForUpdate(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	if request.Status == "rejected" {
		return s.GetProposalByID(ctx, proposalID)
	}

	if request.Status != "submitted" && request.Status != "changes_requested" {
		return nil, fmt.Errorf(
			"can only reject submitted or changes_requested proposals, current: %s",
			request.Status,
		)
	}

	updated, err := s.queries.UpdateCatalogProposal(ctx, database.UpdateCatalogProposalParams{
		ID:          proposalID,
		Status:      "rejected",
		ReviewedBy:  &reviewedBy,
		ReviewNotes: &reason,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to reject proposal: %w", err)
	}

	if pendingWeed, delErr := s.queries.GetPendingWeedByProposalID(ctx, proposalID); delErr == nil {
		_ = s.queries.DeletePendingWeed(ctx, pendingWeed.ID)
	}

	_ = updated
	return s.GetProposalByID(ctx, proposalID)
}

// ApplyProposal materializes the draft into a weed and marks the proposal applied.
func (s *Service) ApplyProposal(
	ctx context.Context,
	proposalID int32,
	appliedBy int32,
	note string,
) (*Proposal, error) {
	request, err := s.queries.GetCatalogProposalByIDForUpdate(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	if request.Status == "applied" {
		return s.GetProposalByID(ctx, proposalID)
	}

	if request.Status != "submitted" && request.Status != "changes_requested" {
		return nil, fmt.Errorf(
			"can only apply submitted or changes_requested proposals, current: %s",
			request.Status,
		)
	}

	pendingWeed, err := s.queries.GetPendingWeedByProposalID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending weed: %w", err)
	}

	var (
		images            []database.PendingWeedImage
		analyses          []database.PendingWeedAnalysis
		dbStats           database.PendingWeedStat
		statsErr          error
		pendingWeedLoadID = pendingWeed.ID
	)
	{
		g, gctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			var err error
			images, err = s.queries.GetPendingWeedImages(gctx, pendingWeedLoadID)
			if err != nil {
				return fmt.Errorf("failed to get pending images: %w", err)
			}
			return nil
		})
		g.Go(func() error {
			var err error
			analyses, err = s.queries.GetPendingWeedAnalyses(gctx, pendingWeedLoadID)
			if err != nil {
				return fmt.Errorf("failed to get pending analyses: %w", err)
			}
			return nil
		})
		g.Go(func() error {
			var err error
			dbStats, err = s.queries.GetPendingWeedStatsByPendingWeedID(gctx, pendingWeedLoadID)
			statsErr = err
			return nil
		})
		if err := g.Wait(); err != nil {
			return nil, err
		}
	}

	var analysisIDs []string
	for _, a := range analyses {
		if _, err := uuid.Parse(a.AnalysisID); err == nil {
			analysisIDs = append(analysisIDs, a.AnalysisID)
		}
	}

	var excludedObjects []int64
	var statistics *stats.WeedStatistics
	if statsErr == nil {
		excludedObjects, _ = s.weedStatsSvc.DeserializeExcludedObjects(dbStats.ExcludedObjects)
		statistics = weedStatisticsFromPendingWeedStat(dbStats, excludedObjects)
	}

	var weedID int32
	if request.TargetWeedID != nil {
		weedID = *request.TargetWeedID
		_, err = s.weedService.UpdateWeed(ctx, weedID, weed.SaveWeedParams{
			Name:            pendingWeed.Name,
			LatinName:       gopt.FromPtr(pendingWeed.LatinName).UnwrapOr(""),
			Description:     gopt.FromPtr(pendingWeed.Description).UnwrapOr(""),
			Length:          gopt.FromPtr(pendingWeed.Length).UnwrapOr(float32(0)),
			Width:           gopt.FromPtr(pendingWeed.Width).UnwrapOr(float32(0)),
			MainGroup:       gopt.FromPtr(pendingWeed.MainGroup).UnwrapOr(""),
			MainSubgroup:    gopt.FromPtr(pendingWeed.MainSubgroup).UnwrapOr(""),
			Subgroup:        gopt.FromPtr(pendingWeed.Subgroup).UnwrapOr(""),
			IsQuarantine:    pendingWeed.IsQuarantine,
			Harmfulness:     gopt.FromPtr(pendingWeed.Harmfulness).UnwrapOr(""),
			AnalysisIDs:     analysisIDs,
			Statistics:      statistics,
			ExcludedObjects: excludedObjects,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update weed: %w", err)
		}

		if err := s.weedImageSvc.DeleteAllWeedImages(ctx, weedID); err != nil {
			return nil, fmt.Errorf("failed to delete old images: %w", err)
		}
		if err := s.applyPendingImagesToWeed(ctx, weedID, images); err != nil {
			return nil, fmt.Errorf("failed to add images: %w", err)
		}
	} else {
		weed, err := s.weedService.CreateWeed(ctx, weed.SaveWeedParams{
			Name:            pendingWeed.Name,
			LatinName:       gopt.FromPtr(pendingWeed.LatinName).UnwrapOr(""),
			Description:     gopt.FromPtr(pendingWeed.Description).UnwrapOr(""),
			Length:          gopt.FromPtr(pendingWeed.Length).UnwrapOr(float32(0)),
			Width:           gopt.FromPtr(pendingWeed.Width).UnwrapOr(float32(0)),
			MainGroup:       gopt.FromPtr(pendingWeed.MainGroup).UnwrapOr(""),
			MainSubgroup:    gopt.FromPtr(pendingWeed.MainSubgroup).UnwrapOr(""),
			Subgroup:        gopt.FromPtr(pendingWeed.Subgroup).UnwrapOr(""),
			IsQuarantine:    pendingWeed.IsQuarantine,
			Harmfulness:     gopt.FromPtr(pendingWeed.Harmfulness).UnwrapOr(""),
			AnalysisIDs:     analysisIDs,
			Statistics:      statistics,
			ExcludedObjects: excludedObjects,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create weed: %w", err)
		}
		weedID = weed.ID

		if err := s.applyPendingImagesToWeed(ctx, weedID, images); err != nil {
			return nil, fmt.Errorf("failed to add images: %w", err)
		}
	}

	var reviewNotes *string
	if note != "" {
		reviewNotes = &note
	}

	_, err = s.queries.MarkCatalogProposalApplied(ctx, database.MarkCatalogProposalAppliedParams{
		ID:            proposalID,
		ReviewedBy:    &appliedBy,
		ReviewNotes:   reviewNotes,
		AppliedBy:     &appliedBy,
		AppliedWeedID: &weedID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to mark proposal as applied: %w", err)
	}

	if err := s.queries.DeletePendingWeed(ctx, pendingWeed.ID); err != nil {
		return nil, fmt.Errorf("failed to delete pending weed: %w", err)
	}

	return s.GetProposalByID(ctx, proposalID)
}

// CancelProposal marks a submitted proposal cancelled and drops the pending weed.
func (s *Service) CancelProposal(ctx context.Context, proposalID int32) (*Proposal, error) {
	request, err := s.queries.GetCatalogProposalByIDForUpdate(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	if request.Status == "cancelled" {
		return s.GetProposalByID(ctx, proposalID)
	}

	if request.Status != "submitted" {
		return nil, fmt.Errorf("can only cancel submitted proposals, current: %s", request.Status)
	}

	updated, err := s.queries.UpdateCatalogProposalStatus(
		ctx,
		database.UpdateCatalogProposalStatusParams{
			ID:     proposalID,
			Status: "cancelled",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel proposal: %w", err)
	}

	if pendingWeed, delErr := s.queries.GetPendingWeedByProposalID(ctx, proposalID); delErr == nil {
		_ = s.queries.DeletePendingWeed(ctx, pendingWeed.ID)
	}

	_ = updated
	return s.GetProposalByID(ctx, proposalID)
}

// UploadProposalImage stores an object and links it to the proposal draft.
func (s *Service) UploadProposalImage(
	ctx context.Context,
	proposalID int32,
	fileContent io.Reader,
	fileSize int64,
	contentType string,
	originalFilename string,
) (PendingWeedImageURL, error) {
	pendingWeed, err := s.queries.GetPendingWeedByProposalID(ctx, proposalID)
	if err != nil {
		return PendingWeedImageURL{}, fmt.Errorf(
			"failed to load pending weed for proposal %d: %w",
			proposalID,
			err,
		)
	}

	request, err := s.queries.GetCatalogProposalByID(ctx, proposalID)
	if err != nil {
		return PendingWeedImageURL{}, fmt.Errorf("failed to get proposal: %w", err)
	}
	if request.Status != "submitted" && request.Status != "changes_requested" {
		return PendingWeedImageURL{}, errors.New(
			"images can only be added to submitted or changes_requested proposals",
		)
	}

	existing, err := s.queries.GetPendingWeedImages(ctx, pendingWeed.ID)
	if err != nil {
		return PendingWeedImageURL{}, fmt.Errorf("failed to load existing images: %w", err)
	}
	hasPrimary := false
	for _, img := range existing {
		if img.IsPrimary {
			hasPrimary = true
			break
		}
	}

	extension := filepath.Ext(originalFilename)
	objectKey := fmt.Sprintf(
		"pending-weed-images/%d/%s%s",
		pendingWeed.ID,
		uuid.New().String(),
		extension,
	)
	if err := s.objectStore.UploadObject(
		ctx,
		objectKey,
		fileContent,
		fileSize,
		contentType,
	); err != nil {
		return PendingWeedImageURL{}, fmt.Errorf("failed to upload image: %w", err)
	}

	shouldPrimary := !hasPrimary
	dbImage, err := s.queries.AddPendingWeedImage(ctx, database.AddPendingWeedImageParams{
		PendingWeedID: pendingWeed.ID,
		ImageKey:      objectKey,
		IsPrimary:     shouldPrimary,
	})
	if err != nil {
		_ = s.objectStore.DeleteObject(ctx, objectKey)
		return PendingWeedImageURL{}, fmt.Errorf("failed to add image to database: %w", err)
	}

	url, err := s.objectStore.GeneratePresignedURL(ctx, dbImage.ImageKey, time.Hour*24*7)
	if err != nil {
		url = ""
	}
	return PendingWeedImageURL{
		ID:            dbImage.ID,
		PendingWeedID: dbImage.PendingWeedID,
		URL:           url,
		IsPrimary:     dbImage.IsPrimary,
		ImageKey:      dbImage.ImageKey,
	}, nil
}

// DeleteProposalImage removes a pending image row and the backing object.
func (s *Service) DeleteProposalImage(ctx context.Context, proposalID int32, imageID int32) error {
	request, err := s.queries.GetCatalogProposalByID(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %w", err)
	}
	if request.Status != "submitted" && request.Status != "changes_requested" {
		return errors.New(
			"images can only be deleted from submitted or changes_requested proposals",
		)
	}

	img, err := s.queries.GetPendingWeedImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	if err := s.queries.DeletePendingWeedImage(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	_ = s.objectStore.DeleteObject(ctx, img.ImageKey)
	return nil
}

func (s *Service) applyPendingImagesToWeed(
	ctx context.Context,
	weedID int32,
	images []database.PendingWeedImage,
) error {
	if len(images) == 0 {
		return nil
	}
	imageKeys := make([]string, len(images))
	isPrimary := make([]bool, len(images))
	for i, img := range images {
		imageKeys[i] = img.ImageKey
		isPrimary[i] = img.IsPrimary
	}
	return s.weedImageSvc.BulkAddWeedImages(ctx, weedID, imageKeys, isPrimary)
}
