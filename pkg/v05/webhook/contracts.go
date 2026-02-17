package webhook

type AuthScheme string

const (
	AuthSchemeNone   AuthScheme = "none"
	AuthSchemeHMAC   AuthScheme = "hmac"
	AuthSchemeBearer AuthScheme = "bearer"
	AuthSchemeMutual AuthScheme = "mutual"
)

type SecurityMode string

const (
	SecurityModePlain     SecurityMode = "plain"
	SecurityModeSigned    SecurityMode = "signed"
	SecurityModeEncrypted SecurityMode = "encrypted"
)

var DefaultSecurityMode = SecurityModePlain

var ModeGate = map[AuthScheme][]SecurityMode{
	AuthSchemeNone:   {SecurityModePlain},
	AuthSchemeHMAC:   {SecurityModeSigned, SecurityModeEncrypted},
	AuthSchemeBearer: {SecurityModeSigned, SecurityModeEncrypted},
	AuthSchemeMutual: {SecurityModeEncrypted},
}

type IdempotencyStrategy string

const (
	IdempotencyRequired IdempotencyStrategy = "required"
	IdempotencyOptional IdempotencyStrategy = "optional"
	IdempotencyDisabled IdempotencyStrategy = "disabled"
)

type RetryClass string

const (
	RetryClassImmediate RetryClass = "immediate"
	RetryClassDeferred  RetryClass = "deferred"
	RetryClassAbort     RetryClass = "abort"
)

var RetryClassMapping = map[int]RetryClass{
	408: RetryClassDeferred,
	429: RetryClassDeferred,
	500: RetryClassDeferred,
	502: RetryClassDeferred,
	503: RetryClassDeferred,
	504: RetryClassDeferred,
}

type EndpointPolicy struct {
	Auth          AuthScheme
	Idempotency   IdempotencyStrategy
	SecurityMode  SecurityMode
	RetryDecision RetryClass
}

func NewEndpointPolicy(auth AuthScheme, idempotency IdempotencyStrategy, mode SecurityMode) EndpointPolicy {
	return EndpointPolicy{
		Auth:          auth,
		Idempotency:   idempotency,
		SecurityMode:  mode,
		RetryDecision: RetryClassForStatus(200),
	}
}

func (p EndpointPolicy) AllowedSecurityModes() []SecurityMode {
	if modes, ok := ModeGate[p.Auth]; ok {
		return append([]SecurityMode(nil), modes...)
	}
	return []SecurityMode{DefaultSecurityMode}
}

func (p EndpointPolicy) IsPlaintextAllowed() bool {
	for _, candidate := range p.AllowedSecurityModes() {
		if candidate == SecurityModePlain {
			return true
		}
	}
	return false
}

func RetryClassForStatus(status int) RetryClass {
	if class, ok := RetryClassMapping[status]; ok {
		return class
	}
	switch {
	case status >= 500:
		return RetryClassDeferred
	case status >= 400 && status < 500:
		return RetryClassAbort
	default:
		return RetryClassImmediate
	}
}
