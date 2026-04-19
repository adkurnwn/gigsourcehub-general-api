package helpers

import (
	"context"
)

type contextKey string

const (
	ActorIDKey   contextKey = "actor_id"
	IPAddressKey contextKey = "ip_address"
	UserAgentKey contextKey = "user_agent"
	EndpointKey  contextKey = "endpoint"
)

func SetActorID(ctx context.Context, actorID string) context.Context {
	return context.WithValue(ctx, ActorIDKey, actorID)
}

func GetActorID(ctx context.Context) string {
	id, _ := ctx.Value(ActorIDKey).(string)
	return id
}

func SetIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, IPAddressKey, ip)
}

func GetIPAddress(ctx context.Context) string {
	ip, _ := ctx.Value(IPAddressKey).(string)
	return ip
}

func SetUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, UserAgentKey, ua)
}

func GetUserAgent(ctx context.Context) string {
	ua, _ := ctx.Value(UserAgentKey).(string)
	return ua
}

func SetEndpoint(ctx context.Context, ep string) context.Context {
	return context.WithValue(ctx, EndpointKey, ep)
}

func GetEndpoint(ctx context.Context) string {
	ep, _ := ctx.Value(EndpointKey).(string)
	return ep
}
