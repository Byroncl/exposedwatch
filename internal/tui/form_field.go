package tui

// FormField is one row of the Telegram settings form.
type FormField int

// Rows of the form, in display order.
const (
	FieldEnabled FormField = iota
	FieldToken
	FieldChatID
	FieldMinSeverity
	FieldCooldown
	FieldInterval
	FieldSave
	FieldTest
	formFieldCount
)

// fieldKind tells how a row reacts to enter.
type fieldKind int

const (
	kindToggle fieldKind = iota
	kindSecret
	kindText
	kindCycle
	kindNumber
	kindAction
)

// fieldSpec describes one row.
type fieldSpec struct {
	label string
	kind  fieldKind
}

var fieldSpecs = [formFieldCount]fieldSpec{
	FieldEnabled:     {fieldLabelEnabled, kindToggle},
	FieldToken:       {fieldLabelToken, kindSecret},
	FieldChatID:      {fieldLabelChatID, kindText},
	FieldMinSeverity: {fieldLabelMinSeverity, kindCycle},
	FieldCooldown:    {fieldLabelCooldown, kindNumber},
	FieldInterval:    {fieldLabelInterval, kindNumber},
	FieldSave:        {fieldLabelSave, kindAction},
	FieldTest:        {fieldLabelTest, kindAction},
}

// label is the text of the row.
func (field FormField) label() string {
	return fieldSpecs[field].label
}

func (field FormField) kind() fieldKind {
	return fieldSpecs[field].kind
}

// isEditable tells whether enter starts typing into the row.
func (field FormField) isEditable() bool {
	kind := field.kind()
	return kind == kindSecret || kind == kindText || kind == kindNumber
}

// isAction tells whether the row is a button.
func (field FormField) isAction() bool {
	return field.kind() == kindAction
}

// String makes fields readable in messages.
func (field FormField) String() string {
	return field.label()
}

// next and previous clamp at the ends: a form is not a ring.
func (field FormField) next() FormField {
	return min(field+1, formFieldCount-1)
}

func (field FormField) previous() FormField {
	return max(field-1, 0)
}
