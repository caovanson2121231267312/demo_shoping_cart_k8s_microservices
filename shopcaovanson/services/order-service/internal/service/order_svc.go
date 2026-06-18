package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/order-service/internal/client"
	"github.com/caovanson/shopcaovanson/order-service/internal/domain"
	"github.com/caovanson/shopcaovanson/order-service/internal/kafka"
	"github.com/caovanson/shopcaovanson/order-service/internal/rbac"
	"github.com/caovanson/shopcaovanson/order-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrForbidden     = errors.New("forbidden")
	ErrEmptyCart     = errors.New("cart is empty")
	ErrInvalidStatus = errors.New("invalid status")
)

type CartService struct {
	cartRepo      repository.CartRepository
	productClient *client.ProductClient
	cartTTL       time.Duration
}

func NewCartService(cartRepo repository.CartRepository, productClient *client.ProductClient, cartTTL time.Duration) *CartService {
	return &CartService{
		cartRepo:      cartRepo,
		productClient: productClient,
		cartTTL:       cartTTL,
	}
}

func (s *CartService) GetCart(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	return s.cartRepo.Get(ctx, userID)
}

func (s *CartService) AddItem(ctx context.Context, userID uuid.UUID, input domain.AddCartItemInput) (*domain.Cart, error) {
	if input.Quantity < 1 {
		return nil, fmt.Errorf("%w: quantity must be at least 1", ErrInvalidInput)
	}
	product, err := s.productClient.GetProduct(ctx, input.ProductID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if !product.IsActive {
		return nil, fmt.Errorf("%w: product is not active", ErrInvalidInput)
	}
	if product.Stock < input.Quantity {
		return nil, fmt.Errorf("%w: insufficient stock", ErrInvalidInput)
	}

	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	found := false
	for i, item := range cart.Items {
		if item.ProductID == input.ProductID {
			newQty := item.Quantity + input.Quantity
			if product.Stock < newQty {
				return nil, fmt.Errorf("%w: insufficient stock", ErrInvalidInput)
			}
			cart.Items[i].Quantity = newQty
			cart.Items[i].UnitPrice = product.EffectivePrice()
			cart.Items[i].ProductName = product.Name
			cart.Items[i].ProductImage = product.PrimaryImage()
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, domain.CartItem{
			ProductID:    input.ProductID,
			Quantity:     input.Quantity,
			UnitPrice:    product.EffectivePrice(),
			ProductName:  product.Name,
			ProductImage: product.PrimaryImage(),
		})
	}

	if err := s.cartRepo.Save(ctx, cart, s.cartTTL); err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) UpdateItem(ctx context.Context, userID, productID uuid.UUID, input domain.UpdateCartItemInput) (*domain.Cart, error) {
	if input.Quantity < 1 {
		return nil, fmt.Errorf("%w: quantity must be at least 1", ErrInvalidInput)
	}
	product, err := s.productClient.GetProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if product.Stock < input.Quantity {
		return nil, fmt.Errorf("%w: insufficient stock", ErrInvalidInput)
	}

	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			cart.Items[i].Quantity = input.Quantity
			cart.Items[i].UnitPrice = product.EffectivePrice()
			cart.Items[i].ProductName = product.Name
			cart.Items[i].ProductImage = product.PrimaryImage()
			found = true
			break
		}
	}
	if !found {
		return nil, ErrNotFound
	}

	if err := s.cartRepo.Save(ctx, cart, s.cartTTL); err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (*domain.Cart, error) {
	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	newItems := make([]domain.CartItem, 0, len(cart.Items))
	found := false
	for _, item := range cart.Items {
		if item.ProductID == productID {
			found = true
			continue
		}
		newItems = append(newItems, item)
	}
	if !found {
		return nil, ErrNotFound
	}
	cart.Items = newItems
	if err := s.cartRepo.Save(ctx, cart, s.cartTTL); err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.cartRepo.Delete(ctx, userID)
}

type OrderService struct {
	orderRepo     repository.OrderRepository
	cartRepo      repository.CartRepository
	cartSvc       *CartService
	productClient *client.ProductClient
	couponSvc     *CouponService
	producer      *kafka.Producer
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	cartSvc *CartService,
	productClient *client.ProductClient,
	couponSvc *CouponService,
	producer *kafka.Producer,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		cartSvc:       cartSvc,
		productClient: productClient,
		couponSvc:     couponSvc,
		producer:      producer,
	}
}

