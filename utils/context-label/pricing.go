package context_label

import (
	"fmt"
	"github.com/eolinker/eosc/eocontext"
)

const (
	LabelPriceCost           = "price_cost"
	LabelPriceSale           = "price_sale"
	LabelPriceOfficial       = "price_official"
	LabelAmountCost          = "cost_amount"
	LabelAmountSale          = "sale_amount"
	LabelAmountOfficial      = "official_amount"
	LabelChargeRule          = "charge_rule" // 计费规则ID
	LabelPriceVariable       = "price_variable"
	LabelPriceVariables      = "price_variables"
	LabelExprCost            = "cost_expr"
	LabelExprSale            = "sale_expr"
	LabelExprOfficial        = "official_expr"
	LabelPriceMatchRule      = "price_variables"
	LabelPriceMatchCondition = "price_conditions"
)

func SetPriceMatchCondition(ctx eocontext.EoContext, value interface{}) {
	ctx.WithValue(LabelPriceMatchCondition, value)
}

func GetPriceMatchCondition(ctx eocontext.EoContext) interface{} {
	return ctx.Value(LabelPriceMatchCondition)
}

func SetPriceMatchRule(ctx eocontext.EoContext, value interface{}) {
	ctx.WithValue(LabelPriceMatchRule, value)
}

func GetPriceMatchRule(ctx eocontext.EoContext) interface{} {
	return ctx.Value(LabelPriceMatchRule)
}

func SetExprCost(ctx eocontext.EoContext, value string) {
	ctx.SetLabel(LabelExprCost, value)
}

func GetExprCost(ctx eocontext.EoContext) string {
	return ctx.GetLabel(LabelExprCost)
}

func SetExprSale(ctx eocontext.EoContext, value string) {
	ctx.SetLabel(LabelExprSale, value)
}

func GetExprSale(ctx eocontext.EoContext) string {
	return ctx.GetLabel(LabelExprSale)
}

func SetExprOfficial(ctx eocontext.EoContext, value string) {
	ctx.SetLabel(LabelExprOfficial, value)
}

func GetExprOfficial(ctx eocontext.EoContext) string {
	return ctx.GetLabel(LabelExprOfficial)
}

func SetPriceVariable(ctx eocontext.EoContext, key string, value interface{}) {
	ctx.WithValue(fmt.Sprintf("%s.%s", LabelPriceVariable, key), value)
}

func GetPriceVariable(ctx eocontext.EoContext, key string) interface{} {
	return ctx.Value(fmt.Sprintf("%s.%s", LabelPriceVariable, key))
}

func SetPriceVariables(ctx eocontext.EoContext, value map[string]interface{}) {
	ctx.WithValue(LabelPriceVariables, value)
}

func GetPriceVariables(ctx eocontext.EoContext) map[string]interface{} {
	return ctx.Value(LabelPriceVariables).(map[string]interface{})
}

func GetChargeRule(ctx eocontext.EoContext) string {
	return ctx.GetLabel(LabelChargeRule)
}

func SetChargeRule(ctx eocontext.EoContext, value string) {
	ctx.SetLabel(LabelChargeRule, value)
}

func SetPrice(ctx eocontext.EoContext, label string, value string) {
	ctx.SetLabel(label, value)
}

func GetPrice(ctx eocontext.EoContext, label string) string {
	return ctx.GetLabel(label)
}

func SetPriceCost(ctx eocontext.EoContext, value string) {
	SetPrice(ctx, LabelPriceCost, value)
}

func GetPriceCost(ctx eocontext.EoContext) string {
	return GetPrice(ctx, LabelPriceCost)
}

func SetPriceSale(ctx eocontext.EoContext, value string) {
	SetPrice(ctx, LabelPriceSale, value)
}

func GetPriceSale(ctx eocontext.EoContext) string {
	return GetPrice(ctx, LabelPriceSale)
}

func SetPriceOfficial(ctx eocontext.EoContext, value string) {
	SetPrice(ctx, LabelPriceOfficial, value)
}

func GetPriceOfficial(ctx eocontext.EoContext) string {
	return GetPrice(ctx, LabelPriceOfficial)
}

func SetAmount(ctx eocontext.EoContext, label string, value string) {
	ctx.SetLabel(label, value)
}

func GetAmount(ctx eocontext.EoContext, label string) string {
	return ctx.GetLabel(label)
}

func SetAmountCost(ctx eocontext.EoContext, value string) {
	SetAmount(ctx, LabelAmountCost, value)
}

func GetAmountCost(ctx eocontext.EoContext) string {
	return GetAmount(ctx, LabelAmountCost)
}

func SetAmountSale(ctx eocontext.EoContext, value string) {
	SetAmount(ctx, LabelAmountSale, value)
}

func GetAmountSale(ctx eocontext.EoContext) string {
	return GetAmount(ctx, LabelAmountSale)
}

func SetAmountOfficial(ctx eocontext.EoContext, value string) {
	SetAmount(ctx, LabelAmountOfficial, value)
}

func GetAmountOfficial(ctx eocontext.EoContext) string {
	return GetAmount(ctx, LabelAmountOfficial)
}
