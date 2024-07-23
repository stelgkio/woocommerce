package woocommerce

import (
	"fmt"
)

const (
	reportsBasePath = "reports"
)

// ReportService is an interface for interfacing with the report endpoints of the WooCommerce API
// https://woocommerce.github.io/woocommerce-rest-api-docs/#reports
type ReportService interface {
	Get(reportID string, options interface{}) (*Report, error)
	List(options interface{}) ([]Report, error)
	GetTotalOrders(options interface{}) ([]TotalOrdersReport, error)
	GetTotalCustomers(options interface{}) ([]TotalCustomersReport, error)
	GetTotalProducts(options interface{}) ([]TotalProductsReport, error)
}

// ReportServiceOp handles communication with the report related methods of the WooCommerce API
type ReportServiceOp struct {
	client *Client
}

// Report represents a WooCommerce Report
// https://woocommerce.github.io/woocommerce-rest-api-docs/#report-properties
type Report struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Total string `json:"total,omitempty"`
}

// TotalOrdersReport represents a report for total orders
type TotalOrdersReport struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Total int    `json:"total"`
}

// TotalCustomersReport represents a report for total customers
type TotalCustomersReport struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Total int    `json:"total"`
}

// TotalProductsReport represents a report for total products
type TotalProductsReport struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Total int    `json:"total"`
}

// Get individual report
func (r *ReportServiceOp) Get(reportID string, options interface{}) (*Report, error) {
	path := fmt.Sprintf("%s/%s", reportsBasePath, reportID)
	resource := new(Report)
	err := r.client.Get(path, resource, options)
	return resource, err
}

// List all reports
func (r *ReportServiceOp) List(options interface{}) ([]Report, error) {
	path := fmt.Sprintf("%s", reportsBasePath)
	resource := make([]Report, 0)
	err := r.client.Get(path, &resource, options)
	return resource, err
}

// GetTotalOrders retrieves a report for total orders
func (r *ReportServiceOp) GetTotalOrders(options interface{}) ([]TotalOrdersReport, error) {
	path := fmt.Sprintf("%s/orders/totals", reportsBasePath)
	resource := make([]TotalOrdersReport, 0)
	err := r.client.Get(path, &resource, options)
	return resource, err
}

// GetTotalCustomers retrieves a report for total customers
func (r *ReportServiceOp) GetTotalCustomers(options interface{}) ([]TotalCustomersReport, error) {
	path := fmt.Sprintf("%s/customers/totals", reportsBasePath)
	resource := make([]TotalCustomersReport, 0)
	err := r.client.Get(path, &resource, options)
	return resource, err
}

// GetTotalProducts retrieves a report for total products
func (r *ReportServiceOp) GetTotalProducts(options interface{}) ([]TotalProductsReport, error) {
	path := fmt.Sprintf("%s/products/totals", reportsBasePath)
	resource := make([]TotalProductsReport, 0)
	err := r.client.Get(path, &resource, options)
	return resource, err
}
