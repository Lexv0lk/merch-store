package domain

//region InvalidArgumentsError

type InvalidArgumentsError struct {
	Msg string
}

func (e *InvalidArgumentsError) Error() string {
	return e.Msg
}

func (e *InvalidArgumentsError) Is(target error) bool {
	_, ok := target.(*InvalidArgumentsError)
	return ok
}

//endregion

//region InsufficientBalanceError

type InsufficientBalanceError struct {
	Msg string
}

func (e *InsufficientBalanceError) Error() string {
	return e.Msg
}

func (e *InsufficientBalanceError) Is(target error) bool {
	_, ok := target.(*InsufficientBalanceError)
	return ok
}

//endregion

//region UserNotFoundError

type UserNotFoundError struct {
	Msg string
}

func (e *UserNotFoundError) Error() string {
	return e.Msg
}

func (e *UserNotFoundError) Is(target error) bool {
	_, ok := target.(*UserNotFoundError)
	return ok
}

//endregion

//region GoodNotFoundError

type GoodNotFoundError struct {
	Msg string
}

func (e *GoodNotFoundError) Error() string {
	return e.Msg
}

func (e *GoodNotFoundError) Is(target error) bool {
	_, ok := target.(*GoodNotFoundError)
	return ok
}

//endregion

//region BalanceNotFoundError

type BalanceNotFoundError struct {
	Msg string
}

func (e *BalanceNotFoundError) Error() string {
	return e.Msg
}

func (e *BalanceNotFoundError) Is(target error) bool {
	_, ok := target.(*BalanceNotFoundError)
	return ok
}

//endregion
