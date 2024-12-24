package server

const (
	mActiveConnections       = "connections.active"
	mConnectionDuration      = "connections.duration"
	mTargetDialDuration      = "connections.target_dial_duration"
	mErrorsAccept            = "errors.accept"
	mErrorsUnauthorizedIP    = "errors.unauthorized_ip"
	mErrorsGreeting          = "errors.greeting"
	mErrorsVersion           = "errors.unsupported_version"
	mErrorsAuthMethods       = "errors.auth_methods"
	mErrorsRequestHeader     = "errors.request_header"
	mErrorsCommand           = "errors.unsupported_command"
	mErrorsAddressType       = "errors.unsupported_address_type"
	mErrorsPort              = "errors.port"
	mErrorsDNSResolve        = "errors.dns_resolve"
	mErrorsConnectionRefused = "errors.connection_refused"
	mErrorsCopyToTarget      = "errors.copy_to_target"
	mErrorsCopyToClient      = "errors.copy_to_client"
	mBytesSent               = "traffic.bytes_sent"
	mBytesReceived           = "traffic.bytes_received"
)
