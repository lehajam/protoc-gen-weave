package plugin

const tests = `
{{ range $message := .Messages}}

func Test{{ $message.Name }}HandlerCheck(t *testing.T) {
	// _, signer := helpers.MakeKey()

	testHandlerCheck(
		t,
		[]testcase{
			// Add your testcases here
		},
	)
}

func Test{{ $message.Name }}HandlerDeliver(t *testing.T) {
	// _, signer := helpers.MakeKey()

	testHandlerDeliver(
		t,
		[]testcase{
			// Add your testcases here
		},
	)
}

{{ end }}`
