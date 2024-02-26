package format

// NewField creates a new Field with the given name and value format string. See
// the Field struct for more information.
func NewField(name, valueFormat string) Field {
	return Field{Name: name, ValueFormat: valueFormat}
}

// NewDisplayer creates a new Displayer with the given payload, default format,
// and fields.
func NewDisplayer[T any](payload T, defaultFormat Format, fields []Field) Displayer {
	return &internalDisplayer[T]{
		payload:       payload,
		fields:        fields,
		defaultFormat: defaultFormat,
	}
}

type internalDisplayer[T any] struct {
	payload       T
	fields        []Field
	defaultFormat Format
}

func (i *internalDisplayer[T]) DefaultFormat() Format   { return i.defaultFormat }
func (i *internalDisplayer[T]) FieldTemplates() []Field { return i.fields }
func (i *internalDisplayer[T]) Payload() any            { return i.payload }
