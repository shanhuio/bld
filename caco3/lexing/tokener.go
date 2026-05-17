package lexing

// Tokener is token emitting interface.
type Tokener interface {
	// Token returns the next token
	Token() *Token

	// Errs returns the error list on tokening
	Errs() []*Error
}

// Tokens takes a lexer that is already setup and returns
// its tokens and errors.
func Tokens(tokener Tokener) ([]*Token, []*Error) {
	var ret []*Token
	for {
		t := tokener.Token()
		ret = append(ret, t)
		if t.Type == EOF {
			break
		}
	}
	return ret, tokener.Errs()
}
