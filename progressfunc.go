package fynewidgets

// ProgressFunc contains a function, a channel for monitoring progress, and another for sending output.
// It is designed to be wired to a GUI element so that progress can be displayed and output sent to the
// right place.
//
// Functions send their progress to the channel directly, so the progress channel appears in the constructor
type ProgressFunc struct {
	Func     func(c chan float64, vars ...any) any
	progress chan float64
	output   chan any
	vars     []any
	reset    bool
}

// Constructor for ProgressFunc - takes a function that returns an any type, with a reporting channel and an output channel
// Arguments are variadic, of any type. They are just called by Execute()
//   - The function should update the progress channel with its current completion (0.0-1.0)
//   - the progressChannel is the specific channel to use to monitor progress - it may be created by a UI call
//   - the output is the channel in which final results are placed
//   - reset - progress is set to 0 at the end of the function if this is true
func NewProgressFunc(f func(cp chan float64, vars ...any) any, progressChannel chan float64, output chan any, reset bool, vars ...any) *ProgressFunc {
	p := &ProgressFunc{
		Func:     f,
		progress: progressChannel,
		vars:     vars,
		output:   output,
		reset:    reset,
	}
	return p
}

// runs the function and sends output to a channel.
//   - Execute is generally triggered by a UI event or similar.
//   - The progress channel is updated during the function, and reset to zero at the end
//   - The output channel is used to send the output to something that can use it, as the UI
//     is not aware of the specifics of functions.
func (p *ProgressFunc) Execute() {
	p.output <- p.Func(p.progress, p.vars...)
	if p.reset {
		p.progress <- 0.0
	}
}
