package gen_id

import gonanoid "github.com/matoous/go-nanoid/v2"

type IDGenerator interface {
	New() (string, error)
}

type NanoIDGenerator struct{}

func (i NanoIDGenerator) New() (string, error) {
	return gonanoid.New()
}

func CreateIDGenerator() IDGenerator {
	return NanoIDGenerator{}
}
