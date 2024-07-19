package woocommerce

import (
	"fmt"
	"net/http"
)

const (
	productsBasePath = "products"
)

// ProductService is an interface for interfacing with the products endpoints of WooCommerce API
// https://woocommerce.github.io/woocommerce-rest-api-docs/#products
type ProductService interface {
	Create(product Product) (*Product, error)
	Get(productID int64, options interface{}) (*Product, error)
	List(options interface{}) ([]Product, error)
	Update(product *Product) (*Product, error)
	Delete(productID int64, options interface{}) (*Product, error)
	Batch(option ProductBatchOption) (*ProductBatchResource, error)
}

// ProductServiceOp handles communication with the product related methods of the WooCommerce API
type ProductServiceOp struct {
	client *Client
}



// ProductBatchOption allows for batch operations on products
// https://woocommerce.github.io/woocommerce-rest-api-docs/#batch-update-products
type ProductBatchOption struct {
	Create []Product `json:"create,omitempty"`
	Update []Product `json:"update,omitempty"`
	Delete []int64   `json:"delete,omitempty"`
}

// ProductBatchResource handles the response struct for ProductBatchOption request
type ProductBatchResource struct {
	Create []*Product `json:"create,omitempty"`
	Update []*Product `json:"update,omitempty"`
	Delete []*Product `json:"delete,omitempty"`
}

// Product represents a WooCommerce Product
// https://woocommerce.github.io/woocommerce-rest-api-docs/#product-properties
type Product struct {
	ID                int64         `json:"id,omitempty"`
	Name              string        `json:"name,omitempty"`
	Slug              string        `json:"slug,omitempty"`
	Permalink         string        `json:"permalink,omitempty"`
	DateCreated       string        `json:"date_created,omitempty"`
	DateCreatedGmt    string        `json:"date_created_gmt,omitempty"`
	DateModified      string        `json:"date_modified,omitempty"`
	DateModifiedGmt   string        `json:"date_modified_gmt,omitempty"`
	Type              string        `json:"type,omitempty"`
	Status            string        `json:"status,omitempty"`
	Featured          bool          `json:"featured,omitempty"`
	CatalogVisibility string        `json:"catalog_visibility,omitempty"`
	Description       string        `json:"description,omitempty"`
	ShortDescription  string        `json:"short_description,omitempty"`
	Sku               string        `json:"sku,omitempty"`
	Price             string        `json:"price,omitempty"`
	RegularPrice      string        `json:"regular_price,omitempty"`
	SalePrice         string        `json:"sale_price,omitempty"`
	DateOnSaleFrom    string        `json:"date_on_sale_from,omitempty"`
	DateOnSaleFromGmt string        `json:"date_on_sale_from_gmt,omitempty"`
	DateOnSaleTo      string        `json:"date_on_sale_to,omitempty"`
	DateOnSaleToGmt   string        `json:"date_on_sale_to_gmt,omitempty"`
	PriceHtml         string        `json:"price_html,omitempty"`
	OnSale            bool          `json:"on_sale,omitempty"`
	Purchasable       bool          `json:"purchasable,omitempty"`
	TotalSales        string           `json:"total_sales,omitempty"`
	Virtual           bool          `json:"virtual,omitempty"`
	Downloadable      bool          `json:"downloadable,omitempty"`
	Downloads         []Download    `json:"downloads,omitempty"`
	DownloadLimit     int           `json:"download_limit,omitempty"`
	DownloadExpiry    int           `json:"download_expiry,omitempty"`
	ExternalUrl       string        `json:"external_url,omitempty"`
	ButtonText        string        `json:"button_text,omitempty"`
	TaxStatus         string        `json:"tax_status,omitempty"`
	TaxClass          string        `json:"tax_class,omitempty"`
	ManageStock       bool          `json:"manage_stock,omitempty"`
	StockQuantity     int           `json:"stock_quantity,omitempty"`
	StockStatus       string        `json:"stock_status,omitempty"`
	Backorders        string        `json:"backorders,omitempty"`
	BackordersAllowed bool          `json:"backorders_allowed,omitempty"`
	Backordered       bool          `json:"backordered,omitempty"`
	SoldIndividually  bool          `json:"sold_individually,omitempty"`
	Weight            string        `json:"weight,omitempty"`
	Dimensions        *Dimensions   `json:"dimensions,omitempty"`
	ShippingRequired  bool          `json:"shipping_required,omitempty"`
	ShippingTaxable   bool          `json:"shipping_taxable,omitempty"`
	ShippingClass     string        `json:"shipping_class,omitempty"`
	ShippingClassId   int64         `json:"shipping_class_id,omitempty"`
	ReviewsAllowed    bool          `json:"reviews_allowed,omitempty"`
	AverageRating     string        `json:"average_rating,omitempty"`
	RatingCount       int           `json:"rating_count,omitempty"`
	RelatedIds        []int64       `json:"related_ids,omitempty"`
	UpsellIds         []int64       `json:"upsell_ids,omitempty"`
	CrossSellIds      []int64       `json:"cross_sell_ids,omitempty"`
	ParentId          int64         `json:"parent_id,omitempty"`
	PurchaseNote      string        `json:"purchase_note,omitempty"`
	Categories        []Category    `json:"categories,omitempty"`
	Tags              []Tag         `json:"tags,omitempty"`
	Images            []Image       `json:"images,omitempty"`
	Attributes        []Attribute   `json:"attributes,omitempty"`
	DefaultAttributes []DefaultAttr `json:"default_attributes,omitempty"`
	Variations        []int64       `json:"variations,omitempty"`
	GroupedProducts   []int64       `json:"grouped_products,omitempty"`
	MenuOrder         int           `json:"menu_order,omitempty"`
	MetaData          []MetaData    `json:"meta_data,omitempty"`
	Links             Links         `json:"_links,omitempty"`
}

