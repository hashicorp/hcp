// package operation is used to wait for an HCP Operation to complete.
package operation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/stable/2020-05-05/client/operation_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// Waiter is used to wait for an Operation to complete.
type Waiter interface {
	// Wait waits for the operation to complete or for the context to be
	// cancelled.
	Wait(context.Context, *models.HashicorpCloudOperationOperation) (*models.HashicorpCloudOperationOperation, error)
}

type waiter struct {
	c operation_service.ClientService
	l hclog.Logger
}

func New(c operation_service.ClientService, logger hclog.Logger) Waiter {
	return &waiter{
		c: c,
		l: logger.Named("operation_waiter"),
	}
}

func (w *waiter) Wait(ctx context.Context, op *models.HashicorpCloudOperationOperation) (*models.HashicorpCloudOperationOperation, error) {
	resultCh := make(chan result, 1)
	defer close(resultCh)
	go w.wait(ctx, op, resultCh)

	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	case res := <-resultCh:
		return res.operation, res.err
	}
}

type result struct {
	operation *models.HashicorpCloudOperationOperation
	err       error
}

func (w *waiter) wait(ctx context.Context, op *models.HashicorpCloudOperationOperation, resultCh chan<- result) {
	waitReq := operation_service.NewWaitParams()
	waitReq.ID = op.ID
	waitReq.LocationOrganizationID = op.Location.OrganizationID
	waitReq.LocationProjectID = op.Location.ProjectID
	w.l.Debug("waiting on operation", "id", op.ID)

	for {
		// Make the request
		resp, err := w.c.Wait(waitReq, nil)

		// Check for the happy path
		if err == nil {
			// Ensure we have a valid response
			if err := validateResponse(resp); err != nil {
				w.l.Debug("invalid Wait response", "error", err)
				resultCh <- result{err: err}
				return
			}

			o := resp.Payload.Operation
			if *o.State == models.HashicorpCloudOperationOperationStateDONE {
				resultCh <- result{operation: o}
				return
			}

			// Making another wait call
			continue
		}

		// Check if the error is retriable
		var waitErr *operation_service.WaitDefault
		if errors.As(err, &waitErr) {
			switch waitErr.Code() {
			case http.StatusInternalServerError:
			case http.StatusTooManyRequests:
				// Sleep for a few seconds
				time.Sleep(3 * time.Second)
			case http.StatusConflict:
			case http.StatusServiceUnavailable:
			default:
				// Non-retrieable error
				w.l.Debug("received non-retriable error", "error", err, "code", waitErr.Code())
				resultCh <- result{err: waitErr}
				return
			}

			w.l.Debug("retrying wait", "error", err, "code", waitErr.Code())
			continue
		} else {
			// Unknown error
			w.l.Debug("received unknown error", "error", err)
			resultCh <- result{err: waitErr}
			return
		}
	}
}

// validateResponse validates we received a valid server response.
func validateResponse(resp *operation_service.WaitOK) error {
	if resp == nil {
		return fmt.Errorf("received nil response")
	} else if resp.Payload == nil {
		return fmt.Errorf("received nil payload")
	} else if resp.Payload.Operation == nil {
		return fmt.Errorf("received nil operation in response")
	} else if resp.Payload.Operation.State == nil {
		return fmt.Errorf("received nil operation state")
	}

	return nil
}
