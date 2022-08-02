// This file contains the test for valid json tags.

package a

type Key struct {
	A, B string
}

type StrAlias string

type ChanAlias chan<- int

type Textable [2]byte

func (t Textable) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%02x%02x", t[0], t[1])), nil
}

func (t *Textable) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(string(text), "%02x%02x", &t[0], &t[1])
	return err
}

type ValidJsonTest struct {
	A int                           `json:"A"`
	B int                           `json:"B,omitempty"`
	C int                           `json:",omitempty"`
	D int                           `json:"-"`
	E chan<- int                    `json:"E"` // want "struct field has json tag but non-serializable type chan<- int"
	F chan<- int                    `json:"-"`
	G chan<- int                    `json:"-,"` // want "struct field has json tag but non-serializable type chan<- int"
	H ChanAlias                     `json:"H"`  // want "struct field has json tag but non-serializable type a.ChanAlias"
	I map[int]int                   `json:"I"`
	J map[int64]int                 `json:"J"`
	K map[Key]int                   `json:"K"` // want "struct field has json tag but non-serializable type"
	L map[struct{ A, B string }]int `json:"L"` // want "struct field has json tag but non-serializable type"
	M map[string]int                `json:"M"`
	N map[StrAlias]int              `json:"N"`
	O func()                        `json:"O"` // want "struct field has json tag but non-serializable type func()"
	P complex64                     `json:"P"` // want "struct field has json tag but non-serializable type complex64"
	Q complex128                    `json:"Q"` // want "struct field has json tag but non-serializable type complex128"
	R StrAlias                      `json:"R"`
	S map[Textable]int              `json:"S"`
}
