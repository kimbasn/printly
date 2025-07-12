package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kimbasn/printly/internal/dto"
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
	service  service.OrderService
	validate *validator.Validate
}

func NewOrderController(service service.OrderService, validate *validator.Validate) OrderController {
	return &orderController{
		service:  service,
		validate: validate,
	}
}

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Creates a new order with one or more documents and returns the created order. Requires authentication.
// @Tags         Print Centers
// @Accept       json
// @Produce      json
// @Param        id     path      string                  true  "Print Center ID"
// @Param        order  body      dto.CreateOrderRequest  true  "Order creation request"
// @Success      201    {object}  entity.Order
// @Failure      400    {object}  dto.ErrorResponse "Invalid input"
// @Failure      401    {object}  dto.ErrorResponse "Unauthorized"
// @Failure      404    {object}  dto.ErrorResponse "Print center not found"
// @Failure      500    {object}  dto.ErrorResponse "Failed to create order"
// @Router       /centers/{id}/orders [post]
func (c *orderController) CreateOrder(ctx *gin.Context) {
	userUID, exists := ctx.Get("userUID")
	if !exists {
		// This should not happen if the AuthMiddleware is applied correctly.
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user UID not found in context"})
		return
	}

	centerID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid print center ID"})
		return
	}

	var req dto.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.validate.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	order, err := c.service.CreateOrder(userUID.(string), uint(centerID), req)
	if err != nil {
		HandleServiceError(ctx, err, "failed to create order")
		return
	}

	ctx.JSON(http.StatusCreated, order)
}

// GetOrderByID godoc
// @Summary      Get an order by ID
// @Description  Retrieves a single order by its ID. Requires admin role.
// @Tags         Admin
// @Produce      json
// @Param        id   path      string       true  "Order ID"
// @Success      200  {object}  entity.Order
// @Failure      400  {object}  dto.ErrorResponse "Invalid ID"
// @Failure      404  {object}  dto.ErrorResponse "Order not found"
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch order"
// @Router       /admin/orders/{id} [get]
func (c *orderController) GetOrderByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
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
// @Param        id   path      string       true  "Print Center ID"
// @Success      200  {array}   entity.Order
// @Failure      400  {object}  dto.ErrorResponse "Invalid ID"
// @Failure      500  {object}  dto.ErrorResponse "Failed to fetch orders"
// @Router       /centers/{id}/orders [get]
func (c *orderController) GetOrdersForCenter(ctx *gin.Context) {
	centerID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
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
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order ID"})
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	if err := c.validate.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.service.UpdateOrderStatus(uint(id), req.Status); err != nil {
		HandleServiceError(ctx, err, "failed to update order status")
		return
	}
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "status updated"})
}

// DeleteOrder godoc
// @Summary      Delete an order (admin)
// @Description  Deletes an order. Requires admin role.
// @Tags         Admin
// @Produce      json
// @Param        id   path      string  true  "Order ID"
// @Success      200  {object}  dto.SuccessResponse "Order deleted successfully"
// @Failure      400  {object}  dto.ErrorResponse "Invalid ID"
// @Failure      404  {object}  dto.ErrorResponse   "Order not found"
// @Failure      500  {object}  dto.ErrorResponse   "Failed to delete order"
// @Router       /admin/orders/{id} [delete]
func (c *orderController) DeleteOrder(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order ID"})
		return
	}

	if err := c.service.DeleteOrder(uint(id)); err != nil {
		HandleServiceError(ctx, err, "failed to delete order")
		return
	}
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "order deleted successfully"})
}