func (s *OrderService) Checkout(ctx context.Context, userID uuid.UUID, userEmail string, input domain.CheckoutInput) (*domain.Order, error) {
	if strings.TrimSpace(input.ShippingName) == "" ||
		strings.TrimSpace(input.ShippingPhone) == "" ||
		strings.TrimSpace(input.ShippingAddress) == "" {
		return nil, fmt.Errorf("%w: shipping information is required", ErrInvalidInput)
	}

	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, ErrEmptyCart
	}

	var subtotal float64
	orderItems := make([]domain.OrderItem, 0, len(cart.Items))
	now := time.Now().UTC()
	orderID := uuid.New()

	for _, item := range cart.Items {
		product, err := s.productClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("%w: product %s unavailable", ErrInvalidInput, item.ProductID)
		}
		if !product.IsActive {
			return nil, fmt.Errorf("%w: product %s is not active", ErrInvalidInput, product.Name)
		}
		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("%w: insufficient stock for %s", ErrInvalidInput, product.Name)
		}

		unitPrice := product.EffectivePrice()
		lineTotal := unitPrice * float64(item.Quantity)
		subtotal += lineTotal

		img := product.PrimaryImage()
		var imgPtr *string
		if img != "" {
			imgPtr = &img
		}

		orderItems = append(orderItems, domain.OrderItem{
			ID:                   uuid.New(),
			OrderID:              orderID,
			ProductID:            item.ProductID,
			ProductNameSnapshot:  product.Name,
			ProductImageSnapshot: imgPtr,
			UnitPrice:            unitPrice,
			Quantity:             item.Quantity,
		})
	}

	var discount float64
	var couponCode *string
	if code := strings.TrimSpace(input.CouponCode); code != "" && s.couponSvc != nil {
		d, coupon, err := s.couponSvc.Apply(ctx, code, subtotal)
		if err != nil {
			return nil, err
		}
		discount = d
		couponCode = &coupon.Code
	}
	total := subtotal - discount
	if total < 0 {
		total = 0
	}

	order := &domain.Order{
		ID:              orderID,
		UserID:          userID,
		Status:          domain.OrderStatusPending,
		SubtotalAmount:  subtotal,
		DiscountAmount:  discount,
		CouponCode:      couponCode,
		TotalAmount:     total,
		ShippingName:    strings.TrimSpace(input.ShippingName),
		ShippingPhone:   strings.TrimSpace(input.ShippingPhone),
		ShippingAddress: strings.TrimSpace(input.ShippingAddress),
		CreatedAt:       now,
		UpdatedAt:       now,
		Items:           orderItems,
	}

	orderNumber, err := s.orderRepo.NextOrderNumber(ctx)
	if err != nil {
		return nil, err
	}
	order.OrderNumber = orderNumber

	if err := s.orderRepo.Create(ctx, order, orderItems); err != nil {
		return nil, err
	}
	if err := s.cartRepo.Delete(ctx, userID); err != nil {
		return nil, err
	}

	eventItems := make([]kafka.OrderCreatedItem, len(orderItems))
	for i, item := range orderItems {
		eventItems[i] = kafka.OrderCreatedItem{
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductNameSnapshot,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}
	event := kafka.OrderCreatedEvent{
		OrderID:         order.ID.String(),
		OrderNumber:     order.OrderNumber,
		UserID:          order.UserID.String(),
		UserEmail:       userEmail,
		Items:           eventItems,
		SubtotalAmount:  order.SubtotalAmount,
		DiscountAmount:  order.DiscountAmount,
		CouponCode:      "",
		TotalAmount:     order.TotalAmount,
		ShippingName:    order.ShippingName,
		ShippingPhone:   order.ShippingPhone,
		ShippingAddress: order.ShippingAddress,
		CreatedAt:       order.CreatedAt.Format(time.RFC3339),
	}
	if order.CouponCode != nil {
		event.CouponCode = *order.CouponCode
	}
	_ = s.producer.PublishOrderCreated(ctx, event)
	_ = s.producer.PublishInvoiceGenerate(ctx, event)

	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, userID uuid.UUID, role string, orderID uuid.UUID) (*domain.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}
	if !rbac.CanViewAllOrders(role) && order.UserID != userID {
		return nil, ErrForbidden
	}
	return order, nil
}

func (s *OrderService) TrackOrder(ctx context.Context, input domain.TrackOrderInput) (*domain.Order, error) {
	orderNumber := strings.TrimSpace(input.OrderNumber)
	phone := strings.TrimSpace(input.ShippingPhone)
	if orderNumber == "" || phone == "" {
		return nil, fmt.Errorf("%w: order_number and shipping_phone are required", ErrInvalidInput)
	}
	order, err := s.orderRepo.GetByOrderNumberAndPhone(ctx, orderNumber, phone)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}
	return order, nil
}

func (s *OrderService) SearchOrders(ctx context.Context, filter domain.OrderSearchFilter) (*domain.OrderListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Status != "" && !isValidStatus(filter.Status) {
		return nil, ErrInvalidStatus
	}
	return s.orderRepo.Search(ctx, filter)
}

func (s *OrderService) OrderStats(ctx context.Context) (*domain.OrderStats, error) {
	return s.orderRepo.Stats(ctx)
}

func (s *OrderService) GetByOrderNumber(ctx context.Context, userID uuid.UUID, role, orderNumber string) (*domain.Order, error) {
	order, err := s.orderRepo.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}
	if !rbac.CanViewAllOrders(role) && order.UserID != userID {
		return nil, ErrForbidden
	}
	return order, nil
}

func (s *OrderService) ListMyOrders(ctx context.Context, userID uuid.UUID, page, limit int) (*domain.OrderListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.orderRepo.ListByUser(ctx, userID, page, limit)
}

func (s *OrderService) CancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*domain.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}
	if order.UserID != userID {
		return nil, ErrForbidden
	}
	if order.Status != domain.OrderStatusPending {
		return nil, fmt.Errorf("%w: only pending orders can be cancelled", ErrInvalidInput)
	}
	if err := s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusCancelled); err != nil {
		return nil, err
	}
	return s.orderRepo.GetByID(ctx, orderID)
}

func (s *OrderService) ListAllOrders(ctx context.Context, status string, page, limit int) (*domain.OrderListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if status != "" && !isValidStatus(status) {
		return nil, ErrInvalidStatus
	}
	return s.orderRepo.ListAll(ctx, status, page, limit)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) (*domain.Order, error) {
	if !isValidStatus(status) {
		return nil, ErrInvalidStatus
	}
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}
	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		return nil, err
	}
	return s.orderRepo.GetByID(ctx, orderID)
}

func isValidStatus(status string) bool {
	for _, s := range domain.ValidOrderStatuses {
		if s == status {
			return true
		}
	}
	return false
}
