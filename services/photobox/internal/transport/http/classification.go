package http

import (
	"csort.ru/coffeebot/internal/classification"
	"csort.ru/coffeebot/internal/dto"
	"csort.ru/coffeebot/internal/transport/response"
	"github.com/gofiber/fiber/v3"
)

type ClassificationHandler struct{}

func NewClassificationHandler() *ClassificationHandler { return &ClassificationHandler{} }

func (h *ClassificationHandler) GetClassifications(c fiber.Ctx) error {
	hierarchy := classification.HierarchyMap()
	if hierarchy == nil {
		hierarchy = map[string]map[string][]string{}
	}
	return response.OK(c, dto.ClassificationsResponse{
		MainGroups:    classification.MainGroupMap(),
		MainSubgroups: classification.MainSubgroupMap(),
		Subgroups:     classification.SubgroupMap(),
		Hierarchy:     hierarchy,
	})
}
