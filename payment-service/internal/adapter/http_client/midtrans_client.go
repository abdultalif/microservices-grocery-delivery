package httpclient

import (
	"payment-service/config"

	"github.com/labstack/gommon/log"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransClientInterface interface {
	CreateTransaction(amount int64, orderID, customerName, customerEmail string) (string, string, error)
}

type midtransClient struct {
	cfg *config.Config
}

// CreateTransaction implements MidtransClientInterface.
func (m *midtransClient) CreateTransaction(amount int64, orderID, customerName, customerEmail string) (string, string, error) {
	midtrans.ServerKey = m.cfg.Midtrans.ServerKey
	midtrans.Environment = midtrans.Sandbox

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: customerName,
			Email: customerEmail,
		},
	}

	snapRes, err := snap.CreateTransaction(snapReq)
	if err != nil {
		log.Errorf("[MidtransClient-1] Failed to create transaction: %v", err)
		return "", "", err
	}

	return snapRes.Token, snapRes.RedirectURL, nil
}

func NewMidtransClient(cfg *config.Config) MidtransClientInterface {
	return &midtransClient{cfg: cfg}
}
