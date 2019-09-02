package runtime

//HandleCrash catch panics and notify register function
func HandleCrash(panicHandlers ...func(interface{})) {
	if err := recover(); err != nil {
		for _, handler := range panicHandlers {
			handler(err)
		}
	}
}
