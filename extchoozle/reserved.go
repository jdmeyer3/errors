package extchoozle

type cErr struct {
	defaultMessage string
	httpCode       int
}

type ChoozleErrorCode int

const (
	ErrorProvider = 1104
)

var choozleErrors = map[ChoozleErrorCode]cErr{
	ErrorProvider: {
		defaultMessage: "Provider Error",
		httpCode:       400,
	},
}
