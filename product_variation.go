package woocommerce

import (
	"fmt"
	"net/http"
	"time"
)

const (
	variationsBasePath = "products/%d/variations"
)

// ProductVariationService is an interface for interfacing with the product variation endpoints of WooCommerce API
// https://woocommerce.github.io/woocommerce-rest-api-docs/#product-variations
type ProductVariationService interface {
	Create(productID int64, variation Product) (*Product, error)
	Get(productID, variationID int64, options interface{}) (*Product, error)
	List(productID int64, options interface{}) ([]Product, *Pagination, error)
	Update(productID, variationID int64, variation *Product) (*Product, error)
	Delete(productID, variationID int64, options interface{}) (*Product, error)
	Batch(productID int64, data ProductBatchOption) (*ProductBatchResource, error)
}

// ProductVariationServiceOp handles communication with the product variation related methods of the WooCommerce API
type ProductVariationServiceOp struct {
	client *Client
}

// ProductVariationListOptions represents the optional parameters for listing variations
type ProductVariationListOptions struct {
	ListOptions
	ModifiedAfter  time.Time `json:"modified_after,omitempty" url:"modified_after,omitempty"`
	ModifiedBefore time.Time `json:"modified_before,omitempty" url:"modified_before,omitempty"`
	DatesAreGMT    bool      `json:"dates_are_gmt,omitempty" url:"dates_are_gmt,omitempty"`
	Slug           string    `json:"slug,omitempty" url:"slug,omitempty"`
	Status         string    `json:"status,omitempty" url:"status,omitempty"`
	StockStatus    string    `json:"stock_status,omitempty" url:"stock_status,omitempty"`
	MinPrice       string    `json:"min_price,omitempty" url:"min_price,omitempty"`
	MaxPrice       string    `json:"max_price,omitempty" url:"max_price,omitempty"`
}

// Create new product variation
func (p *ProductVariationServiceOp) Create(productID int64, variation Product) (*Product, error) {
	path := fmt.Sprintf(variationsBasePath, productID)
	resource := new(Product)
	err := p.client.Post(path, variation, &resource)
	return resource, err
}

// Get individual product variation
func (p *ProductVariationServiceOp) Get(productID, variationID int64, options interface{}) (*Product, error) {
	path := fmt.Sprintf("%s/%d", fmt.Sprintf(variationsBasePath, productID), variationID)
	resource := new(Product)
	err := p.client.Get(path, resource, options)
	return resource, err
}

// List product variations
func (p *ProductVariationServiceOp) List(productID int64, options interface{}) ([]Product, *Pagination, error) {
	variations, pagination, err := p.ListWithPagination(productID, options)
	return variations, pagination, err
}

// ListWithPagination lists product variations and returns pagination to retrieve next/previous results.
func (p *ProductVariationServiceOp) ListWithPagination(productID int64, options interface{}) ([]Product, *Pagination, error) {
	path := fmt.Sprintf(variationsBasePath, productID)
	resource := make([]Product, 0)
	headers := http.Header{}
	headers, err := p.client.createAndDoGetHeaders("GET", path, nil, options, &resource)
	if err != nil {
		return nil, nil, err
	}
	// Extract pagination info from header
	linkHeader := headers.Get("Link")
	fmt.Println(linkHeader)
	pagination, err := extractPagination(linkHeader)
	if err != nil {
		return nil, nil, err
	}

	return resource, pagination, err
}

// Update existing product variation
func (p *ProductVariationServiceOp) Update(productID, variationID int64, variation *Product) (*Product, error) {
	path := fmt.Sprintf("%s/%d", fmt.Sprintf(variationsBasePath, productID), variationID)
	resource := new(Product)
	err := p.client.Put(path, variation, &resource)
	return resource, err
}

// Delete existing product variation
func (p *ProductVariationServiceOp) Delete(productID, variationID int64, options interface{}) (*Product, error) {
	path := fmt.Sprintf("%s/%d", fmt.Sprintf(variationsBasePath, productID), variationID)
	resource := new(Product)
	err := p.client.Delete(path, options, &resource)
	return resource, err
}

// Batch implements ProductVariationService.
func (p *ProductVariationServiceOp) Batch(productID int64, data ProductBatchOption) (*ProductBatchResource, error) {
	path := fmt.Sprintf("%s/%d/variations/batch", productsBasePath, productID)
	resource := new(ProductBatchResource)
	err := p.client.Post(path, data, &resource)
	return resource, err
}
