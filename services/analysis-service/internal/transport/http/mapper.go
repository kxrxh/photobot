package http

import (
	"csort.ru/analysis-service/internal/analysis"
	"csort.ru/analysis-service/internal/dto"
	"csort.ru/analysis-service/internal/objects"
	"csort.ru/analysis-service/internal/requests"
)

func AnalysisToResponse(a analysis.Analysis) dto.AnalysisResponse {
	return dto.AnalysisResponse{
		ID:             a.ID,
		DateTime:       a.DateTime,
		Product:        a.Product,
		UserID:         a.UserID,
		Source:         a.Source,
		BotMessage:     a.BotMessage,
		FilesSource:    a.FilesSource,
		FilesOutput:    a.FilesOutput,
		ScaleMmPixel:   a.ScaleMmPixel,
		AnalysisParams: a.AnalysisParams,
	}
}

func AnalysisListItemToResponse(a analysis.AnalysisListItem) dto.AnalysisResponse {
	return dto.AnalysisResponse{
		ID:           a.ID,
		DateTime:     a.DateTime,
		Product:      a.Product,
		UserID:       a.UserID,
		Source:       a.Source,
		BotMessage:   a.BotMessage,
		FilesSource:  a.FilesSource,
		FilesOutput:  a.FilesOutput,
		ScaleMmPixel: a.ScaleMmPixel,
	}
}

func AnalysisWithObjectsToResponse(a analysis.AnalysisWithObjects) dto.AnalysisWithObjectsResponse {
	objs := make([]dto.ObjectResponse, len(a.Objects))
	for i := range a.Objects {
		objs[i] = ObjectToResponse(a.Objects[i])
	}
	return dto.AnalysisWithObjectsResponse{
		AnalysisResponse: AnalysisToResponse(a.Analysis),
		Objects:          objs,
	}
}

func ObjectToResponse(o objects.Object) dto.ObjectResponse {
	return dto.ObjectResponse{
		ID:       o.ID,
		IDImage:  o.IDImage,
		Class:    o.Class,
		Geometry: o.Geometry,
		File:     o.File,
		MH:       o.MH,
		MS:       o.MS,
		MV:       o.MV,
		MR:       o.MR,
		MG:       o.MG,
		MB:       o.MB,
		LAvg:     o.LAvg,
		WAvg:     o.WAvg,
		BrtAvg:   o.BrtAvg,
		RAvg:     o.RAvg,
		GAvg:     o.GAvg,
		BAvg:     o.BAvg,
		HAvg:     o.HAvg,
		SAvg:     o.SAvg,
		VAvg:     o.VAvg,
		H:        o.H,
		S:        o.S,
		V:        o.V,
		HM:       o.HM,
		SM:       o.SM,
		VM:       o.VM,
		RM:       o.RM,
		GM:       o.GM,
		BM:       o.BM,
		BrtM:     o.BrtM,
		WM:       o.WM,
		LM:       o.LM,
		L:        o.L,
		W:        o.W,
		LW:       o.LW,
		Pr:       o.Pr,
		R:        o.R,
		G:        o.G,
		B:        o.B,
		Sq:       o.Sq,
		Brt:      o.Brt,
		MinH:     o.MinH,
		MinS:     o.MinS,
		MinV:     o.MinV,
		MaxH:     o.MaxH,
		MaxS:     o.MaxS,
		MaxV:     o.MaxV,
		ColorRhs: o.ColorRhs,
		Solid:    o.Solid,
		Entropy:  o.Entropy,
		SqSqcrl:  o.SqSqcrl,
		Hu1:      o.Hu1,
		Hu2:      o.Hu2,
		Hu3:      o.Hu3,
		Hu4:      o.Hu4,
		Hu5:      o.Hu5,
		Hu6:      o.Hu6,
		Mass1000: o.Mass1000,
		Mass:     o.Mass,
	}
}

