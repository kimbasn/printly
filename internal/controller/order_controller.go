package controller

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/entity"
	"github.com/kimbasn/printly/internal/service"
)

type OrderController interface {
	CreateOrder(ctx *gin.Context)
	GetOrderByID(ctx *gin.Context)
	GetOrderByCode(ctx *gin.Context)
	GetOrdersForCenter(ctx *gin.Context)
	GetAllOrders(ctx *gin.Context)
	UpdateOrderStatus(ctx *gin.Context)
	DeleteOrder(ctx *gin.Context)
}

type orderController struct {
	service        service.OrderService
	storageService service.StorageService
	validate       *validator.Validate
	logger         *zap.Logger
}

func NewOrderController(service service.OrderService, storageService service.StorageService, validate *validator.Validate, logger *zap.Logger) OrderController {
	return &orderController{
		service:        service,
		storageService: storageService,
		validate:       validate,
		logger:         logger,
	}
}

const (
	MAX_FILE_SIZE_MB = 50
	MAX_FORM_SIZE    = MAX_FILE_SIZE_MB << 20
)

func (c *orderController) validateFile(fileHeader *multipart.FileHeader) error {
	// Check file size
	maxSize := int64(MAX_FILE_SIZE_MB << 20)
	if fileHeader.Size > maxSize {
		return fmt.Errorf("file too large (max %dMB)", MAX_FILE_SIZE_MB)
	}

	// Check file extension
	allowedExtensions := []string{".pdf", ".doc", ".docx", ".txt", ".jpg", ".jpeg", ".png"}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	isAllowed := false
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("unsupported file type %s", ext)
	}

	// Check MIME type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("missing content type")
	}

	// Additional MIME type validation
	allowedMimeTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
		"image/jpeg",
		"image/jpg",
		"image/png",
	}

	isAllowedMime := false
	for _, allowedMime := range allowedMimeTypes {
		if contentType == allowedMime {
			isAllowedMime = true
			break
		}
	}

	if !isAllowedMime {
		return fmt.Errorf("unsupported content type %s", contentType)
	}

	return nil
}

// isValidPrintMode validates the print mode value
func isValidPrintMode(mode string) bool {
	return mode == string(entity.PrePrint) || mode == string(entity.PrintUponArrival)
}

// cleanupUploadedFiles removes uploaded files if order creation fails
func (c *orderController) cleanupUploadedFiles(documents []dto.CreateDocumentRequest) {
	var cleanupErrors []error

	for _, doc := range documents {
		if doc.StoragePath != "" {
			if err := c.storageService.DeleteFile(doc.StoragePath); err != nil {
				cleanupErrors = append(cleanupErrors, err)
				c.logger.Error("failed to cleanup file",
					zap.String("storage_path", doc.StoragePath),
					zap.Error(err))
			}
		}
	}

	if len(cleanupErrors) > 0 {
		c.logger.Error("cleanup completed with errors",
			zap.Int("failed_count", len(cleanupErrors)))
	}
}

var fileNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

// normalizeFileName cleans a file name for safe storage. It performs the following steps:
// 1. Removes the file extension.
// 2. Replaces spaces and any non-alphanumeric characters (except underscores) with a single underscore.
// 3. Trims any leading or trailing underscores that might result from the replacement.
// 4. Converts the entire string to lowercase.
// For example, "My Document (v2).PDF" becomes "my_document_v2".
func normalizeFileName(fileName string) string {
	// Get the base name without the extension.
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	// Replace any non-alphanumeric characters (except underscore) with an underscore.
	sanitized := fileNameSanitizer.ReplaceAllString(base, "_")
	// Trim leading/trailing underscores that might result from the replacement.
	trimmed := strings.Trim(sanitized, "_")
	return strings.ToLower(trimmed)
}

