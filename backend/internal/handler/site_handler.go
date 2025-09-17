package handler

import (
	"strconv"

	"github.com/engramiq/engramiq-backend/internal/repository"
	"github.com/gofiber/fiber/v2"
)

type SiteHandler struct {
	siteRepo repository.SiteRepository
}

func NewSiteHandler(siteRepo repository.SiteRepository) *SiteHandler {
	return &SiteHandler{
		siteRepo: siteRepo,
	}
}

func (h *SiteHandler) ListSites(c *fiber.Ctx) error {
	// Parse pagination parameters
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	sites, total, err := h.siteRepo.GetSites(page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sites",
		})
	}

	return c.JSON(fiber.Map{
		"data": sites,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *SiteHandler) GetSite(c *fiber.Ctx) error {
	siteID := c.Params("id")
	if siteID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Site ID is required",
		})
	}

	site, err := h.siteRepo.GetSite(siteID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Site not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": site,
	})
}