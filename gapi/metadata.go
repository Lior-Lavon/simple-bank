package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	userAgentHeader = "user-agent"
	xForwardedFor   = "x-forwarded-for"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if clientIPS := md.Get(xForwardedFor); len(clientIPS) > 0 {
			mtdt.ClientIP = clientIPS[0]
		}
	}

	return mtdt
}
