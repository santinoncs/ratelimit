package ratelimit

import (
	pb_legacy "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v2"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/golang/protobuf/jsonpb"
	"github.com/lyft/gostats"
	"golang.org/x/net/context"
)

type RateLimitLegacyServiceServer interface {
	pb_legacy.RateLimitServiceServer
}

// legacyService is used to implement v2 rls.proto (https://github.com/envoyproxy/data-plane-api/blob/master/envoy/service/ratelimit/v2/rls.proto)
// the legacyService receives RateLimitRequests, converts the request, and calls the service's ShouldRateLimit method.
type legacyService struct {
	s                          *service
	shouldRateLimitLegacyStats shouldRateLimitLegacyStats
}

type shouldRateLimitLegacyStats struct {
	reqConversionError   stats.Counter
	respConversionError  stats.Counter
	shouldRateLimitError stats.Counter
}

func newShouldRateLimitLegacyStats(scope stats.Scope) shouldRateLimitLegacyStats {
	s := scope.Scope("call.should_rate_limit_legacy")
	return shouldRateLimitLegacyStats{
		reqConversionError:   s.NewCounter("req_conversion_error"),
		respConversionError:  s.NewCounter("resp_conversion_error"),
		shouldRateLimitError: s.NewCounter("should_rate_limit_error"),
	}
}

func (this *legacyService) ShouldRateLimit(
	ctx context.Context,
	legacyRequest *pb_legacy.RateLimitRequest) (finalResponse *pb_legacy.RateLimitResponse, finalError error) {

	request, err := ConvertLegacyRequest(legacyRequest)
	if err != nil {
		this.shouldRateLimitLegacyStats.reqConversionError.Inc()
		return nil, err
	}
	resp, err := this.s.ShouldRateLimit(ctx, request)
	if err != nil {
		this.shouldRateLimitLegacyStats.shouldRateLimitError.Inc()
		return nil, err
	}

	legacyResponse, err := ConvertResponse(resp)
	if err != nil {
		this.shouldRateLimitLegacyStats.respConversionError.Inc()
		return nil, err
	}

	return legacyResponse, nil
}

func ConvertLegacyRequest(legacyRequest *pb_legacy.RateLimitRequest) (*pb.RateLimitRequest, error) {
	if legacyRequest == nil {
		return nil, nil
	}

	m := &jsonpb.Marshaler{}
	s, err := m.MarshalToString(legacyRequest)
	if err != nil {
		return nil, err
	}

	req := &pb.RateLimitRequest{}
	err = jsonpb.UnmarshalString(s, req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func ConvertResponse(response *pb.RateLimitResponse) (*pb_legacy.RateLimitResponse, error) {
	if response == nil {
		return nil, nil
	}

	m := &jsonpb.Marshaler{}
	s, err := m.MarshalToString(response)
	if err != nil {
		return nil, err
	}

	resp := &pb_legacy.RateLimitResponse{}
	err = jsonpb.UnmarshalString(s, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
