package log

type ServerError interface {
	Status() int
	Desc() *string
	Error() error
	File() *string
	Line() int
}

type serverError struct {
	status int
	desc   *string
	err    error
	file   *string
	line   int
}

func New(status int, desc string, err error) ServerError {
	file, line := getCaller()
	return &serverError{
		status: status,
		desc:   &desc,
		err:    err,
		file:   &file,
		line:   line,
	}
}

func NewError(status int, desc string) ServerError {
	file, line := getCaller()
	return &serverError{
		status: status,
		desc:   &desc,
		file:   &file,
		line:   line,
	}
}

func NewInternalError(err error) ServerError {
	file, line := getCaller()
	return &serverError{
		err:  err,
		file: &file,
		line: line,
	}
}

func (e *serverError) Status() int {
	return e.status
}

func (e *serverError) Desc() *string {
	return e.desc
}

func (e *serverError) Error() error {
	return e.err
}

func (e *serverError) File() *string {
	return e.file
}

func (e *serverError) Line() int {
	return e.line
}