// CreateOrder godoc
// @Summary      Create a new order with file uploads
// @Description  Creates a new order with one or more documents uploaded as files. Each document can have its own print mode and options. Requires authentication.
// @Tags         Print Centers
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string                  true  "Print Center ID"
// @Param        files        formData  file                    true  "Document files (multiple files allowed)"
// @Param        document_configs formData string               true  "JSON array of document configurations (print_mode and print_options for each file)"
// @Success      201          {object}  entity.Order
// @Failure      400          {object}  dto.ErrorResponse "Invalid input"
// @Failure      401          {object}  dto.ErrorResponse "Unauthorized"
// @Failure      404          {object}  dto.ErrorResponse "Print center not found"
// @Failure      413          {object}  dto.ErrorResponse "File too large"
// @Failure      500          {object}  dto.ErrorResponse "Failed to create order"
// @Router       /centers/{id}/orders [post]
func (c *orderController) CreateOrder(ctx *gin.Context) {
	userUID, exists := ctx.Get("userUID")
	if !exists {
		c.logger.Error("user UID not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user UID not found in context"})
		return
	}

	centerID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		c.logger.Error("invalid print center ID", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid print center ID"})
		return
	}

	// Parse multipart form
	if err := ctx.Request.ParseMultipartForm(MAX_FORM_SIZE); err != nil {
		c.logger.Error("failed to parse multipart form", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "failed to parse multipart form"})
		return
	}

	// Get files
	form, err := ctx.MultipartForm()
	if err != nil {
		c.logger.Error("failed to get multipart form", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "failed to get multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.logger.Error("no files provided")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "at least one file is required"})
		return
	}

	// Get document configurations (print mode + print options for each file)
	documentConfigsJSON := ctx.PostForm("document_configs")
	if documentConfigsJSON == "" {
		c.logger.Error("document_configs not provided")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "document_configs is required",
		})
		return
	}

	var documentConfigs []dto.DocumentPrintRequest
	if err := json.Unmarshal([]byte(documentConfigsJSON), &documentConfigs); err != nil {
		c.logger.Error("failed to parse document_configs JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "invalid document_configs JSON",
		})
		return
	}

	// Validate that we have configurations for each file
	if len(documentConfigs) != len(files) {
		c.logger.Error("document_configs count mismatch",
			zap.Int("configs_count", len(documentConfigs)),
			zap.Int("files_count", len(files)))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: fmt.Sprintf("document_configs count (%d) must match files count (%d)", len(documentConfigs), len(files)),
		})
		return
	}

	// Process files and create document requests
	var documentRequests []dto.CreateDocumentRequest

	for i, fileHeader := range files {
		// Validate file
		if err := c.validateFile(fileHeader); err != nil {
			c.logger.Error("file validation failed",
				zap.Int("file_index", i),
				zap.String("filename", fileHeader.Filename),
				zap.Error(err))
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: fmt.Sprintf("file %d (%s): %s", i+1, fileHeader.Filename, err.Error()),
			})
			return
		}

		// Validate document configuration for this file
		if err := c.validate.Struct(documentConfigs[i]); err != nil {
			c.logger.Error("document config validation failed",
				zap.Int("config_index", i),
				zap.Error(err))
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: fmt.Sprintf("document_configs[%d]: %s", i, err.Error()),
			})
			return
		}

		// Validate print mode
		if !isValidPrintMode(documentConfigs[i].PrintMode) {
			c.logger.Error("invalid print mode",
				zap.Int("config_index", i),
				zap.String("print_mode", documentConfigs[i].PrintMode))
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: fmt.Sprintf("document_configs[%d]: invalid print_mode '%s'", i, documentConfigs[i].PrintMode),
			})
			return
		}

		// Open and process the file
		file, err := fileHeader.Open()
		if err != nil {
			c.logger.Error("failed to open file",
				zap.String("filename", fileHeader.Filename),
				zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: fmt.Sprintf("failed to open file %s", fileHeader.Filename),
			})
			return
		}

		// Upload file to storage
		normalizedFileName := normalizeFileName(fileHeader.Filename)
		storagePath, err := c.storageService.UploadFile(file, normalizedFileName, userUID.(string))
		file.Close() // Close immediately after use to prevent memory leaks

		if err != nil {
			c.logger.Error("failed to upload file",
				zap.String("filename", fileHeader.Filename),
				zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: fmt.Sprintf("failed to upload file %s", fileHeader.Filename),
			})
			return
		}

		// Create document request with individual print mode and options
		documentRequests = append(documentRequests, dto.CreateDocumentRequest{
			FileName:     fileHeader.Filename,
			MimeType:     fileHeader.Header.Get("Content-Type"),
			Size:         fileHeader.Size,
			StoragePath:  storagePath,
			PrintMode:    entity.PrintMode(documentConfigs[i].PrintMode),
			PrintOptions: documentConfigs[i].PrintOptions,
		})
	}

	req := dto.CreateOrderRequest{
		Documents: documentRequests,
	}

	// Validate the complete request
	if err := c.validate.Struct(req); err != nil {
		c.logger.Error("order request validation failed", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Create the order
	order, err := c.service.CreateOrder(userUID.(string), uint(centerID), req)
	if err != nil {
		c.cleanupUploadedFiles(documentRequests)
		HandleServiceError(ctx, err, "failed to create order")
		return
	}

	c.logger.Info("order created successfully",
		zap.String("user_uid", userUID.(string)),
		zap.Uint64("center_id", centerID),
		zap.Uint("order_id", order.ID))

	ctx.JSON(http.StatusCreated, order)
}

// GetOrderByID godoc
// @Summary      Get an order by ID
// @Description  Retrieves a single order by its ID. Requires admin role.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string       true  "Order ID"
// @Success      200  {object}  entity.Order
// @Failure      400  {object}  dto.ErrorResponse "Invalid ID"
// @Failure      404  {object}  dto.ErrorResponse "Order not found"
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch order"
// @Router       /admin/orders/{id} [get]
func (c *orderController) GetOrderByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		c.logger.Error("invalid order ID", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order ID"})
		return
	}

	order, err := c.service.GetOrderByID(uint(id))
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch order")
		return
	}

	ctx.JSON(http.StatusOK, order)
}