type Dimensions struct {
	Length string `json:"length,omitempty"`
	Width  string `json:"width,omitempty"`
	Height string `json:"height,omitempty"`
}

type Download struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	File string `json:"file,omitempty"`
}

type Category struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type Tag struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type Image struct {
	Id              int64  `json:"id,omitempty"`
	DateCreated     string `json:"date_created,omitempty"`
	DateCreatedGmt  string `json:"date_created_gmt,omitempty"`
	DateModified    string `json:"date_modified,omitempty"`
	DateModifiedGmt string `json:"date_modified_gmt,omitempty"`
	Src             string `json:"src,omitempty"`
	Name            string `json:"name,omitempty"`
	Alt             string `json:"alt,omitempty"`
	Position        int    `json:"position,omitempty"`
}

type Attribute struct {
	Id        int64    `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Position  int      `json:"position,omitempty"`
	Visible   bool     `json:"visible,omitempty"`
	Variation bool     `json:"variation,omitempty"`
	Options   []string `json:"options,omitempty"`
}

type DefaultAttr struct {
	Id     int64  `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Option string `json:"option,omitempty"`
}

func (p *ProductServiceOp) List(options interface{}) ([]Product, error) {
	products, _, err := p.ListWithPagination(options)
	return products, err
}

// ListWithPagination lists products and returns pagination to retrieve next/previous results.
func (p *ProductServiceOp) ListWithPagination(options interface{}) ([]Product, *Pagination, error) {
	path := fmt.Sprintf("%s", productsBasePath)
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

func (p *ProductServiceOp) Create(product Product) (*Product, error) {
	path := fmt.Sprintf("%s", productsBasePath)
	resource := new(Product)
	err := p.client.Post(path, product, &resource)
	return resource, err
}

// Get individual product
func (p *ProductServiceOp) Get(productID int64, options interface{}) (*Product, error) {
	path := fmt.Sprintf("%s/%d", productsBasePath, productID)
	resource := new(Product)
	err := p.client.Get(path, resource, options)
	return resource, err
}

// Update existing product
func (p *ProductServiceOp) Update(product *Product) (*Product, error) {
	path := fmt.Sprintf("%s/%d", productsBasePath, product.ID)
	resource := new(Product)
	err := p.client.Put(path, product, &resource)
	return resource, err
}

// Delete existing product
func (p *ProductServiceOp) Delete(productID int64, options interface{}) (*Product, error) {
	path := fmt.Sprintf("%s/%d", productsBasePath, productID)
	resource := new(Product)
	err := p.client.Delete(path, options, &resource)
	return resource, err
}
// Batch implements ProductService.
func (p *ProductServiceOp) Batch(data ProductBatchOption) (*ProductBatchResource, error) {
	path := fmt.Sprintf("%s/batch", productsBasePath)
	resource := new(ProductBatchResource)
	err := p.client.Post(path, data, &resource)
	return resource, err
}