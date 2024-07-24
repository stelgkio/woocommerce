package woocommerce

import (
	"testing"
)

func init() {
	app := App{
		CustomerKey:    customerKey,
		CustomerSecret: customerSecret,
	}

	client = NewClient(app, shopUrl,
		WithLog(&LeveledLogger{
			Level: LevelDebug, // you should open this for debug in dev environment,  usefully.
		}),
		WithRetry(3))
}


func TestReportServiceOp_GetTotalCustomers(t *testing.T) {
	report, err := client.Report.GetTotalCustomers(nil)
	
	t.Logf("report : %v, err: %v", report, err)
}

func TestReportServiceOp_GetTotalOrders(t *testing.T) {
	report, err := client.Report.GetTotalOrders(nil)
	
	t.Logf("report : %v, err: %v", report, err)
}

func TestReportServiceOp_GetTotalProducts(t *testing.T) {
	report, err := client.Report.GetTotalProducts(nil)
	
	t.Logf("report : %v, err: %v", report, err)
}