package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/entity"
	"github.com/kimbasn/printly/internal/service"
)

type PrintCenterController interface {
	CreatePrintCenter(ctx *gin.Context)
	GetPrintCenterByID(ctx *gin.Context)
	GetAllPublicPrintCenters(ctx *gin.Context)
	GetPendingPrintCenters(ctx *gin.Context)
	GetAllPrintCenters(ctx *gin.Context)
	UpdatePrintCenter(ctx *gin.Context)
	UpdatePrintCenterStatus(ctx *gin.Context)
	DeletePrintCenter(ctx *gin.Context)
}

type printCenterController struct {
	service  service.PrintCenterService
	validate *validator.Validate
}

func NewPrintCenterController(service service.PrintCenterService, validate *validator.Validate) PrintCenterController {
	return &printCenterController{
		service:  service,
		validate: validate,
	}
}

// CreatePrintCenter godoc
// @Summary      Register a new print center
// @Description  Registers a new print center, which will be pending approval. Requires authentication.
// @Tags         Print Centers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        center  body      dto.CreatePrintCenterRequest  true  "Print Center to create"
// @Success      201     {object}  entity.PrintCenter
// @Failure      400     {object}  dto.ErrorResponse "Invalid input"
// @Failure      500     {object}  dto.ErrorResponse "Failed to register print center"
// @Router       /centers [post]
func (c *printCenterController) CreatePrintCenter(ctx *gin.Context) {
	var req dto.CreatePrintCenterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.validate.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ownerUID, exists := ctx.Get("userUID")
	if !exists {
		// This should not happen if the AuthenticationMiddleware is applied correctly.
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user UID not found in context"})
		return
	}

	center := &entity.PrintCenter{
		Name:         req.Name,
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		Location:     req.Location,
		Services:     req.Services,
		WorkingHours: req.WorkingHours,
		OwnerUID:     ownerUID.(string),
	}

	created, err := c.service.Register(center)
	if err != nil {
		HandleServiceError(ctx, err, "failed to register print center")
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

// GetPrintCenterByID godoc
// @Summary      Get a print center by ID
// @Description  Retrieves a single print center by its ID.
// @Tags         Print Centers
// @Produce      json
// @Param        id   path      string       true  "Print Center ID"
// @Success      200  {object}  entity.PrintCenter
// @Failure      404  {object}  dto.ErrorResponse "Print center not found"
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch print center"
// @Router       /centers/{id} [get]
func (c *printCenterController) GetPrintCenterByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 8)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	center, err := c.service.GetByID(uint(id))
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch print center")
		return
	}
	ctx.JSON(http.StatusOK, center)
}

// GetAllPublicPrintCenters godoc
// @Summary      Get all public print centers
// @Description  Retrieves a list of all approved print centers.
// @Tags         Print Centers
// @Produce      json
// @Success      200  {array}   entity.PrintCenter
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch print centers"
// @Router       /centers [get]
func (c *printCenterController) GetAllPublicPrintCenters(ctx *gin.Context) {
	centers, err := c.service.GetAllPublic()
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch print centers")
		return
	}
	ctx.JSON(http.StatusOK, centers)
}

// UpdatePrintCenter godoc
// @Summary      Update a print center's profile
// @Description  Updates a print center's information. Requires owner or admin role.
// @Tags         Print Centers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                      true  "Print Center ID"
// @Param        center  body      dto.UpdatePrintCenterRequest  true  "Print Center data to update"
// @Success      200     {object}  dto.SuccessResponse "Print center updated"
// @Failure      400     {object}  dto.ErrorResponse   "Invalid input"
// @Failure      404     {object}  dto.ErrorResponse   "Print center not found"
// @Failure      500     {object}  dto.ErrorResponse   "Failed to update print center"
// @Router       /centers/{id} [put]
func (c *printCenterController) UpdatePrintCenter(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 8)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.UpdatePrintCenterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	if err := c.validate.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.PhoneNumber != nil {
		updates["phone_number"] = *req.PhoneNumber
	}
	// Note: Updating nested structs like Location, Services, and WorkingHours
	// via a generic map is complex and better handled with more specific service logic.
	// This implementation focuses on updating top-level fields.

	if err := c.service.Update(uint(id), updates); err != nil {
		HandleServiceError(ctx, err, "failed to update print center")
		return
	}
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "print center updated"})
}

// DeletePrintCenter godoc
// @Summary      Delete a print center
// @Description  Deletes a print center. Requires admin role.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Print Center ID"
// @Success      200  {object}  dto.SuccessResponse "Print center deleted successfully"
// @Failure      404  {object}  dto.ErrorResponse   "Print center not found"
// @Failure      500  {object}  dto.ErrorResponse   "Failed to delete print center"
// @Router       /admin/centers/{id} [delete]
func (c *printCenterController) DeletePrintCenter(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 8)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	if err := c.service.Delete(uint(id)); err != nil {
		HandleServiceError(ctx, err, "failed to delete print center")
		return
	}
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "print center deleted successfully"})
}

// --- Admin-specific handlers ---

// GetPendingPrintCenters godoc
// @Summary      Get all pending print centers
// @Description  Retrieves a list of all print centers awaiting approval. Requires admin role.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   entity.PrintCenter
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch pending centers"
// @Router       /admin/centers/pending [get]
func (c *printCenterController) GetPendingPrintCenters(ctx *gin.Context) {
	centers, err := c.service.GetPending()
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch pending centers")
		return
	}
	ctx.JSON(http.StatusOK, centers)
}

// GetAllPrintCenters godoc
// @Summary      Get all print centers (admin)
// @Description  Retrieves a list of all print centers, regardless of status. Requires admin role.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
func (c *printCenterController) GetAllPrintCenters(ctx *gin.Context) {
	centers, err := c.service.GetAll()
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch all centers")
		return
	}
	ctx.JSON(http.StatusOK, centers)
}

// UpdatePrintCenterStatus godoc
// @Summary      Update a print center's status
// @Description  Approves, rejects, or suspends a print center. Requires admin role.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                              true  "Print Center ID"
// @Param        status  body      dto.UpdatePrintCenterStatusRequest  true  "New status"
// @Success      200     {object}  dto.SuccessResponse "Status updated"
// @Failure      400     {object}  dto.ErrorResponse   "Invalid input"
// @Failure      404     {object}  dto.ErrorResponse   "Print center not found"
// @Failure      500     {object}  dto.ErrorResponse   "Failed to update status"
// @Router       /admin/centers/{id}/status [patch]
func (c *printCenterController) UpdatePrintCenterStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 8)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.UpdatePrintCenterStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	if err := c.validate.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.service.UpdateStatus(uint(id), req.Status); err != nil {
		HandleServiceError(ctx, err, "failed to update status")
		return
	}
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "status updated"})
}
