package woocommerce

import (
	"fmt"
	"net/http"
)

const (
	customersBasePath = "customers"
)

// CustomerService is an interface for interfacing with the customer endpoints of WooCommerce API
// https://woocommerce.github.io/woocommerce-rest-api-docs/#customers
type CustomerService interface {
	Create(customer Customer) (*Customer, error)
	Get(customerID int64, options interface{}) (*Customer, error)
	List(options interface{}) ([]Customer, error)
	Update(customer *Customer) (*Customer, error)
	Delete(customerID int64, options interface{}) (*Customer, error)
	Batch(option CustomerBatchOption) (*CustomerBatchResource, error)
}

// CustomerServiceOp handles communication with the customer related methods of the WooCommerce API
type CustomerServiceOp struct {
	client *Client
}

// CustomerListOption lists all the customer list option request params
// Reference URL:
// https://woocommerce.github.io/woocommerce-rest-api-docs/#list-all-customers
type CustomerListOption struct {
	ListOptions
	Email    string `url:"email,omitempty"`
	Role     string `url:"role,omitempty"`
}

// CustomerBatchOption allows for batch operations on customers
// https://woocommerce.github.io/woocommerce-rest-api-docs/#batch-update-customers
type CustomerBatchOption struct {
	Create []Customer `json:"create,omitempty"`
	Update []Customer `json:"update,omitempty"`
	Delete []int64    `json:"delete,omitempty"`
}

// CustomerBatchResource handles the response struct for CustomerBatchOption request
type CustomerBatchResource struct {
	Create []*Customer `json:"create,omitempty"`
	Update []*Customer `json:"update,omitempty"`
	Delete []*Customer `json:"delete,omitempty"`
}

// Customer represents a WooCommerce Customer
// https://woocommerce.github.io/woocommerce-rest-api-docs/#customer-properties
type Customer struct {
	ID                int64         `json:"id,omitempty"`
	Email             string        `json:"email,omitempty"`
	FirstName         string        `json:"first_name,omitempty"`
	LastName          string        `json:"last_name,omitempty"`
	Role              string        `json:"role,omitempty"`
	Username          string        `json:"username,omitempty"`
	Billing           *Billing      `json:"billing,omitempty"`
	Shipping          *Shipping     `json:"shipping,omitempty"`
	DateCreated       string        `json:"date_created,omitempty"`
	DateCreatedGmt    string        `json:"date_created_gmt,omitempty"`
	DateModified      string        `json:"date_modified,omitempty"`
	DateModifiedGmt   string        `json:"date_modified_gmt,omitempty"`
	OrdersCount       int           `json:"orders_count,omitempty"`
	TotalSpent        string        `json:"total_spent,omitempty"`
	AvatarURL         string        `json:"avatar_url,omitempty"`
	MetaData          []MetaData    `json:"meta_data,omitempty"`
	Links             Links         `json:"_links,omitempty"`
}

func (c *CustomerServiceOp) List(options interface{}) ([]Customer, error) {
	customers, _, err := c.ListWithPagination(options)
	return customers, err
}

// ListWithPagination lists customers and returns pagination to retrieve next/previous results.
func (c *CustomerServiceOp) ListWithPagination(options interface{}) ([]Customer, *Pagination, error) {
	path := fmt.Sprintf("%s", customersBasePath)
	resource := make([]Customer, 0)
	headers := http.Header{}
	headers, err := c.client.createAndDoGetHeaders("GET", path, nil, options, &resource)
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

func (c *CustomerServiceOp) Create(customer Customer) (*Customer, error) {
	path := fmt.Sprintf("%s", customersBasePath)
	resource := new(Customer)
	err := c.client.Post(path, customer, &resource)
	return resource, err
}

// Get individual customer
func (c *CustomerServiceOp) Get(customerID int64, options interface{}) (*Customer, error) {
	path := fmt.Sprintf("%s/%d", customersBasePath, customerID)
	resource := new(Customer)
	err := c.client.Get(path, resource, options)
	return resource, err
}

func (c *CustomerServiceOp) Update(customer *Customer) (*Customer, error) {
	path := fmt.Sprintf("%s/%d", customersBasePath, customer.ID)
	resource := new(Customer)
	err := c.client.Put(path, customer, &resource)
	return resource, err
}

func (c *CustomerServiceOp) Delete(customerID int64, options interface{}) (*Customer, error) {
	path := fmt.Sprintf("%s/%d", customersBasePath, customerID)
	resource := new(Customer)
	err := c.client.Delete(path, options, &resource)
	return resource, err
}

func (c *CustomerServiceOp) Batch(data CustomerBatchOption) (*CustomerBatchResource, error) {
	path := fmt.Sprintf("%s/batch", customersBasePath)
	resource := new(CustomerBatchResource)
	err := c.client.Post(path, data, &resource)
	return resource, err
}
