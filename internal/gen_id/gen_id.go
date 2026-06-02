package gen_id

import gonanoid "github.com/matoous/go-nanoid/v2"

type IdGenerator interface {
	New() (string, error)
}

type NanoIdGenerator struct{}

func (i NanoIdGenerator) New() (string, error) {
	return gonanoid.New()
}

func CreateIdGenerator() IdGenerator {
	return NanoIdGenerator{}
}
