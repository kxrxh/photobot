package http

import (
	"csort.ru/classification-service/internal/classification"
	"csort.ru/classification-service/internal/correlation"
	"csort.ru/classification-service/internal/dto"
	"csort.ru/classification-service/internal/markup"
	"csort.ru/classification-service/internal/params"
	"csort.ru/classification-service/internal/product"
)

func SaveProductRequestToDomain(req dto.SaveProductRequest) product.SaveProduct {
	return product.SaveProduct{Name: req.Name}
}

func ProductToResponse(p product.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func ProductPtrToResponse(p *product.Product) dto.ProductResponse {
	if p == nil {
		return dto.ProductResponse{}
	}
	return ProductToResponse(*p)
}

func ProductsToResponse(products []product.Product) []dto.ProductResponse {
	out := make([]dto.ProductResponse, len(products))
	for i := range products {
		out[i] = ProductToResponse(products[i])
	}
	return out
}

func SaveParamRequestToDomain(req dto.SaveParamRequest) params.SaveClassificationParam {
	return params.SaveClassificationParam{Name: req.Name}
}

func ParamToResponse(p params.ClassificationParam) dto.ParamResponse {
	return dto.ParamResponse{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func ParamPtrToResponse(p *params.ClassificationParam) dto.ParamResponse {
	if p == nil {
		return dto.ParamResponse{}
	}
	return ParamToResponse(*p)
}

func ParamsToResponse(items []params.ClassificationParam) []dto.ParamResponse {
	out := make([]dto.ParamResponse, len(items))
	for i := range items {
		out[i] = ParamToResponse(items[i])
	}
	return out
}

func OwnershipTransferRequestToParams(
	req dto.OwnershipTransferRequest,
) (fromUserID, toUserID int32) {
	return req.FromUserID, req.ToUserID
}

func SaveMarkupRequestToDomain(
	req dto.SaveMarkupRequest,
	createdBy int32,
) markup.SaveMarkupRequest {
	fractions := make([]markup.SaveMarkupFraction, len(req.Fractions))
	for i, f := range req.Fractions {
		fractions[i] = markup.SaveMarkupFraction{
			ObjectIDs: f.ObjectIDs,
			Name:      f.Name,
		}
	}
	return markup.SaveMarkupRequest{
		Name:        req.Name,
		CreatedBy:   createdBy,
		Fractions:   fractions,
		AnalysesIDs: req.AnalysesIDs,
	}
}

func MarkupToResponse(m markup.Markup) dto.MarkupResponse {
	fractions := make([]dto.MarkupFractionResponse, len(m.Fractions))
	for i, f := range m.Fractions {
		fractions[i] = dto.MarkupFractionResponse{
			ID:        f.ID,
			Name:      f.Name,
			ObjectIDs: f.ObjectIDs,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		}
	}
	return dto.MarkupResponse{
		ID:          m.ID,
		Name:        m.Name,
		CreatedBy:   m.CreatedBy,
		Fractions:   fractions,
		AnalysesIDs: m.AnalysesIDs,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func MarkupsToResponse(markups []markup.Markup) []dto.MarkupResponse {
	out := make([]dto.MarkupResponse, len(markups))
	for i := range markups {
		out[i] = MarkupToResponse(markups[i])
	}
	return out
}

func MarkupPtrToResponse(m *markup.Markup) dto.MarkupResponse {
	if m == nil {
		return dto.MarkupResponse{}
	}
	return MarkupToResponse(*m)
}

func productRefToDomain(p dto.ProductRef) product.Product {
	return product.Product{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func productToRef(p product.Product) dto.ProductRef {
	return dto.ProductRef{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func paramRefToDomain(p dto.ParamRef) classification.Param {
	return classification.Param{
		ID:       p.ID,
		Name:     p.Name,
		Operator: p.Operator,
		Value:    p.Value,
	}
}

func paramToRef(p classification.Param) dto.ParamRef {
	return dto.ParamRef{
		ID:       p.ID,
		Name:     p.Name,
		Operator: p.Operator,
		Value:    p.Value,
	}
}

func conditionRefToDomain(c dto.ConditionRef) classification.Condition {
	params := make([]classification.Param, len(c.Params))
	for i, p := range c.Params {
		params[i] = paramRefToDomain(p)
	}
	return classification.Condition{
		ID:         c.ID,
		Name:       c.Name,
		Operator:   c.Operator,
		Connection: c.Connection,
		OrderIndex: c.OrderIndex,
		Params:     params,
	}
}

func conditionToRef(c classification.Condition) dto.ConditionRef {
	params := make([]dto.ParamRef, len(c.Params))
	for i, p := range c.Params {
		params[i] = paramToRef(p)
	}
	return dto.ConditionRef{
		ID:         c.ID,
		Name:       c.Name,
		Operator:   c.Operator,
		Connection: c.Connection,
		OrderIndex: c.OrderIndex,
		Params:     params,
	}
}

func fractionRefToDomain(f dto.FractionRef) classification.Fraction {
	conditions := make([]classification.Condition, len(f.Conditions))
	for i, c := range f.Conditions {
		conditions[i] = conditionRefToDomain(c)
	}
	return classification.Fraction{
		ID:         f.ID,
		Name:       f.Name,
		OrderIndex: f.OrderIndex,
		Conditions: conditions,
	}
}

func fractionToRef(f classification.Fraction) dto.FractionRef {
	conditions := make([]dto.ConditionRef, len(f.Conditions))
	for i, c := range f.Conditions {
		conditions[i] = conditionToRef(c)
	}
	return dto.FractionRef{
		ID:         f.ID,
		Name:       f.Name,
		OrderIndex: f.OrderIndex,
		Conditions: conditions,
	}
}

func SaveClassificationRequestToDomain(
	req dto.SaveClassificationRequest,
) classification.SaveCompleteClassificationRequest {
	fractions := make([]classification.Fraction, len(req.Fractions))
	for i, f := range req.Fractions {
		fractions[i] = fractionRefToDomain(f)
	}
	return classification.SaveCompleteClassificationRequest{
		Product:   productRefToDomain(req.Product),
		Fractions: fractions,
		Name:      req.Name,
		IsPublic:  req.IsPublic,
	}
}

func ClassificationToResponse(c classification.Classification) dto.ClassificationResponse {
	return dto.ClassificationResponse{
		ID:        c.ID,
		Name:      c.Name,
		CreatedBy: c.CreatedBy,
		Product:   productToRef(c.Product),
		IsPublic:  c.IsPublic,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func CompleteClassificationToResponse(
	c classification.CompleteClassification,
) dto.CompleteClassificationResponse {
	fractions := make([]dto.FractionRef, len(c.Fractions))
	for i, f := range c.Fractions {
		fractions[i] = fractionToRef(f)
	}
	return dto.CompleteClassificationResponse{
		Classification: ClassificationToResponse(c.Classification),
		Fractions:      fractions,
	}
}

func ClassificationPtrToResponse(c *classification.Classification) *dto.ClassificationResponse {
	if c == nil {
		return nil
	}
	resp := ClassificationToResponse(*c)
	return &resp
}

func ClassificationsToResponse(
	classifications []classification.Classification,
) []dto.ClassificationResponse {
	out := make([]dto.ClassificationResponse, len(classifications))
	for i := range classifications {
		out[i] = ClassificationToResponse(classifications[i])
	}
	return out
}

func CorrelationRequestToDomain(req dto.CorrelationRequest) correlation.CorrelationRequest {
	fractions := make([]correlation.ObjectGroup, len(req.Fractions))
	for i, f := range req.Fractions {
		fractions[i] = correlation.ObjectGroup{
			Name:      f.Name,
			ObjectIDs: f.ObjectIDs,
		}
	}
	return correlation.CorrelationRequest{
		Fractions:       fractions,
		ParameterGroups: req.ParameterGroups,
	}
}

func CorrelationWithTestToResponse(
	c correlation.CorrelationWithTest,
) dto.CorrelationWithTestResponse {
	resp := dto.CorrelationWithTestResponse{
		Name:       c.Name,
		Conditions: make([]dto.CorrelationConditionRef, len(c.Conditions)),
	}
	for i, cond := range c.Conditions {
		resp.Conditions[i] = dto.CorrelationConditionRef{
			Attribute: cond.Attribute,
			Operator:  cond.Operator,
			Value:     cond.Value,
		}
	}
	if c.TestResults != nil {
		resp.TestResults = &dto.ConditionTestResultRef{
			TruePositives:  c.TestResults.TruePositives,
			FalsePositives: c.TestResults.FalsePositives,
			TrueNegatives:  c.TestResults.TrueNegatives,
			FalseNegatives: c.TestResults.FalseNegatives,
			Precision:      c.TestResults.Precision,
			Recall:         c.TestResults.Recall,
			Accuracy:       c.TestResults.Accuracy,
			F1Score:        c.TestResults.F1Score,
		}
	}
	return resp
}

func CorrelationsWithTestToResponse(
	items []correlation.CorrelationWithTest,
) []dto.CorrelationWithTestResponse {
	out := make([]dto.CorrelationWithTestResponse, len(items))
	for i := range items {
		out[i] = CorrelationWithTestToResponse(items[i])
	}
	return out
}
