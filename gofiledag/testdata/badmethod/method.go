package badmethod

// Bar is a method of Foo, but lives in a different file from Foo.
func (f *Foo) Bar() {}
