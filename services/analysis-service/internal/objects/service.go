package objects

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/dto"
	"csort.ru/analysis-service/internal/identity"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/repository/kalibr"
	"csort.ru/analysis-service/internal/repository/requests"
	"csort.ru/analysis-service/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kxrxh/gopt"
	"github.com/rs/zerolog"
)

type Service struct {
	repoKalibr   *kalibr.Queries
	repoRequests *requests.Queries
	tempStorage  *storage.Client
	logger       zerolog.Logger
}

func NewObjectsService(
	repoKalibr *kalibr.Queries,
	repoRequests *requests.Queries,
	tempStorage *storage.Client,
) *Service {
	return &Service{
		repoKalibr:   repoKalibr,
		repoRequests: repoRequests,
		tempStorage:  tempStorage,
		logger:       logger.GetLogger("objects.service"),
	}
}

func (s *Service) tempObjectRefs(ctx context.Context, tempID string) ([]*dto.ObjectRef, error) {
	if s.tempStorage == nil {
		return nil, errors.New("temp storage client is not configured")
	}
	refs, err := s.tempStorage.ListTempAnalysisObjects(ctx, tempID, 1*time.Hour)
	if err != nil {
		return nil, err
	}
	res := make([]*dto.ObjectRef, len(refs))
	for i := range refs {
		res[i] = &dto.ObjectRef{
			ID:           refs[i].ID,
			PresignedURL: refs[i].PresignedURL,
		}
	}
	return res, nil
}

func (s *Service) GetByIDs(
	ctx context.Context,
	analysisID string,
	objectIndices []int32,
) ([]*ObjectMetadata, error) {
	s.logger.Debug().
		Str("analysisID", analysisID).
		Int("indices_count", len(objectIndices)).
		Msg("getting objects by analysis ID and indices")

	id, err := uuid.Parse(analysisID)
	if err != nil {
		s.logger.Warn().Str("analysisID", analysisID).Msg("invalid analysis ID format")
		return nil, apierrors.BadRequest("invalid analysis ID")
	}

	repoAnalysis, err := s.repoKalibr.GetAnalysisByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.NotFound("analysis not found")
		}
		s.logger.Error().Err(err).Str("analysisID", analysisID).Msg("failed to get analysis")
		return nil, err
	}

	objs, err := objectsFromJSONB(repoAnalysis.Objects)
	if err != nil {
		s.logger.Error().Err(err).Str("analysisID", analysisID).Msg("failed to parse objects JSONB")
		return nil, err
	}

	result := make([]*ObjectMetadata, 0, len(objectIndices))
	for _, idx := range objectIndices {
		if idx < 0 || int(idx) >= len(objs) {
			continue
		}
		o := objs[idx]
		result = append(result, &ObjectMetadata{
			ID:              o.ID,
			IDImage:         o.IDImage,
			Class:           o.Class,
			Geometry:        o.Geometry,
			MH:              o.MH,
			MS:              o.MS,
			MV:              o.MV,
			MR:              o.MR,
			MG:              o.MG,
			MB:              o.MB,
			LAvg:            o.LAvg,
			WAvg:            o.WAvg,
			BrtAvg:          o.BrtAvg,
			RAvg:            o.RAvg,
			GAvg:            o.GAvg,
			BAvg:            o.BAvg,
			HAvg:            o.HAvg,
			SAvg:            o.SAvg,
			VAvg:            o.VAvg,
			H:               o.H,
			S:               o.S,
			V:               o.V,
			HM:              o.HM,
			SM:              o.SM,
			VM:              o.VM,
			RM:              o.RM,
			GM:              o.GM,
			BM:              o.BM,
			BrtM:            o.BrtM,
			WM:              o.WM,
			LM:              o.LM,
			L:               o.L,
			W:               o.W,
			LW:              o.LW,
			Pr:              o.Pr,
			Sq:              o.Sq,
			Brt:             o.Brt,
			R:               o.R,
			G:               o.G,
			B:               o.B,
			Solid:           o.Solid,
			MinH:            o.MinH,
			MinS:            o.MinS,
			MinV:            o.MinV,
			MaxH:            o.MaxH,
			MaxS:            o.MaxS,
			MaxV:            o.MaxV,
			Entropy:         o.Entropy,
			ColorRhs:        o.ColorRhs,
			SqSqcrl:         o.SqSqcrl,
			Hu1:             o.Hu1,
			Hu2:             o.Hu2,
			Hu3:             o.Hu3,
			Hu4:             o.Hu4,
			Hu5:             o.Hu5,
			Hu6:             o.Hu6,
			Mass1000:        o.Mass1000,
			Mass:            o.Mass,
			NgtdmCoarseness: o.NgtdmCoarseness,
			NgtdmContrast:   o.NgtdmContrast,
			NgtdmBusyness:   o.NgtdmBusyness,
			NgtdmComplexity: o.NgtdmComplexity,
			NgtdmStrngth:    o.NgtdmStrngth,
			Corners:         o.Corners,
		})
	}
	return result, nil
}

func objectsFromJSONB(data []byte) ([]Object, error) {
	if len(data) == 0 {
		return []Object{}, nil
	}
	var objs []Object
	if err := json.Unmarshal(data, &objs); err != nil {
		return nil, err
	}
	for i := range objs {
		objs[i].ID = int32(i)
	}
	return objs, nil
}

func (s *Service) GetByRequestID(
	ctx context.Context,
	pairs []identity.UserPlatformPair,
	requestID string,
) ([]*dto.ObjectRef, error) {
	s.logger.Debug().Str("requestID", requestID).Msg("getting objects by requestID")
	if requestID == "" {
		return nil, apierrors.BadRequest("requestID is empty")
	}
	if len(pairs) == 0 {
		return nil, apierrors.Unauthorized("at least one user platform pair is required")
	}

	p1 := pairs[0]
	p2 := p1
	if len(pairs) > 1 {
		p2 = pairs[1]
	}

	request, err := s.repoRequests.GetRequestByIDAndUserPairs(
		ctx,
		requests.GetRequestByIDAndUserPairsParams{
			ID:         requestID,
			UserID:     p1.UserID,
			Platform:   p1.Platform,
			UserID_2:   p2.UserID,
			Platform_2: p2.Platform,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Str("requestID", requestID).Msg("request not found")
			return nil, apierrors.NotFound("request not found")
		}
		s.logger.Error().
			Err(err).
			Str("requestID", requestID).
			Msg("failed to get request by requestID")
		return nil, apierrors.InternalWrap(err, "failed to get request by requestID")
	}

	tempIDOpt := gopt.FromPtr(request.TempID)
	if tempIDOpt.IsNone() {
		s.logger.Warn().Str("requestID", requestID).Msg("request is not ready (tempID missing)")
		return nil, apierrors.BadRequest("request is not ready (tempID missing)")
	}
	tempID := tempIDOpt.Unwrap()
	tempObjs, err := s.tempObjectRefs(ctx, tempID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("requestID", requestID).
			Str("tempID", tempID).
			Msg("failed to get temp objects from storage")
		return nil, apierrors.InternalWrap(err, "failed to get temp objects from storage")
	}
	return tempObjs, nil
}
