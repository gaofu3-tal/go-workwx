package httpapi

import (
	"context"
	"io"
	"net/http"

	"github.com/gaofu3-tal/go-workwx/internal/lowlevel/envelope"
)

type EnvelopeHandler interface {
	OnIncomingEnvelope(ctx context.Context, rx envelope.Envelope) error
}

func (h *LowlevelHandler) eventHandler(
	rw http.ResponseWriter,
	r *http.Request,
) {
	// request bodies are assumed small
	// we can't do streaming parse/decrypt/verification anyway
	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// signature verification is inside EnvelopeProcessor
	ev, err := h.ep.HandleIncomingMsg(r.URL, body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.eh.OnIncomingEnvelope(r.Context(), ev)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// currently we always return empty 200 responses
	// any reply is to be sent asynchronously
	// this might change in the future (maybe save a couple of RTT or so)
	rw.WriteHeader(http.StatusOK)
}
