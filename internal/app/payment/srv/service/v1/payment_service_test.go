package v1

import (
    "testing"
    "strings"
    "emshop/internal/app/pkg/options"
    "github.com/stretchr/testify/assert"
)

func TestPayment_generatePaymentSn(t *testing.T) {
    ps := &paymentService{}
    sn := ps.generatePaymentSn()
    assert.True(t, strings.HasPrefix(sn, "PAY"))
    assert.Greater(t, len(sn), 8)
}

func TestDTMManager_ProcessNoError(t *testing.T) {
    dm := NewDTMManager(&options.DtmOptions{GrpcServer: "dtm-grpc"})
    err := dm.ProcessOrderSubmission(nil, &OrderSubmissionRequest{OrderSn: "OSN-1", Amount: 10})
    assert.NoError(t, err)
    err = dm.ProcessPaymentSuccess(nil, &PaymentSuccessRequest{PaymentSn: "PSN-1", OrderSn: "OSN-1", ReceiverName: "A", ReceiverAddress: "B"})
    assert.NoError(t, err)
}

