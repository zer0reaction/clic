package lexer

func (l *Lexer) DebugCacheToken() error {
	return l.cacheToken()
}

func (l *Lexer) DebugReadToken() *Token {
	return l.popToken()
}
