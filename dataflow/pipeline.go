package dataflow

import "fmt"

// In your dataflow package

// Stage is a processing step: input T → output T (or error).
type Stage[T any] func(T) (T, error)

// Pipeline is an ordered chain of stages for type T.
type Pipeline[T any] struct {
	stages []Stage[T]
}

func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{}
}

func (p *Pipeline[T]) Add(stage Stage[T]) {
	p.stages = append(p.stages, stage)
}

// Run executes all stages on input, returning final output or first error.
func (p *Pipeline[T]) Run(input T) (T, error) {
	current := input
	var err error
	for _, st := range p.stages {
		current, err = st(current)
		if err != nil {
			return current, err
		}
	}
	return current, nil
}

// Descriptor that UIs and configs work with.
type StageDescriptor struct {
	ID     string         // unique stage ID
	Name   string         // display name
	Params map[string]any // stage-specific config
	Kind   string         // e.g. "blur", "resize", "kalman"
}

// A flow is a linear sequence (you can later upgrade to a DAG).
type Flow struct {
	Stages []StageDescriptor
}

// Registry mapping Kind -> factory for Stage[T].
type StageFactory[T any] func(desc StageDescriptor) (Stage[T], error)

type Registry[T any] struct {
	factories map[string]StageFactory[T]
}

func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{factories: make(map[string]StageFactory[T])}
}

func (r *Registry[T]) Register(kind string, factory StageFactory[T]) {
	r.factories[kind] = factory
}

func (r *Registry[T]) BuildPipeline(flow Flow) (*Pipeline[T], error) {
	p := NewPipeline[T]()
	for _, desc := range flow.Stages {
		factory, ok := r.factories[desc.Kind]
		if !ok {
			return nil, fmt.Errorf("unknown stage kind %q", desc.Kind)
		}
		st, err := factory(desc)
		if err != nil {
			return nil, err
		}
		p.Add(st)
	}
	return p, nil
}
