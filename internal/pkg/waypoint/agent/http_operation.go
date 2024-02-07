package agent

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-hclog"
)

type HTTPOperation struct {
	URL string
}

func (h *HTTPOperation) Run(ctx context.Context, log hclog.Logger) (OperationStatus, error) {
	resp, err := http.Get(h.URL)
	if err != nil {
		return errStatus, err
	}

	defer resp.Body.Close()

	return cleanStatus, nil
}
