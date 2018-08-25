package api

import "fmt"

type TaxClient interface {
	Register(email string) error
	Login(smsPassword string) (*TaxLoginResponse, error)
	Restore() error

	BillCheck(fiscalNumber, fiscalDocument, fiscalDocumentAttr string, total float64) error
	BillDetail(fiscalNumber, fiscalDocument, fiscalDocumentAttr string) (*TaxBillCheckResponse, error)
}

type TaxRegisterRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type TaxRestoreRequest struct {
	Phone string `json:"phone"`
}

type TaxLoginResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type TaxBillCheckResponse struct {
	Document struct {
		Receipt struct {
			OperationType        int    `json:"operationType"`
			FiscalSign           int64  `json:"fiscalSign"`
			DateTime             string `json:"dateTime"`
			RawData              string `json:"rawData"`
			TotalSum             int    `json:"totalSum"`
			Nds10                int    `json:"nds10"`
			UserInn              string `json:"userInn"`
			TaxationType         int    `json:"taxationType"`
			Operator             string `json:"operator"`
			FiscalDocumentNumber int    `json:"fiscalDocumentNumber"`
			Properties           []struct {
				Value string `json:"value"`
				Key   string `json:"key"`
			} `json:"properties"`
			ReceiptCode       int    `json:"receiptCode"`
			RequestNumber     int    `json:"requestNumber"`
			User              string `json:"user"`
			KktRegID          string `json:"kktRegId"`
			FiscalDriveNumber string `json:"fiscalDriveNumber"`
			Items             []struct {
				Sum      int    `json:"sum"`
				Price    int    `json:"price"`
				Name     string `json:"name"`
				Quantity int    `json:"quantity"`
				Nds10    int    `json:"nds10"`
			} `json:"items"`
			EcashTotalSum      int    `json:"ecashTotalSum"`
			RetailPlaceAddress string `json:"retailPlaceAddress"`
			CashTotalSum       int    `json:"cashTotalSum"`
			ShiftNumber        int    `json:"shiftNumber"`
		} `json:"receipt"`
	} `json:"document"`
}

type ErrUnexpectedHTTPStatus struct {
	Status int
	Body   []byte
}

func (s ErrUnexpectedHTTPStatus) Error() string {
	return fmt.Sprintf("unexpected HTTP status: %d with body: %s", s.Status, string(s.Body))
}
