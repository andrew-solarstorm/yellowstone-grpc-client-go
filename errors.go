package yellowstone

type GeyserGrpcClientError struct {
	Type    string
	Message string
	Err     error
}

func (e *GeyserGrpcClientError) Error() string {
	if e.Err != nil {
		return e.Type + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Type + ": " + e.Message
}

func (e *GeyserGrpcClientError) Unwrap() error {
	return e.Err
}

func NewGrpcStatusError(err error) *GeyserGrpcClientError {
	return &GeyserGrpcClientError{
		Type:    "GrpcStatus",
		Message: "gRPC status",
		Err:     err,
	}
}

func NewSubscribeSendError(err error) *GeyserGrpcClientError {
	return &GeyserGrpcClientError{
		Type:    "SubscribeSendError",
		Message: "Failed to send subscribe request",
		Err:     err,
	}
}

type GeyserGrpcBuilderError struct {
	Type    string
	Message string
	Err     error
}

func (e *GeyserGrpcBuilderError) Error() string {
	if e.Err != nil {
		return e.Type + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Type + ": " + e.Message
}

func (e *GeyserGrpcBuilderError) Unwrap() error {
	return e.Err
}

func NewMetadataValueError(err error) *GeyserGrpcBuilderError {
	return &GeyserGrpcBuilderError{
		Type:    "MetadataValueError",
		Message: "Failed to parse x-token",
		Err:     err,
	}
}

func NewTransportError(err error) *GeyserGrpcBuilderError {
	return &GeyserGrpcBuilderError{
		Type:    "TransportError",
		Message: "gRPC transport error",
		Err:     err,
	}
}

func NewInvalidUriError(uri string) *GeyserGrpcBuilderError {
	return &GeyserGrpcBuilderError{
		Type:    "InvalidUri",
		Message: "Invalid URI: " + uri,
	}
}