func ObjectMetadataToResponse(o *objects.ObjectMetadata) dto.ObjectResponse {
	if o == nil {
		return dto.ObjectResponse{}
	}
	return dto.ObjectResponse{
		ID:       o.ID,
		IDImage:  o.IDImage,
		Class:    o.Class,
		Geometry: o.Geometry,
		File:     nil,
		MH:       o.MH,
		MS:       o.MS,
		MV:       o.MV,
		MR:       o.MR,
		MG:       o.MG,
		MB:       o.MB,
		LAvg:     o.LAvg,
		WAvg:     o.WAvg,
		BrtAvg:   o.BrtAvg,
		RAvg:     o.RAvg,
		GAvg:     o.GAvg,
		BAvg:     o.BAvg,
		HAvg:     o.HAvg,
		SAvg:     o.SAvg,
		VAvg:     o.VAvg,
		H:        o.H,
		S:        o.S,
		V:        o.V,
		HM:       o.HM,
		SM:       o.SM,
		VM:       o.VM,
		RM:       o.RM,
		GM:       o.GM,
		BM:       o.BM,
		BrtM:     o.BrtM,
		WM:       o.WM,
		LM:       o.LM,
		L:        o.L,
		W:        o.W,
		LW:       o.LW,
		Pr:       o.Pr,
		R:        o.R,
		G:        o.G,
		B:        o.B,
		Sq:       o.Sq,
		Brt:      o.Brt,
		MinH:     o.MinH,
		MinS:     o.MinS,
		MinV:     o.MinV,
		MaxH:     o.MaxH,
		MaxS:     o.MaxS,
		MaxV:     o.MaxV,
		ColorRhs: o.ColorRhs,
		Solid:    o.Solid,
		Entropy:  o.Entropy,
		SqSqcrl:  o.SqSqcrl,
		Hu1:      o.Hu1,
		Hu2:      o.Hu2,
		Hu3:      o.Hu3,
		Hu4:      o.Hu4,
		Hu5:      o.Hu5,
		Hu6:      o.Hu6,
		Mass1000: o.Mass1000,
		Mass:     o.Mass,
	}
}

func RequestToResponse(r requests.Request) dto.RequestResponse {
	return dto.RequestResponse{
		ID:           r.ID,
		UserID:       r.UserID,
		Platform:     r.Platform,
		Product:      r.Product,
		Status:       dto.RequestStatus(r.Status),
		Year:         r.Year,
		MassLiter:    r.MassLiter,
		Location:     r.Location,
		Mass1000:     r.Mass1000,
		Mass:         r.Mass,
		Images:       r.Images,
		TempID:       r.TempID,
		ErrorMessage: r.ErrorMessage,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func RequestsToGetRequestsResponse(reqs []requests.Request) dto.GetRequestsResponse {
	out := make([]dto.RequestResponse, len(reqs))
	for i := range reqs {
		out[i] = RequestToResponse(reqs[i])
	}
	return dto.GetRequestsResponse{Requests: out, Total: len(reqs)}
}

func GetAnalysesQueryRequestToListParams(q dto.GetAnalysesQueryRequest) analysis.ListParams {
	return analysis.ListParams{
		Limit:     q.Limit,
		Offset:    q.Offset,
		Product:   q.Product,
		IDFilter:  q.ID,
		SortBy:    q.SortBy,
		SortOrder: q.SortOrder,
	}
}

func GetRequestsQueryRequestToParams(
	q dto.GetRequestsQueryRequest,
	userID string,
	platform *string,
) requests.GetRequestsRequest {
	var status *requests.RequestStatus
	if q.Status != nil {
		s := requests.RequestStatus(*q.Status)
		status = &s
	}
	limit := q.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}
	return requests.GetRequestsRequest{
		UserID:   userID,
		Platform: platform,
		Status:   status,
		Limit:    limit,
		Offset:   offset,
	}
}
