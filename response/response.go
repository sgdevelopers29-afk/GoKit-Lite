package response

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(data any) Response {
	return Response{
		Success: true,
		Message: "success",
		Data:    data,
	}
}

func Error(message string) Response {
	return Response{
		Success: false,
		Message: message,
		Data:    nil,
	}
}
