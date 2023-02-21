package grpc_descriptor

import (
	"github.com/eolinker/eosc"
	"github.com/fullstorydev/grpcurl"
)

const (
	ServiceSkill = "github.com/eolinker/apinto/grpc-transcode.transcode.IDescriptor"
)

type IDescriptor interface {
	eosc.IWorker
	Descriptor() grpcurl.DescriptorSource
}

// CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == ServiceSkill
}