// GetOrderByCode godoc
// @Summary      Get order status by pickup code
// @Description  Retrieves the status of an order using its public pickup code.
// @Tags         Orders
// @Produce      json
// @Param        code path      string       true  "Pickup Code"
// @Success      200  {object}  entity.Order
// @Failure      404  {object}  dto.ErrorResponse "Order not found"
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch order status"
// @Router       /orders/status/{code} [get]
func (c *orderController) GetOrderByCode(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		c.logger.Error("empty pickup code provided")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "pickup code is required"})
		return
	}

	order, err := c.service.GetOrderByCode(code)
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch order status")
		return
	}

	ctx.JSON(http.StatusOK, order)
}

// GetOrdersForCenter godoc
// @Summary      Get orders for a print center
// @Description  Retrieves all orders for a specific print center. Requires manager or admin role.
// @Tags         Print Centers
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string       true  "Print Center ID"
// @Success      200  {array}   entity.Order
// @Failure      400  {object}  dto.ErrorResponse "Invalid ID"
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch orders"
// @Router       /centers/{id}/orders [get]
func (c *orderController) GetOrdersForCenter(ctx *gin.Context) {
	centerID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		c.logger.Error("invalid print center ID", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid print center ID"})
		return
	}

	orders, err := c.service.GetOrdersForCenter(uint(centerID))
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch orders for center")
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// GetAllOrders godoc
// @Summary      Get all orders (admin)
// @Description  Retrieves a list of all orders across the platform. Requires admin role.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   entity.Order
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch all orders"
// @Router       /admin/orders [get]
func (c *orderController) GetAllOrders(ctx *gin.Context) {
	orders, err := c.service.GetAllOrders()
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch all orders")
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus godoc
// @Summary      Update an order's status
// @Description  Updates the status of an order. Requires manager or admin role.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                      true  "Order ID"
// @Param        status  body      dto.UpdateOrderStatusRequest  true  "New status"
// @Success      200     {object}  dto.SuccessResponse "Status updated"
// @Failure      400     {object}  dto.ErrorResponse   "Invalid input"
// @Failure      404     {object}  dto.ErrorResponse   "Order not found"
// @Failure      500     {object}  dto.ErrorResponse   "Failed to update status"
// @Router       /orders/{id}/status [patch]
func (c *orderController) UpdateOrderStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		c.logger.Error("invalid order ID", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order ID"})
		return
	}

	userUID, exists := ctx.Get("userUID")
	if !exists {
		c.logger.Error("user UID not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user UID not found in context"})
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("failed to bind request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.validate.Struct(req); err != nil {
		c.logger.Error("request validation failed", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.service.UpdateOrderStatus(uint(id), req.Status, userUID.(string)); err != nil {
		HandleServiceError(ctx, err, "failed to update order status")
		return
	}

	c.logger.Info("order status updated",
		zap.Uint64("order_id", id),
		zap.String("new_status", string(req.Status)),
		zap.String("updated_by", userUID.(string)))

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "status updated"})
}

// DeleteOrder godoc
// @Summary      Delete an order (admin)
// @Description  Deletes an order. Requires admin role.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Order ID"
// @Success      200  {object}  dto.SuccessResponse "Order deleted successfully"
// @Failure      400  {object}  dto.ErrorResponse "Invalid ID"
// @Failure      404  {object}  dto.ErrorResponse   "Order not found"
// @Failure      500  {object}  dto.ErrorResponse   "Failed to delete order"
// @Router       /admin/orders/{id} [delete]
func (c *orderController) DeleteOrder(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		c.logger.Error("invalid order ID", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order ID"})
		return
	}

	if err := c.service.DeleteOrder(uint(id)); err != nil {
		HandleServiceError(ctx, err, "failed to delete order")
		return
	}

	c.logger.Info("order deleted", zap.Uint64("order_id", id))
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "order deleted successfully"})
}
