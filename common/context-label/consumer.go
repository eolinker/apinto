package context_label

import "github.com/eolinker/eosc/eocontext"

const (
	LabelConsumerType = "consumer_type"
)

type ConsumerType string

const (
	ConsumerTypeUser          = "user"
	ConsumerTypeResourceGroup = "resource_group"
)

func GetConsumerType(ctx eocontext.EoContext) ConsumerType {
	value := ctx.GetLabel(LabelConsumerType)
	return ConsumerType(value)
}

func SetConsumerType(ctx eocontext.EoContext, consumerType ConsumerType) {
	ctx.SetLabel(LabelConsumerType, string(consumerType))
}

func IsUserConsumer(ctx eocontext.EoContext) bool {
	return GetConsumerType(ctx) == ConsumerTypeUser
}

func IsResourceGroupConsumer(ctx eocontext.EoContext) bool {
	return GetConsumerType(ctx) == ConsumerTypeResourceGroup
}
