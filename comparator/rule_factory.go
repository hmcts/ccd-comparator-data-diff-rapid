package comparator

import (
	"ccd-comparator-data-diff-rapid/config"
	"github.com/pkg/errors"
	"strings"
)

type RuleFactory struct {
	configuration   *config.Configurations
	enabledRuleList []Rule
}

func NewRuleFactory(configuration *config.Configurations) *RuleFactory {
	factory := &RuleFactory{configuration: configuration}
	rules, err := factory.createEnabledRuleList(configuration.Active)
	if err != nil {
		panic(err)
	}
	factory.enabledRuleList = rules
	return factory
}

func (f RuleFactory) createEnabledRuleList(activeAnalyzeRules string) ([]Rule, error) {
	if activeAnalyzeRules == "" {
		return nil, errors.New("No active rule defined")
	}

	enabledRuleTypes := parseActiveAnalyzeRules(activeAnalyzeRules)

	var ruleConfig = f.configuration.Scan
	rules := make([]Rule, 0)

	if enabledRuleTypes[RuleTypeStaticFieldChange] {
		rules = append(rules, NewStaticFieldChangeRule(ruleConfig.Concurrent.Event.ThresholdMilliseconds,
			ruleConfig.Report.MaskValue))
	}
	if enabledRuleTypes[RuleTypeDynamicFieldChange] {
		rules = append(rules, NewDynamicFieldChangeRule(ruleConfig.Concurrent.Event.ThresholdMilliseconds,
			ruleConfig.Report.MaskValue, ruleConfig.Filter))
	}
	if enabledRuleTypes[RuleTypeFieldChangeCount] {
		rules = append(rules, NewFieldChangeCountRule(ruleConfig.FieldChange.Threshold))
	}

	return rules, nil
}

func parseActiveAnalyzeRules(activeAnalyzeRules string) map[RuleType]bool {
	ruleMap := make(map[RuleType]bool)

	ruleNames := strings.Split(activeAnalyzeRules, ",")
	for _, name := range ruleNames {
		name = strings.TrimSpace(name)
		if ruleType, ok := ruleTypeFromString(name); ok {
			ruleMap[ruleType] = true
		}
	}

	return ruleMap
}

func ruleTypeFromString(name string) (RuleType, bool) {
	switch name {
	case "staticfieldchange":
		return RuleTypeStaticFieldChange, true
	case "dynamicfieldchange":
		return RuleTypeDynamicFieldChange, true
	case "fieldchangecount":
		return RuleTypeFieldChangeCount, true
	default:
		return RuleTypeUnknown, false
	}
}

func (f RuleFactory) GetEnabledRuleList() []Rule {
	return f.enabledRuleList
}

type RuleType int

const (
	RuleTypeUnknown RuleType = iota
	RuleTypeStaticFieldChange
	RuleTypeFieldChangeCount
	RuleTypeDynamicFieldChange
)
