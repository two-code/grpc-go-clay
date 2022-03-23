package transport

import (
	"github.com/utrack/clay/v3/transport/swagger"
)

type CompoundServiceDesc struct {
	svc []ServiceDesc
}

func NewCompoundServiceDesc(desc ...ServiceDesc) *CompoundServiceDesc {
	return &CompoundServiceDesc{svc: desc}
}

func (d *CompoundServiceDesc) RegisterHTTP(r Router) {
	for _, svc := range d.svc {
		svc.RegisterHTTP(r)
	}
}

func (d *CompoundServiceDesc) SwaggerDef(options ...swagger.Option) []byte {
	j := &swagJoiner{}
	for _, svc := range d.svc {
		j.AddDefinition(svc.SwaggerDef(options...))
	}
	return j.SumDefinitions()
}

func (d *CompoundServiceDesc) Apply(oo ...DescOption) {
	for _, ss := range d.svc {
		if s, ok := ss.(ConfigurableServiceDesc); ok {
			s.Apply(oo...)
		}
	}
}
