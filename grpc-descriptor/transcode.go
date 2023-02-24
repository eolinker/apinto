package grpc_descriptor

import (
	"github.com/fullstorydev/grpcurl"
)

const (
	Skill = "github.com/eolinker/apinto/grpc-transcode.transcode.IDescriptor"
)

type IDescriptor interface {
	Descriptor() grpcurl.DescriptorSource
}

// CheckSkill 检查目标能力是否符合
func CheckSkill(skill string) bool {
	return skill == Skill
}
