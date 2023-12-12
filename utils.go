package generator

// Select returns `whenTrue` if `expr` is true, otherwise `whenFalse`.
func Select[T any](expr bool, whenTrue, whenFalse T) T {
	if expr {
		return whenTrue
	}
	return whenFalse
}

// Ref returns a pointer to `what`.
func Ref[T any](what T) *T {
	return &what
}

// DerefDef returns the dereferenced value of `what` or `defaultValue` if `what` is nil.
func DerefDef[T any](what *T, defaultValue T) T {
	if what != nil {
		return *what
	}
	return defaultValue
}

// Deref returns the dereferenced value of `what` or the zero value if `what` is nil.
func Deref[T any](what *T) T {
	if what != nil {
		return *what
	}
	var def T
	return def
}
