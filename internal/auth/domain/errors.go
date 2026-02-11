package domain

//region CredentialsMismatchError

type CredentialsMismatchError struct {
	Msg string
}

func (e *CredentialsMismatchError) Error() string {
	return e.Msg
}

func (e *CredentialsMismatchError) Is(target error) bool {
	_, ok := target.(*CredentialsMismatchError)
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
